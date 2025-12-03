package dns

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Options struct {
	Args         []string
	WordListPath string
	Concurrency  int
	Protocol     string
}

type Env struct {
	Resolver *net.Resolver
	Domain   string
}

func Run(opt Options) error {
	var wg sync.WaitGroup
	lineCh := make(chan string, opt.Concurrency*2)
	errCh := make(chan error, 1)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	err := validateArgs(&opt)
	if err != nil {
		return err
	}
	env := newEnv(opt)

	file, err := os.Open(opt.WordListPath)
	if err != nil {
		return fmt.Errorf("error opening word list file: %w", err)
	}
	defer file.Close()

	go readWordList(ctx, file, errCh, lineCh)

	wg.Add(opt.Concurrency)
	for i := 0; i < opt.Concurrency; i++ {
		go worker(ctx, env, lineCh, &wg)
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

func newEnv(opt Options) *Env {
	dialer := &net.Dialer{
		Timeout: time.Second * 2,
	}
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialer.DialContext(ctx, opt.Protocol, "1.1.1.1:53")
		},
	}

	return &Env{
		Resolver: resolver,
		Domain:   opt.Args[0],
	}
}

func readWordList(ctx context.Context, file *os.File, errCh chan<- error, lineCh chan<- string) {
	defer close(lineCh)
	defer close(errCh)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
			return
		case lineCh <- line:
		}
	}

	if err := scanner.Err(); err != nil {
		errCh <- fmt.Errorf("error reading file: %w", err)
		return
	}

	errCh <- nil
}

func worker(ctx context.Context, env *Env, lineCh <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case line, ok := <-lineCh:
			if !ok {
				return
			}

			subDomainHost := fmt.Sprintf("%s.%s", line, env.Domain)
			ips, err := env.Resolver.LookupHost(ctx, subDomainHost)

			if err == nil {
				fmt.Printf("%s: %s\n", subDomainHost, strings.Join(ips, ","))
			}
		}
	}
}
