package main

import (
	appPkg "github.com/parvez3019/go-swagger3/app"
	"log"
)

func main() {
	app := appPkg.NewApp()

	args := []string{
		"go-swagger3",
		"--module-path", "../dealer",
		"--main-file-path", "../dealer/cmd/server/main.go",
		"--output", "../dealer/documentation-spec.yaml",
		"--schema-without-pkg",
		"--generate-yaml", "true",
	}
	err := app.Run(args)
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
