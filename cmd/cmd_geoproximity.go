package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// geoproximityCmd represents the geoproximity command
var geoproximityCmd = &cobra.Command{
	Use:   "geoproximity",
	Short: "geoproximity configuration",
}

// geoproximityDiscoverCmd represents the discover sonar command
var geoproximityDiscoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "fetch GeoProximity configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		outputFile, err := cmd.Flags().GetString("output")
		if err != nil {
			return err
		}

		proximities, err := GetGeoProximities()
		if err != nil {
			return err
		}
		logger.Printf("Found %d GeoProximities\n", len(proximities))

		return writeDiscoveryResult(proximities, outputFile)
	},
}

// geoproximitySyncCmd represents the sync sonar command
var geoproximitySyncCmd = &cobra.Command{
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

		config, err := getConfig(configFile)
		if err != nil {
			return err
		}

		// Handle Sonar HTTP Checks
		geops, err := GetGeoProximities()
		if err != nil {
			return err
		}
		activeGeoPs := toResourceMatcher(geops)
		expectedGeoPs := toResourceMatcher(config.GeoProximities)
		err = Sync(expectedGeoPs, activeGeoPs, doit, allowRemoving, "Geoproximities")
		if err != nil {
			return err
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
	rootCmd.AddCommand(geoproximityCmd)

	geoproximityCmd.AddCommand(geoproximityDiscoverCmd)
	geoproximityDiscoverCmd.PersistentFlags().StringP("output", "o", "", "write output in yaml format to file, filepath")

	geoproximityCmd.AddCommand(geoproximitySyncCmd)
	geoproximitySyncCmd.PersistentFlags().StringP("config", "c", "", "configuration file, filepath")
	geoproximitySyncCmd.PersistentFlags().Bool("doit", false, "apply planned changes")
	geoproximitySyncCmd.PersistentFlags().Bool("remove", false, "remove resources which are not present in configuration file")
}
