package cmd

import (
	"fmt"

	"github.com/gmazoyer/peeringdb"
	"github.com/gmazoyer/peeringdb-sync/database"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(syncCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize the database with PeeringDB",
	Long:  `Synchronize PeeringDB records within the local database, updating existing records, adding new ones and deleting outdated ones.`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.GetDatabaseConnection(PeeringdbDbFile)
		if err != nil {
			fmt.Printf("Failed to connect to the database: %s\n", err.Error())
			return
		}

		// Prepare to query the API and the synchronization
		var api *peeringdb.API
		if PeeringdbApiKey == "" {
			api = peeringdb.NewAPI()
		} else {
			api = peeringdb.NewAPIWithAPIKey(PeeringdbApiKey)
		}

		s := database.Synchronization{API: api, DB: db}

		fmt.Println("Starting PeeringDB synchronization...")

		s.SynchronizeOrganizations()
		s.SynchronizeCampuses()
		s.SynchronizeFacilities()
		s.SynchronizeCarriers()
		s.SynchronizeNetworks()
		s.SynchronizeInternetExchanges()
		s.SynchronizeInternetExchangeFacilities()
		s.SynchronizeInternetExchangeLANs()
		s.SynchronizeInternetExchangePrefixes()
		s.SynchronizeNetworkContacts()
		s.SynchronizeNetworkFacilities()
		s.SynchronizeNetworkInternetExchangeLANs()

		fmt.Println("PeeringDB fully synchronized.")
	},
}
