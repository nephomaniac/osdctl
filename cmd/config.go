package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/openshift/osdctl/pkg/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// configCmd is the subcommand "osdctl config" for cobra.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display or update the current configuration",
	Long:  "Display the viper configuration yaml in use, or update a config value with --key and --value flags",
	RunE:  manageConfig,
}

func init() {
	configCmd.Flags().String("key", "", "Configuration key to get or set")
	configCmd.Flags().String("value", "", "Configuration value to set")
}

// manageConfig displays or updates the viper configuration
func manageConfig(cmd *cobra.Command, args []string) error {
	key, _ := cmd.Flags().GetString("key")

	// Check if key flag was provided
	if cmd.Flags().Changed("key") {
		// Key was provided, check if value was also provided
		if cmd.Flags().Changed("value") {
			// Both key and value provided - set the value
			value, _ := cmd.Flags().GetString("value")
			return setConfigValue(key, value)
		}
		// Only key provided - get the value
		return getConfigValue(key)
	}

	// Otherwise, display the config
	return showConfig(cmd, args)
}

// setConfigValue sets a configuration value and writes it to the config file
func setConfigValue(key, value string) error {
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		return fmt.Errorf("no config file in use")
	}

	// Set the value in viper
	viper.Set(key, value)

	// Write the config to file
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("error writing config to file: %w", err)
	}

	fmt.Printf("Successfully set '%s' = '%s' in %s\n", key, value, configFile)
	return nil
}

// getValueFromConfigFile reads the config file directly and extracts the value for a key
func getValueFromConfigFile(key string) (interface{}, bool) {
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		return nil, false
	}

	// Read the config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, false
	}

	// Parse the YAML
	var configData map[string]interface{}
	if err := yaml.Unmarshal(data, &configData); err != nil {
		return nil, false
	}

	// Look for the key (handle nested keys if needed)
	value, exists := configData[key]
	return value, exists
}

// compareValues compares two values for equality, handling type conversions
func compareValues(v1, v2 interface{}) bool {
	// Convert both to strings for comparison
	s1 := fmt.Sprintf("%v", v1)
	s2 := fmt.Sprintf("%v", v2)
	return s1 == s2
}

// getConfigSource determines if a value comes from an environment variable, config file, or other source
func getConfigSource(key string) string {
	// Get the current value from viper
	viperValue := viper.Get(key)

	// Check environment variable
	envKey := strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
	envValue, envExists := os.LookupEnv(envKey)

	// Check config file value
	configFileValue, configFileExists := getValueFromConfigFile(key)

	// Priority order for viper (by default):
	// 1. Explicit Set
	// 2. Flags
	// 3. Environment variables
	// 4. Config file
	// 5. Defaults

	// Check if environment variable matches viper value
	if envExists && compareValues(envValue, viperValue) {
		// Double-check: if config file also has this value, env takes priority
		if configFileExists && compareValues(configFileValue, viperValue) {
			// Both match - viper prefers env over config file
			return "environment"
		}
		return "environment"
	}

	// Check if config file value matches viper value
	if configFileExists && compareValues(configFileValue, viperValue) {
		return "config file"
	}

	// If config file exists but doesn't match, value is from elsewhere
	if configFileExists {
		return "other (flags/explicit set/default)"
	}

	// If environment variable exists but doesn't match
	if envExists {
		return "other (flags/explicit set/default)"
	}

	// Value not in config file or env - likely a default or explicitly set
	return "default/runtime set"
}

// getConfigValue retrieves and displays a specific configuration value
func getConfigValue(key string) error {
	if !viper.IsSet(key) {
		return fmt.Errorf("configuration key '%s' not found", key)
	}

	value := viper.Get(key)
	source := getConfigSource(key)
	printer.PrintfGreen("\"%s\": \"%v\" ", key, value)
	fmt.Printf("(source: %s)\n", source)
	return nil
}

// showConfig displays the current viper configuration as YAML with source information
func showConfig(cmd *cobra.Command, args []string) error {
	// Print the config file path
	configFile := viper.ConfigFileUsed()
	if configFile != "" {
		fmt.Printf("Using config file:")
		printer.PrintfGreen(" '%s'\n\n", configFile)
	}
	fmt.Printf("VIPER SETTINGS:")

	// Get all keys to check their sources
	allKeys := viper.AllKeys()

	// Display each setting with its source
	for _, key := range allKeys {
		value := viper.Get(key)
		source := getConfigSource(key)
		printer.PrintfGreen("%s: %v ", key, value)
		fmt.Printf("(source: %s)\n", source)
	}

	return nil
}
