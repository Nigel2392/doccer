package doccer

import "fmt"

var (

	// ErrNoConfig is returned when there is no config file
	ErrNoConfig = fmt.Errorf("no config file found")
)
