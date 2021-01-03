package appproxy

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"github.com/martinv13/go-shiny/services/portspool"
)

const appDir = "C:/Users/marti/code/shiny-apps/shiny-apps/test-app"
const rscriptPath = "C:/Program Files/R/R-4.0.3/bin/Rscript.exe"

type shinyAppInstance struct {
	AppID string
	State string
	Port  string
}

func SpawnApp(appID string, appDir string) (*shinyAppInstance, error) {
	port, portErr := portspool.GetNext()
	if portErr != nil {
		return nil, portErr
	}
	inst := shinyAppInstance{
		AppID: appID,
		State: "STARTING",
		Port:  port,
	}
	cmd := exec.Command(rscriptPath, "-e", "shiny::runApp('.', port="+inst.Port+")")
	cmd.Dir = appDir

	stdErr, _ := cmd.StderrPipe()
	scanner := bufio.NewScanner(stdErr)
	go func() {
		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "Listening on") {
				inst.State = "RUNNING"
				fmt.Println("app " + inst.AppID + " at " + inst.Port + " is running")
			}
		}
	}()

	fmt.Println("Spawning app " + appID + " at port " + inst.Port)
	cmdErr := cmd.Start()
	if cmdErr != nil {
		return nil, cmdErr
	}
	return &inst, nil
}
