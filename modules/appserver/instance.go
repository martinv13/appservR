package appserver

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/martinv13/go-shiny/modules/config"
	"github.com/martinv13/go-shiny/modules/portspool"
)

type appInstance struct {
	AppName string
	AppDir  string
	Status  string
	Port    string
	StdErr  string
	Cmd     *exec.Cmd
	config  config.Config
}

type instStatusType struct {
	STARTING    string
	RUNNING     string
	PHASING_OUT string
	ERROR       string
	STOPPED     string
}

var instStatus = instStatusType{
	STARTING:    "STARTING",
	RUNNING:     "RUNNING",
	PHASING_OUT: "PHASING_OUT",
	ERROR:       "ERROR",
	STOPPED:     "STOPPED",
}

// Start an instance of the app and relaunch when it fails
func (inst *appInstance) Start() error {
	port, err := portspool.GetNext()
	if err != nil {
		return err
	}
	inst.Port = port
	_, err = os.Stat(inst.AppDir)
	if err != nil {
		inst.Status = instStatus.ERROR
		inst.StdErr = "App source directory does not exist"
		return errors.New("App source directory does not exist")
	}
	inst.Status = instStatus.STARTING
	cmd := exec.Command(inst.config.GetString("Rscript"), "-e", "shiny::runApp('.', port="+inst.Port+")")
	cmd.Dir = inst.AppDir
	cmd.Env = os.Environ()

	inst.Cmd = cmd

	stdErr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(stdErr)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			inst.StdErr += line + "\n"
			if strings.HasPrefix(line, "Listening on") {
				inst.Status = instStatus.RUNNING
				fmt.Println("app " + inst.AppName + " at " + inst.Port + " is running")
				return
			}
		}
	}()

	if err = cmd.Start(); err != nil {
		return err
	}

	go func() {
		err := cmd.Wait()
		if err != nil {
		}
	}()

	return nil
}

// Mark an app instance as phasing out - not accepting new users before it can be stopped
func (inst *appInstance) PhaseOut() {
	if inst.Status == instStatus.RUNNING {
		inst.Status = instStatus.PHASING_OUT
	} else {
		inst.Stop()
	}
}

// Stop an app instance
func (inst *appInstance) Stop() error {
	if inst.Cmd != nil && inst.Cmd.Process != nil {
		err := inst.Cmd.Process.Kill()
		if err != nil {
			return err
		}
	}
	inst.Status = instStatus.STOPPED
	portspool.Release(inst.Port)
	return nil
}
