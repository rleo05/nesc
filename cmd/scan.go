package cmd

import (
	"github.com/rleo05/nesc/internal/scan"
	"github.com/spf13/cobra"
)

var ports string
var scanCmd = &cobra.Command{
	Use:     "scan [address]",
	Args:    cobra.ArbitraryArgs,
	Aliases: []string{"scanner"},
	Short:   "Scan ports on an ip address within a defined port range",
	RunE: func(cmd *cobra.Command, args []string) error {
		return scan.Execute(args, ports)
	},
}

func init() {
	scanCmd.Flags().StringVarP(&ports, "ports", "p", "",
		"Range of ports to scan")

	rootCmd.AddCommand(scanCmd)
}
