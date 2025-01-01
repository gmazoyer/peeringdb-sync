package cmd

import (
	"fmt"

	"github.com/gmazoyer/peeringdb-sync/database"
	"github.com/spf13/cobra"
)

func init() {
	databaseInitCmd.Flags().BoolP("clean", "c", false, "Remove the existing database before initializing it again")

	databaseCmd.AddCommand(databaseInitCmd)
	databaseCmd.AddCommand(databaseDeleteCmd)
	databaseCmd.AddCommand(databaseClearCmd)
	rootCmd.AddCommand(databaseCmd)
}

var databaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Perform database operations",
	Long:  `Create, delete or clear the database.`,
}

var databaseInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database",
	Long:  `Initialize the database with the schema, creating required tables to store records.`,
	Run: func(cmd *cobra.Command, args []string) {
		removeExisting, _ := cmd.Flags().GetBool("clean")
		db, err := database.CreateDatabase(PeeringdbDbFile, removeExisting)
		if err != nil {
			fmt.Printf("Failed to initialize the database: %s\n", err.Error())
			return
		}

		defer db.Close()

		_, err = database.CreateDatabaseSchema(db, database.GetSchema())
		if err != nil {
			fmt.Printf("Failed to create the database schema: %s\n", err.Error())
			return
		}
	},
}

var databaseDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the database",
	Long:  `Delete the database and all its content.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := database.DeleteDatabase(PeeringdbDbFile); err != nil {
			fmt.Printf("Failed to delete the database: %s\n", err.Error())
		}
	},
}

var databaseClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the database",
	Long:  `Clear the database content, keeping the schema.`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.GetDatabaseConnection(PeeringdbDbFile)
		if err != nil {
			fmt.Printf("Failed to connect to the database: %s\n", err.Error())
			return
		}

		defer db.Close()

		if err = database.ClearDatabase(db, database.GetSchema()); err != nil {
			fmt.Printf("Failed to clear the database: %s\n", err.Error())
		}
	},
}
