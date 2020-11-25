package shinyapp

import (
	"log"
	"os/exec"
)

const appDir = "/home/martin/Projets/shiny-proxy/sample-app/shiny-apps/sample-app1"
const rscriptPath = "/usr/lib/R/bin/Rscript"

func RunShinyApp() {
	cmd := exec.Command(rscriptPath, "-e", "shiny::runApp('.', port=3000)")
	cmd.Dir = appDir
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
}
