package main

import (
	appPkg "github.com/parvez3019/go-swagger3/app"
	"log"
	"os"
)

func main() {
	app := appPkg.NewApp()
	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Error: %s", err.Error())
	}
}
