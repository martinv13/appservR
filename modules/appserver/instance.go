package appserver

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/martinv13/go-shiny/modules/config"
	"github.com/martinv13/go-shiny/modules/portspool"
	uuid "github.com/satori/go.uuid"
)

type Instance struct {
	sync.RWMutex
	ID           string
	appName      string
	appDir       string
	status       string
	port         string
	stdErr       string
	cmd          *exec.Cmd
	userCount    int
	restartDelay int
	config       config.Config
}

var instStatus = struct {
	STARTING    string
	RUNNING     string
	PHASING_OUT string
	ERROR       string
	STOPPED     string
}{
	STARTING:    "STARTING",
	RUNNING:     "RUNNING",
	PHASING_OUT: "PHASING_OUT",
	ERROR:       "ERROR",
	STOPPED:     "STOPPED",
}

// Create a new instance of the app
func NewInstance(appName string, appDir string, conf config.Config) *Instance {
	inst := &Instance{
		ID:      uuid.NewV4().String(),
		appName: appName,
		appDir:  appDir,
		config:  conf,
		status:  instStatus.STOPPED,
	}
	return inst
}

// Get instance status
func (inst *Instance) Status() string {
	inst.RLock()
	defer inst.RUnlock()
	return inst.status
}

// Get instance port
func (inst *Instance) Port() string {
	inst.RLock()
	defer inst.RUnlock()
	return inst.port
}

// Get instance status
func (inst *Instance) UserCount() int {
	inst.RLock()
	defer inst.RUnlock()
	return inst.userCount
}

// Get instance status
func (inst *Instance) StdErr() string {
	inst.RLock()
	defer inst.RUnlock()
	return inst.stdErr
}

// Start an instance of the app and relaunch when it fails
func (inst *Instance) Start() error {
	inst.Lock()
	defer inst.Unlock()
	port, err := portspool.GetNext()
	if err != nil {
		return err
	}
	inst.port = port
	_, err = os.Stat(inst.appDir)
	if err != nil {
		inst.status = instStatus.ERROR
		inst.stdErr = "App source directory does not exist"
		return errors.New("App source directory does not exist")
	}
	inst.status = instStatus.STARTING
	cmd := exec.Command(inst.config.GetString("Rscript"), "-e", "shiny::runApp('.', port="+inst.port+")")
	cmd.Dir = inst.appDir
	cmd.Env = os.Environ()
	inst.cmd = cmd
	stdErr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	// Goroutine to read output and save stderr
	scanner := bufio.NewScanner(stdErr)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			inst.Lock()
			inst.stdErr += line + "\n"
			if strings.HasPrefix(line, "Listening on") {
				inst.status = instStatus.RUNNING
				fmt.Println("app " + inst.appName + " at " + inst.port + " is running")
				inst.Unlock()
				return
			}
			inst.Unlock()
		}
	}()

	// Actually starting the subprocess
	if err = cmd.Start(); err != nil {
		return err
	}

	// Goroutine to restart the instance on stop
	go func() {
		err := cmd.Wait()
		inst.Lock()
		defer inst.Unlock()
		restart := inst.status != instStatus.PHASING_OUT
		if err != nil {
			inst.status = instStatus.ERROR
		} else {
			inst.status = instStatus.STOPPED
		}
		if restart {
			inst.restartDelay = inst.restartDelay*2 + 1
			go func(restartDelay int) {
				time.Sleep(time.Duration(restartDelay) * time.Second)
				inst.Start()
			}(inst.restartDelay)
		}
	}()

	return nil
}

// Mark an app instance as phasing out - not accepting new users before it can be stopped
func (inst *Instance) PhaseOut() {
	inst.Lock()
	defer inst.Unlock()
	inst.status = instStatus.PHASING_OUT
	if inst.userCount == 0 {
		inst.doStop()
	}
}

// Stop an app instance
func (inst *Instance) doStop() error {
	if inst.cmd != nil && inst.cmd.Process != nil {
		err := inst.cmd.Process.Kill()
		if err != nil {
			return err
		}
	}
	portspool.Release(inst.port)
	return nil
}

// Stop an app instance
func (inst *Instance) Stop() error {
	inst.Lock()
	defer inst.Unlock()
	return inst.doStop()
}

// Increment or decrement user count
func (inst *Instance) SetUserCount(incr int, relative bool) {
	inst.Lock()
	defer inst.Unlock()
	if relative {
		inst.userCount += incr
	} else {
		inst.userCount = incr
	}
}
