package main

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Nigel2392/doccer/doccer"
)

func matchCommand(d *doccer.Doccer, command string) (err error) {
	switch strings.ToLower(command) {
	case "build":
		if err = d.Load(); err != nil {
			return err
		}
		return d.Build()
	case "serve":
		if err = d.Load(); err != nil {
			return err
		}
		return d.Serve()
	case "init":
		return d.Init()
	default:
		return errors.New("command not found")
	}
}

//go:embed assets/*
//go:embed assets/static/*
//go:embed assets/templates/*
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
	err = matchCommand(d, command)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
