package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// dnsCmd represents the dns command
var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "dns configuration",
}

// dnsDiscoverCmd represents the discover sonar command
var dnsDiscoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "fetch configuration from DNS endpoints (domains, records, geo proximities, etc.)",
}

// dnsDiscoverRecordsCmd fetch existing records for a domain name from Constellix
// https://api.dns.constellix.com/v4/docs#tag/Domain-Records/paths/~1domains~1%7Bdomain_id%7D~1records/get
var dnsDiscoverRecordsCmd = &cobra.Command{
	Use:   "records <domain name>",
	Short: "retrieve DNS records for a domain name",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("requires a domain name (e.g. example.com)")
		}

		if len(args) != 1 {
			return fmt.Errorf("cannot discover multiple domain names at once")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		outputFile, err := cmd.Flags().GetString("output")
		if err != nil {
			return err
		}

		domains, err := GetDNSDomains()
		if err != nil {
			return err
		}

		var domainID int

		for _, domain := range domains {
			if domain.Name == args[0] {
				domainID = domain.ID
			}
		}

		if domainID == 0 {
			return fmt.Errorf("domain %s not found", args[0])
		}
		L.Debug("domain found", slog.String("domain", args[0]), slog.Int("id", domainID))
		records, err := GetDNSRecords(domainID)
		if err != nil {
			return err
		}
		logger.Printf("Found %d DNS records\n", len(records))
		return writeDiscoveryResult(records, outputFile)
	},
}

// dnsDiscoverDomainsCmd fetch existing domains from Constellix
var dnsDiscoverDomainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "retrieve domains registered in Constellix",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		domains, err := GetDNSDomains()
		if err != nil {
			return err
		}

		report := table.NewWriter()
		// For tests, render data in csv format
		defer func() {
			if report.Length() > 0 {
				if reportToTestBuffer {
					report.RenderCSV()
				} else {
					report.Render()
				}
			} else {
				logger.Println("  nothing to do")
			}
		}()

		if reportToTestBuffer {
			// Skip header in tests
			report.SetOutputMirror(testBuffer)
		} else {
			report.SetOutputMirror(os.Stdout)
			report.AppendHeader(table.Row{"ID", "Name", "Status", "GeoIP", "GTD"})
		}

		for _, domain := range domains {
			report.AppendRow(table.Row{
				domain.ID, domain.Name, domain.Status, domain.GeoIPEnabled, domain.GTDEnabled,
			})
		}
		return nil
	},
}

// dnsSyncCmd represents the sync DNS command
var dnsSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync configuration to Constellix",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		// Collect flags
		configFile, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}
		if configFile == "" {
			return fmt.Errorf("provide configuration file location via --config argument")
		}

		doit, err := cmd.Flags().GetBool("doit")
		if err != nil {
			return err
		}

		allowRemoving, err := cmd.Flags().GetBool("remove")
		if err != nil {
			return err
		}

		only, err := cmd.Flags().GetString("only")
		if err != nil {
			return err
		}
		if only != "" {
			logger.Printf("syncing only %s domain", only)
		}

		config, err := getConfig(configFile)
		if err != nil {
			return err
		}

		if len(config.DNS) == 0 {
			logger.Println("No DNS configuration found")
			return nil
		}

		domains, err := GetDNSDomains()
		if err != nil {
			return err
		}

		for domainName := range config.DNS {
			if only != "" && only != domainName {
				continue
			}
			var domainID int

			for _, domain := range domains {
				if domain.Name == domainName {
					domainID = domain.ID
				}
			}

			if domainID == 0 {
				return fmt.Errorf("domain %s not found", domainName)
			}
			L.Debug("domain found", slog.String("domain", domainName), slog.Int("id", domainID))
			records, err := GetDNSRecords(domainID)
			if err != nil {
				return fmt.Errorf("sync domain %s: %w", domainName, err)
			}
			for _, item := range config.DNS[domainName] {
				item.domainIDInConstellix = domainID
			}
			activeRecords := toResourceMatcher(records)
			expectedRecords := toResourceMatcher(config.DNS[domainName])
			err = Sync(expectedRecords, activeRecords, doit, allowRemoving, "DNS records for "+domainName)
			if err != nil {
				return fmt.Errorf("sync domain %s: %w", domainName, err)
			}
		}
		var message string
		if !doit {
			message += "apply changes by passing --doit flag"
		}
		if !allowRemoving {
			if message != "" {
				message += "; "
			}
			message += "allow removing of resources by passing --remove flag"
		}
		if message == "" {
			message = "done"
		}
		logger.Println(message)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dnsCmd)
	dnsCmd.AddCommand(dnsDiscoverCmd)
	dnsDiscoverCmd.AddCommand(dnsDiscoverRecordsCmd)
	dnsDiscoverCmd.PersistentFlags().StringP("output", "o", "", "write output in yaml format to file, filepath")

	dnsDiscoverCmd.AddCommand(dnsDiscoverDomainsCmd)

	dnsCmd.AddCommand(dnsSyncCmd)
	dnsSyncCmd.PersistentFlags().StringP("config", "c", "", "configuration file, filepath")
	dnsSyncCmd.PersistentFlags().Bool("doit", false, "apply planned changes")
	dnsSyncCmd.PersistentFlags().Bool("remove", false, "remove resources which are not present in configuration file")
	dnsSyncCmd.PersistentFlags().String("only", "", "execute sync command only for specified domain name")
}
