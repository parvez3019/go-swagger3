package app

import "github.com/urfave/cli"

type args struct {
	flags []cli.Flag

	modulePath       string
	mainFilePath     string
	handlerPath      string
	output           string
	banStrings       []string
	debug            bool
	strict           bool
	schemaWithoutPkg bool
	generateYaml     bool
}

func LoadArgs(c *cli.Context) *args {
	return &args{
		flags:            flags,
		modulePath:       c.GlobalString("module-path"),
		mainFilePath:     c.GlobalString("main-file-path"),
		handlerPath:      c.GlobalString("handler-path"),
		output:           c.GlobalString("output"),
		banStrings:       c.GlobalStringSlice("hide-refs"),
		debug:            c.GlobalBool("debug"),
		strict:           c.GlobalBool("strict"),
		schemaWithoutPkg: c.GlobalBool("schema-without-pkg"),
		generateYaml:     c.GlobalBool("generate-yaml"),
	}
}

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
	cli.StringSliceFlag{
		Name:  "hide-refs",
		Value: nil,
		Usage: "Hide refs prefixes in custom types",
	},
	cli.BoolFlag{
		Name:  "debug",
		Usage: "show debug message",
	},
	cli.BoolFlag{
		Name:  "strict",
		Usage: "convert go parsing warnings to fatal errors",
	},
	cli.BoolFlag{
		Name:  "schema-without-pkg",
		Usage: "create schemas without package name append to the name",
	},
	cli.BoolFlag{
		Name:  "generate-yaml",
		Usage: "generate yaml spec if true",
	},
}
