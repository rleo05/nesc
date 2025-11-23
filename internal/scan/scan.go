package scan

import "fmt"

func Execute(args []string, ports string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing address arg")
	}
	if ports == "" {
		return fmt.Errorf("missing required flag --ports, -p")
	}

	fmt.Printf("Scanning ports in the range of %s\n", ports)

	return nil
}
