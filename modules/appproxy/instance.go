package appproxy

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/martinv13/go-shiny/modules/portspool"
)

const rscriptPath = "C:/Program Files/R/R-4.0.3/bin/Rscript.exe"

type shinyAppInstance struct {
	AppID  string
	State  string
	Port   string
	StdErr string
}

func SpawnApp(appID string, appDir string) (*shinyAppInstance, error) {
	port, err := portspool.GetNext()
	if err != nil {
		return nil, err
	}
	inst := shinyAppInstance{
		AppID: appID,
		State: "STARTING",
		Port:  port,
	}
	cmd := exec.Command(rscriptPath, "-e", "shiny::runApp('.', port="+inst.Port+")")
	cmd.Dir = appDir
	cmd.Env = append(os.Environ(),
		"R_LIBS_USER=C:\\Users\\marti\\R\\win-library\\4.0",
	)
	stdErr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(stdErr)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			inst.StdErr += line + "\n"
			if strings.HasPrefix(line, "Listening on") {
				inst.State = "RUNNING"
				fmt.Println("app " + inst.AppID + " at " + inst.Port + " is running")
				return
			}
		}
	}()

	fmt.Println("Spawning app " + appID + " at port " + inst.Port)

	if err = cmd.Start(); err != nil {
		return nil, err
	}
	return &inst, nil
}
