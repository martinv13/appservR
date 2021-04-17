package appproxy

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/martinv13/go-shiny/models"
	"github.com/martinv13/go-shiny/modules/portspool"
	"github.com/spf13/viper"
)

type appInstance struct {
	App    *models.ShinyApp
	Status string
	Port   string
	StdErr string
	Cmd    *exec.Cmd
}

// Start an instance of the app and relaunch when it fails
func (inst *appInstance) Start() error {
	port, err := portspool.GetNext()
	if err != nil {
		return err
	}
	inst.Port = port
	inst.Status = "STARTING"
	cmd := exec.Command(viper.GetString("Rscript"), "-e", "shiny::runApp('.', port="+inst.Port+")")
	cmd.Dir = inst.App.AppDir
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
				inst.Status = "RUNNING"
				fmt.Println("app " + inst.App.AppName + " at " + inst.Port + " is running")
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
	if inst.Status == "RUNNING" {
		inst.Status = "PHASEOUT"
	} else {
		inst.Stop()
	}
}

// Stop an app instance
func (inst *appInstance) Stop() error {
	err := inst.Cmd.Process.Kill()
	if err != nil {
		return err
	}
	inst.Status = "STOPPED"
	portspool.Release(inst.Port)
	return nil
}
