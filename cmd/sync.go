package cmd

import (
	"fmt"
	"sync"

	"github.com/gmazoyer/peeringdb"
	"github.com/gmazoyer/peeringdb-sync/database"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func init() {
	rootCmd.AddCommand(syncCmd)
}

type task struct {
	name         string
	dependencies []task
	function     func(bar *mpb.Bar)
}

func (t *task) execute(wg *sync.WaitGroup, dependencyChannels []<-chan struct{}, bar *mpb.Bar) {
	defer wg.Done()

	// Wait for dependencies to complete
	for _, dep := range dependencyChannels {
		<-dep
	}

	t.function(bar)
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
		var opts []peeringdb.Option
		if PeeringdbApiKey != "" {
			opts = append(opts, peeringdb.WithAPIKey(PeeringdbApiKey))
		}
		api := peeringdb.NewAPI(opts...)

		var wg sync.WaitGroup
		doneChannels := make(map[string]chan struct{})
		progress := mpb.New(mpb.WithWaitGroup(&wg), mpb.WithAutoRefresh())

		s := database.Synchronization{API: api, DB: db}

		orgTask := task{name: "Organizations", function: s.SynchronizeOrganizations}
		campusTask := task{name: "Campuses", function: s.SynchronizeCampuses, dependencies: []task{orgTask}}
		facTask := task{name: "Facilities", function: s.SynchronizeFacilities, dependencies: []task{orgTask, campusTask}}
		carrierTask := task{name: "Carriers", function: s.SynchronizeCarriers, dependencies: []task{orgTask}}
		carrierFacTask := task{name: "Carrier Facilities", function: s.SynchronizeCarrierFacilities, dependencies: []task{carrierTask, facTask}}
		netTask := task{name: "Networks", function: s.SynchronizeNetworks, dependencies: []task{orgTask}}
		ixTask := task{name: "Internet Exchanges", function: s.SynchronizeInternetExchanges, dependencies: []task{orgTask}}
		ixfacTask := task{name: "Internet Exchange Facilities", function: s.SynchronizeInternetExchangeFacilities, dependencies: []task{facTask, ixTask}}
		ixlanTask := task{name: "Internet Exchange LANs", function: s.SynchronizeInternetExchangeLANs, dependencies: []task{ixTask}}
		ixpfxTask := task{name: "Internet Exchange Prefixes", function: s.SynchronizeInternetExchangePrefixes, dependencies: []task{ixlanTask}}
		pocTask := task{name: "Network Contacts", function: s.SynchronizeNetworkContacts, dependencies: []task{netTask}}
		netfacTask := task{name: "Network Facilities", function: s.SynchronizeNetworkFacilities, dependencies: []task{netTask, facTask}}
		netixlanTask := task{name: "Network Internet Exchange LANs", function: s.SynchronizeNetworkInternetExchangeLANs, dependencies: []task{netTask, ixTask, ixlanTask}}

		tasks := []task{orgTask, campusTask, facTask, carrierTask, carrierFacTask, netTask, ixTask, ixfacTask, ixlanTask, ixpfxTask, pocTask, netfacTask, netixlanTask}

		// Create all done channels upfront to avoid a data race between
		// the main goroutine writing to the map and worker goroutines
		// reading from it.
		for _, t := range tasks {
			doneChannels[t.name] = make(chan struct{})
		}

		for _, t := range tasks {
			wg.Add(1)

			var dependencyChannels []<-chan struct{}
			for _, dep := range t.dependencies {
				dependencyChannels = append(dependencyChannels, doneChannels[dep.name])
			}

			bar := progress.AddBar(0, // Will be set with task function, but requires manual complete trigger
				mpb.PrependDecorators(
					decor.Name(fmt.Sprintf("%-31s", t.name), decor.WC{C: decor.DindentRight | decor.DextraSpace}),
					decor.Name("fetching", decor.WCSyncSpaceR),
					decor.CountersNoUnit("%d / %d", decor.WCSyncWidth),
				),
				mpb.AppendDecorators(
					decor.OnComplete(decor.Percentage(decor.WC{W: 5}), "done"),
				),
			)

			done := doneChannels[t.name]
			go func(t task, d []<-chan struct{}, b *mpb.Bar) {
				t.execute(&wg, d, b)
				close(done) // Signal task is done
			}(t, dependencyChannels, bar)
		}

		// Wait for all tasks to complete
		progress.Wait()
	},
}
