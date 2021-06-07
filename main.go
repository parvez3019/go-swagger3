package main

import (
	appPkg "github.com/parvez3019/go-swagger3/app"
	"log"
	"os"
)

func main() {
	app := appPkg.NewApp()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
