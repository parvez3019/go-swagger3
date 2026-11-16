package app

import (
	"runtime/debug"

	"github.com/parvez3019/go-swagger3/formatter"
	parserPkg "github.com/parvez3019/go-swagger3/parser"
	"github.com/parvez3019/go-swagger3/writer"
	"github.com/urfave/cli"
)

// version is set at build time by GoReleaser via
// -ldflags "-X github.com/parvez3019/go-swagger3/app.version=...".
// When empty (e.g. `go install ...@v1.2.2`), it falls back to the module
// version recorded in the binary's build info.
var version = ""

func resolveVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

type App struct {
	*cli.App
}

func NewApp() *App {
	cliApp := cli.NewApp()
	cliApp.Name = "go-swagger3"
	cliApp.Version = resolveVersion()
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
