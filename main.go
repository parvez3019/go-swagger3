package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

var version = "v1.0.0"

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
}

func action(c *cli.Context) error {
	p, err := newParser(c.GlobalString("module-path"), c.GlobalString("main-file-path"), c.GlobalString("handler-path"), c.GlobalBool("debug"))
	if err != nil {
		return err
	}
	return p.CreateOASFile(c.GlobalString("output"))
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

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
