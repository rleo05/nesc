package dns

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

type Options struct {
	Args         []string
	WordListPath string
	Concurrency  int
	Protocol     string
}

func Run(opt Options) error {
	var wg sync.WaitGroup
	lineCh := make(chan string, 100)
	errCh := make(chan error, 1)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	err := validateArgs(&opt)
	if err != nil {
		return err
	}

	file, err := os.Open(opt.WordListPath)
	if err != nil {
		return fmt.Errorf("error opening word list file: %w", err)
	}
	defer file.Close()

	go readWordList(ctx, file, errCh, lineCh)

	wg.Add(opt.Concurrency)
	for i := 0; i < opt.Concurrency; i++ {
		go worker(ctx, lineCh, &wg)
	}

	wg.Wait()

	if err := <-errCh; err != nil {
		return err
	}

	return nil
}

func validateArgs(opt *Options) error {
	if len(opt.Args) <= 0 {
		return fmt.Errorf("missing address arg")
	}

	if len(opt.WordListPath) == 0 {
		return fmt.Errorf("missing word list path. Provide a path using --wordlist or -w")
	}

	if filepath.Ext(opt.WordListPath) != ".txt" {
		return fmt.Errorf("invalid wordlist extension. Allowed extensions: txt")
	}

	if _, err := os.Stat(opt.WordListPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("wordlist file '%s' not found", opt.WordListPath)
		}
		return fmt.Errorf("cannot access wordlist file: %w", err)
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

func readWordList(ctx context.Context, file *os.File, errCh chan<- error, lineCh chan<- string) {
	defer close(lineCh)

	scanner := bufio.NewReader(file)
	for {
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
			return
		default:
			textLine, err := scanner.ReadString('\n')
			if err == io.EOF {
				if len(textLine) != 0 {
					lineCh <- strings.TrimRight(textLine, "\r\n")
				}
				errCh <- nil
				return
			}

			if err != nil {
				errCh <- fmt.Errorf("error reading file: %w", err)
				return
			}

			lineCh <- strings.TrimRight(textLine, "\r\n")
		}
	}
}

func worker(ctx context.Context, lineCh <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case line, ok := <-lineCh:
			if !ok {
				return
			}
			fmt.Println(line)
		}
	}
}
