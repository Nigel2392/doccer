package doccer

import "fmt"

var (
	// ErrNoTemplates is returned when there are no templates in a directory
	ErrNoTemplates = fmt.Errorf("no templates found in directory")

	// ErrNoConfig is returned when there is no config file
	ErrNoConfig = fmt.Errorf("no config file found")
)
