package main

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/Nigel2392/doccer/doccer"
)

var LOAD_REQUIRED_COMMANDS = []string{
	"build",
	"serve",
}

func matchCommand(d *doccer.Doccer, command string, args []string) (err error) {

	if slices.Contains(LOAD_REQUIRED_COMMANDS, command) {
		if err = d.Load(); err != nil {
			return err
		}
	}

	// Parse the arguments for the command
	d.ParseArgs(args)

	// Execute the command
	switch strings.ToLower(command) {
	case "build":
		return d.Build()
	case "serve":
		return d.Serve()
	case "init":
		return d.Init()
	default:
		return errors.New("command not found, try 'build -h', 'serve -h' or 'init -h'")
	}
}

//go:embed assets/*
//go:embed assets/static/*
//go:embed assets/static/bootstrap-icons/*.svg
//go:embed assets/templates/*.tmpl
//go:embed assets/templates/hooks/*.tmpl
var embedFS embed.FS

func main() {

	var d, err = doccer.NewDoccer(embedFS, "doccer.yaml")
	if err != nil && !errors.Is(err, doccer.ErrNoConfig) {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("command required")
		os.Exit(1)
	}

	var command = os.Args[1]
	err = matchCommand(d, command, os.Args[2:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
