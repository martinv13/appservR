package shinyapp

import (
	"log"
	"os/exec"
)

//const appDir = "/home/martin/Projets/shiny-proxy/sample-app/shiny-apps/sample-app1"
//const rscriptPath = "/usr/lib/R/bin/Rscript"
const appDir = "C:/Users/Martin/Documents/Projects/shiny-apps/shiny-apps/test-app"
const rscriptPath = "C:/Program Files/R/R-3.6.3/bin/Rscript.exe"

func RunShinyApp() {
	cmd := exec.Command(rscriptPath, "-e", "shiny::runApp('.', port=3053)")
	cmd.Dir = appDir
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
}
