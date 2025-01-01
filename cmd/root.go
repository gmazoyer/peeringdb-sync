package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var PeeringdbApiKey, PeeringdbDbFile string
var rootCmd = &cobra.Command{
	Use:   "peeringdb-sync",
	Short: "Synchronize PeeringDB records locally",
	Long:  `Synchronize PeeringDB data to a local database. The use of an API key is highly recommended.`,
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&PeeringdbApiKey, "api-key", "k", getEnv("PEERINGDB_API_KEY", ""), "PeeringDB API key to use for authentication")
	rootCmd.PersistentFlags().StringVarP(&PeeringdbDbFile, "file", "f", getEnv("PEERINGDB_DATABASE_FILE", "peeringdb.db"), "Path to the file to use as SQLite database")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
