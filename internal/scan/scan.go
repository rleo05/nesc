package scan

import "fmt"

type Options struct {
	Args  []string
	Ports string
}

func Run(opt Options) error {
	if len(opt.Args) == 0 {
		return fmt.Errorf("missing address arg")
	}

	fmt.Printf("Scanning ports in the range of %s\n", opt.Ports)

	return nil
}
