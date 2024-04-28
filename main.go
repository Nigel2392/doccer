package main

import (
	"errors"
	"os"
	"strings"

	"github.com/Nigel2392/doccer/doccer"
)

func matchCommand(d *doccer.Doccer, command string) error {
	switch strings.ToLower(command) {
	case "build":
		return d.Build()
	case "serve":
		return d.Serve()
	default:
		return errors.New("command not found")
	}
}

func main() {

	var d, err = doccer.NewDoccer("doccer.yaml")
	if err != nil {
		panic(err)
	}

	if len(os.Args) < 2 {
		panic("command is required")
	}

	var command = os.Args[1]
	err = matchCommand(d, command)
	if err != nil {
		panic(err)
	}

}
