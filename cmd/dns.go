package cmd

import (
	"github.com/rleo05/nesc/internal/dns"
	"github.com/spf13/cobra"
)

var (
	wordListPath string
	concurrency  int
	protocol     string
)

var dnsCmd = &cobra.Command{
	Use:   "dns [address]",
	Args:  cobra.ArbitraryArgs,
	Short: "Search for subdomains based on a domain and a wordlist",
	RunE: func(cmd *cobra.Command, args []string) error {
		opt := dns.Options{
			Args:         args,
			WordListPath: wordListPath,
			Concurrency:  concurrency,
			Protocol:     protocol,
		}
		return dns.Run(opt)
	},
}

func init() {
	dnsCmd.Flags().StringVarP(&wordListPath, "wordlist", "w", "",
		"Path to the wordlist file containing subdomains")
	dnsCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 10,
		"Number of concurrent workers")
	dnsCmd.Flags().StringVarP(&protocol, "protocol", "p", "udp",
		"Network protocol to use for DNS lookup (udp/tcp)")

	rootCmd.AddCommand(dnsCmd)
}
