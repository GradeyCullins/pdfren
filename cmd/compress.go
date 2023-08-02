package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	compressionLevel string
	isEstimate       bool

	rootCmd = &cobra.Command{
		Use:   "pdffren",
		Short: "pdffren compresses your PDF using Adobe's online PDF compressor tool",
		Long:  "pdffren compresses your PDF using Adobe's online PDF compressor tool",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello world from cobra")
		},
	}
)

func Execute() {
	rootCmd.PersistentFlags().StringVar(&compressionLevel, "compression", "high", "set the compression level \"high|medium|low\"")
	rootCmd.PersistentFlags().BoolVar(&isEstimate, "estimate", false, "when enabled, return estimated file saving when compressing")
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
