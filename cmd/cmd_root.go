package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

const traceLogPath = "/tmp/mech.log"

var debug bool
var trace bool
var constellixAPIKey string
var constellixSecretKey string

var logger *log.Logger
var reportToTestBuffer bool
var testBuffer *bytes.Buffer

var loggerCleanup = func() {}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "mech",
	Short:   "Constellix DNS configuration as code",
	Version: Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		level, tracePath := "", ""
		switch {
		case trace:
			level, tracePath = "trace", traceLogPath
		case debug:
			level = "debug"
		}
		loggerCleanup = initLogger(tracePath, level)

		constellixAPIKey = os.Getenv("CONSTELLIX_API_KEY")
		constellixSecretKey = os.Getenv("CONSTELLIX_SECRET_KEY")
		if constellixAPIKey == "" || constellixSecretKey == "" {
			return fmt.Errorf("provide CONSTELLIX_API_KEY and CONSTELLIX_SECRET_KEY environmental variables")
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	loggerCleanup()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	logger = log.New(os.Stdout, "", 0)
	testBuffer = new(bytes.Buffer)
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false,
		"Debug-level logging on stderr.")
	rootCmd.PersistentFlags().BoolVar(&trace, "trace", false,
		"Trace-level logs to "+traceLogPath+" (truncated each run).")
	// Pre-register --version without a shorthand; otherwise cobra defaults
	// to adding -v as an alias.
	rootCmd.Flags().Bool("version", false, "version for mech")
}
