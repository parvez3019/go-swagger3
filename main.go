package main

import (
	"log"
	"os"

	appPkg "github.com/hanyue2020/go-swagger3/app"
)

func main() {
	app := appPkg.NewApp()
	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Error: %s", err.Error())
	}
}
