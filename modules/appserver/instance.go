package appserver

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/appservR/appservR/modules/config"
	"github.com/appservR/appservR/modules/portspool"
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
	STOPPING    string
	ERROR       string
	STOPPED     string
}{
	STARTING:    "STARTING",
	RUNNING:     "RUNNING",
	PHASING_OUT: "PHASING_OUT",
	STOPPING:    "STOPPING",
	ERROR:       "ERROR",
	STOPPED:     "STOPPED",
}

// Create a new instance of the app
func NewInstance(appName string, appDir string, conf config.Config) *Instance {
	return &Instance{
		ID:      uuid.NewV4().String()[0:6],
		appName: appName,
		appDir:  appDir,
		config:  conf,
		status:  instStatus.STOPPED,
	}
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

// Get instance user count
func (inst *Instance) UserCount() int {
	inst.RLock()
	defer inst.RUnlock()
	return inst.userCount
}

// Get instance console output
func (inst *Instance) StdErr() string {
	inst.RLock()
	defer inst.RUnlock()
	return inst.stdErr
}

// Start an instance of the app and relaunch when it fails
func (inst *Instance) Start() error {
	inst.Lock()
	defer inst.Unlock()
	logger := inst.config.Logger()
	port, err := portspool.GetNext()
	if err != nil {
		return err
	}
	inst.port = port
	_, err = os.Stat(inst.appDir)
	if err != nil {
		inst.status = instStatus.ERROR
		inst.stdErr = "App source directory does not exist"
		return errors.New("app source directory does not exist")
	}
	inst.status = instStatus.STARTING
	cmd := exec.Command(inst.config.GetString("Rscript"), "-e", "shiny::runApp('.', port="+inst.port+")")
	inst.cmd = cmd
	cmd = configCmd(cmd)
	cmd.Dir = inst.appDir
	cmd.Env = os.Environ()
	stdOut, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stdErr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	// Goroutine to read output and save stderr
	go func() {
		merged := io.MultiReader(stdErr, stdOut)
		scanner := bufio.NewScanner(merged)
		for scanner.Scan() {
			line := scanner.Text()
			inst.Lock()
			inst.stdErr += line + "\n"
			if strings.HasPrefix(line, "Listening on") {
				inst.status = instStatus.RUNNING
				logger.Info("app " + inst.appName + " at " + inst.port + " is running (" + inst.ID + ")")
				inst.Unlock()
				return
			}
			inst.Unlock()
		}
	}()

	// Actually starting the subprocess
	logger.Info("starting app " + inst.appName + " (" + inst.ID + ")")
	if err = cmd.Start(); err != nil {
		return err
	}

	// Goroutine to restart the instance on stop
	go func() {
		err := cmd.Wait()
		inst.Lock()
		defer inst.Unlock()
		if inst.status == instStatus.STOPPING {
			logger.Info(inst.appName + " instance stopped (" + inst.ID + ")")
			inst.status = instStatus.STOPPED
		} else {
			if err != nil {
				logger.Info(inst.appName + " instance exited with error (" + inst.ID + ")")
				logger.Info(err.Error())
				inst.status = instStatus.ERROR
			} else {
				logger.Info(inst.appName + " instance exited successfully (" + inst.ID + ")")
				inst.status = instStatus.STOPPED
			}
			inst.restartDelay = inst.restartDelay*2 + 1
			time.AfterFunc(time.Duration(inst.restartDelay)*time.Second, func() { inst.Start() })
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
		inst.status = instStatus.STOPPING
		inst.doStop()
	}
}

// Stop an app instance
func (inst *Instance) doStop() error {
	if inst.cmd != nil {
		err := killCmd(inst.cmd)
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
	inst.status = instStatus.STOPPING
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
	if inst.userCount < 0 {
		inst.userCount = 0
	}
}
