package dns

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Options struct {
	Args         []string
	WordListPath string
	Concurrency  int
	Protocol     string
}

func Run(opt Options) error {
	if len(opt.Args) <= 0 {
		return fmt.Errorf("missing address arg")
	}

	if len(opt.WordListPath) == 0 {
		return fmt.Errorf("missing word list path. Provide a path using --wordlist or -w")
	}

	if filepath.Ext(opt.WordListPath) != ".txt" {
		return fmt.Errorf("invalid wordlist extension. Allowed extensions: (txt)")
	}

	if _, err := os.Stat(opt.WordListPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("wordlist file '%s' not found", opt.WordListPath)
		}
		return fmt.Errorf("cannot acess wordlist file: %w", err)
	}

	p := strings.ToLower(opt.Protocol)
	if p != "udp" && p != "tcp" {
		return fmt.Errorf("invalid protocol flag: '%s'. Use 'udp' or 'tcp'", opt.Protocol)
	}
	opt.Protocol = p

	if opt.Concurrency < 1 || opt.Concurrency > 200 {
		return fmt.Errorf("invalid workers value: '%d'. Allowed range: 1-200", opt.Concurrency)
	}

	return nil
}
