package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/KirillZiborov/GophKeeper/pkg/token"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	cfgFile      string
	tokenStorage token.Storage
)

// rootCmd represents the base gophkeeper command.
var rootCmd = &cobra.Command{
	Use:   "gophkeeper",
	Short: "GophKeeper Client",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	tokenStorage = token.NewFileStorage("token.txt")

	// Define flags and configuration settings.
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "client config filepath, default is $HOME/.gophkeeper.yaml")
	rootCmd.PersistentFlags().StringP("grpc_address", "a", "localhost:8080", "address of the GophKeeper server")
	rootCmd.PersistentFlags().StringP("encryption_key", "k", "", "secret encryption key")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// Search config in home and project directory with name ".gophkeeper".
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".gophkeeper")
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Println("Error parsing config file", err)
		}
		fmt.Println("Config file not found")
	} else {
		fmt.Println("Using config file: ", viper.ConfigFileUsed())
	}

	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

	rootCmd.Flags().VisitAll(func(flag *pflag.Flag) {
		key := strings.ReplaceAll(flag.Name, "-", ".")
		if err := viper.BindPFlag(key, flag); err != nil {
			fmt.Println("Failed to bind flag", err)
		}
	})
}
