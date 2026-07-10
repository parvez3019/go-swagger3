package app

import (
	"github.com/parvez3019/go-swagger3/formatter"
	parserPkg "github.com/parvez3019/go-swagger3/parser"
	"github.com/parvez3019/go-swagger3/writer"
	"github.com/urfave/cli"
)

var version = "v1.0.0"

type App struct {
	*cli.App
}

func NewApp() *App {
	cliApp := cli.NewApp()
	cliApp.Name = "go-swagger3"
	cliApp.Version = version
	cliApp.HideHelp = true
	cliApp.OnUsageError = func(c *cli.Context, err error, isSubcommand bool) error {
		return cli.ShowAppHelp(c)
	}
	cliApp.Flags = flags
	cliApp.Action = action
	cliApp.Commands = []cli.Command{
		{
			Name:  "fmt",
			Usage: "format swagger @ annotations in Go source files",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "d",
					Value: ".",
					Usage: "directory to format",
				},
			},
			Action: fmtAction,
		},
	}

	return &App{
		App: cliApp,
	}
}

func action(c *cli.Context) error {
	args := LoadArgs(c)
	parser, err := parserPkg.NewParser(
		args.modulePath,
		args.mainFilePath,
		args.handlerPath,
		args.debug,
		args.strict,
		args.schemaWithoutPkg,
		args.exclude,
		args.quiet,
	).Init()

	if err != nil {
		return err
	}
	openApiObject, err := parser.Parse()
	if err != nil {
		return err
	}

	fw := writer.NewFileWriter()
	return fw.Write(openApiObject, args.output, args.generateYaml, args.schemaWithoutPkg)
}

func fmtAction(c *cli.Context) error {
	dir := c.String("d")
	if dir == "" {
		dir = "."
	}
	return formatter.FormatDir(dir)
}
