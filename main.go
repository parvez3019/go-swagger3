package main

import (
	"github.com/parvez3019/go-swagger3/parser"
	"github.com/parvez3019/go-swagger3/writer"
	"github.com/urfave/cli"
	"log"
)

var version = "v1.0.1"

var flags = []cli.Flag{
	cli.StringFlag{
		Name:  "module-path",
		Value: "",
		Usage: "go-swagger3 will search @comment under the module",
	},
	cli.StringFlag{
		Name:  "main-file-path",
		Value: "",
		Usage: "go-swagger3 will start to search @comment from this main file",
	},
	cli.StringFlag{
		Name:  "handler-path",
		Value: "",
		Usage: "go-swagger3 only search handleFunc comments under the path",
	},
	cli.StringFlag{
		Name:  "output",
		Value: "oas.json",
		Usage: "output file",
	},
	cli.BoolFlag{
		Name:  "debug",
		Usage: "show debug message",
	},
	cli.BoolFlag{
		Name:  "strict",
		Usage: "convert go parsing warnings to fatal errors",
	},
}

func action(c *cli.Context) error {
	p, err := parser.NewParser(c.GlobalString("module-path"), c.GlobalString("main-file-path"), c.GlobalString("handler-path"), c.GlobalBool("debug"), c.GlobalBool("strict"))
	if err != nil {
		return err
	}
	openApiObject, err := p.Parse()
	if err != nil {
		return err
	}

	fw := writer.NewFileWriter()
	return fw.Write(openApiObject, c.GlobalString("output"))
}

func main() {
	app := cli.NewApp()
	app.Name = "go-swagger3"
	app.Version = version
	app.HideHelp = true
	app.OnUsageError = func(c *cli.Context, err error, isSubcommand bool) error {
		cli.ShowAppHelp(c)
		return nil
	}
	app.Flags = flags
	app.Action = action

	// TODO : remove hardcoded paths
	args := []string{"go-swagger3", "--module-path", "../dealer", "--main-file-path", "../dealer/cmd/server/main.go", "--output", "../dealer/oas.json"}
	err := app.Run(args)
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
