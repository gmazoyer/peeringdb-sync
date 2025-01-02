package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gmazoyer/peeringdb"
	"github.com/vbauerster/mpb/v8"
)

func marshalJSON(v interface{}) string {
	m, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(m)
}

// Synchronization is a structure holding pointers to the PeeringDB API and
// database being used.
type Synchronization struct {
	API *peeringdb.API
	DB  *sql.DB
}

// idExistsInTable returns true of the given ID exists in the given database
// table. It returns false if the ID cannot be found.
func (s *Synchronization) idExistsInTable(table string, id int) bool {
	// Look for the given ID in the given table
	result, err := s.DB.Query(fmt.Sprintf("SELECT id FROM %s WHERE id = %d",
		table, id))
	if err != nil {
		log.Fatal(err)
	}
	defer result.Close()

	// Return true if ID was found
	return result.Next()
}

// insertOrUpdateStatement returns a string corresponding to the SQL query to
// be executed. It checks if the ID exists in the database table. If the ID is
// present, the returned string will be an update query. If the ID cannot be
// found, the returned string will be an insert query.
func (s *Synchronization) insertOrUpdateStatement(forceInsert bool, table string, id int, columns []string) string {
	var statement string

	if forceInsert || !s.idExistsInTable(table, id) {
		// ID not found, insert it must be
		statement = fmt.Sprintf("INSERT INTO %s VALUES (%d", table, id)
		for i := 0; i < len(columns); i++ {
			statement += ", ?"
		}
		statement += ")"
	} else {
		// ID found, update it must be
		statement = fmt.Sprintf("UPDATE %s SET ", table)
		for index, column := range columns {
			if index < (len(columns) - 1) {
				statement += fmt.Sprintf("%s = ?, ", column)
			} else {
				statement += fmt.Sprintf("%s = ?", column)
			}
		}
		statement += fmt.Sprintf(" WHERE id = %d", id)
	}

	return statement
}

func (s *Synchronization) removeDeleted(tx *sql.Tx, table string) error {
	// Prepare the statement.
	result, err := tx.Query(fmt.Sprintf("DELETE FROM %s WHERE status == 'deleted'", table))
	if err != nil {
		return err
	}
	return result.Close()
}

// getLastSyncDate retrieves the timestamp at which the last synchronization
// has occured. The returned value is an int64.
func (s *Synchronization) getLastSyncDate(table string) int64 {
	switch {
	case table == "":
		log.Fatal("table not supplied")
	case strings.Contains(table, " "):
		log.Fatal("table malformed")
	}

	updated := time.Unix(0, 0)

	// Query for last sync date
	result, err := s.DB.Query(fmt.Sprintf("SELECT updated FROM %s ORDER BY updated DESC LIMIT 1", table))
	if err != nil {
		log.Fatal(err)
	}
	defer result.Close()

	// Get the value
	if result.Next() {
		result.Scan(&updated)
	}

	return updated.Unix()
}

// executeInsertOrUpdate will execute an update or insert query whether the
// given ID is found in the database table. The given transaction must be
// commited after calling this function. It returns a non-nil error if an issue
// has occured.
func (s *Synchronization) executeInsertOrUpdate(tx *sql.Tx, forceInsert bool, table string, id int, columns []string, values ...interface{}) error {
	// Prepare the database insertion
	statement, err := tx.Prepare(s.insertOrUpdateStatement(forceInsert, table,
		id, columns))
	if err != nil {
		return err
	}
	defer statement.Close()

	// Execute it with the given values
	_, err = statement.Exec(values...)
	if err != nil {
		return err
	}

	return nil
}

func (s *Synchronization) SynchronizeOrganizations(bar *mpb.Bar) {
	table := GetSchema().Tables["peeringdb_organization"]
	since := s.getLastSyncDate(table.Name)
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed organizations objects since the given timestamp
	organizations, err := s.API.GetOrganization(search)
	if err != nil {
		panic(err)
	}

	// Slice is empty, nothing to sync
	if len(*organizations) < 1 {
		fmt.Printf("No organizations to sync since %s.\n", time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	tx, err := s.DB.Begin()
	if err != nil {
		panic(err)
	}

	// Put values in the database
	bar.SetTotal(int64(len(*organizations)), false)
	for _, organization := range *organizations {
		err = s.executeInsertOrUpdate(
			tx, (since == 0), table.Name, organization.ID, table.GetColumnsNames(), organization.Created,
			organization.Updated, organization.Status, organization.Name, organization.NameLong, organization.AKA,
			organization.Website, marshalJSON(organization.SocialMedia), organization.Notes, organization.Address1,
			organization.Address2, organization.Country, organization.City, organization.State, organization.Zipcode,
			organization.Suite, organization.Floor, organization.Latitude, organization.Longitude,
		)
		if err != nil {
			panic(err)
		}

		bar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, table.Name)

	tx.Commit()
	bar.SetTotal(-1, true)
}

func (s *Synchronization) SynchronizeCampuses(bar *mpb.Bar) {
	table := GetSchema().Tables["peeringdb_campus"]
	since := s.getLastSyncDate(table.Name)
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed campuses objects since the given timestamp
	campuses, err := s.API.GetCampus(search)
	if err != nil {
		panic(err)
	}

	// Slice is empty, nothing to sync
	if len(*campuses) < 1 {
		fmt.Printf("No campuses to sync since %s.\n", time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	tx, err := s.DB.Begin()
	if err != nil {
		panic(err)
	}

	// Put values in the database
	bar.SetTotal(int64(len(*campuses)), false)
	for _, campus := range *campuses {
		err = s.executeInsertOrUpdate(
			tx, (since == 0), table.Name, campus.ID, table.GetColumnsNames(), campus.Created, campus.Updated,
			campus.Status, campus.Name, campus.NameLong, campus.AKA, campus.Website, marshalJSON(campus.SocialMedia),
			campus.Notes, campus.Country, campus.City, campus.State, campus.Zipcode, campus.OrganizationID,
		)
		if err != nil {
			panic(err)
		}

		bar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, table.Name)

	tx.Commit()
	bar.SetTotal(-1, true)
}

func (s *Synchronization) SynchronizeFacilities(bar *mpb.Bar) {
	table := GetSchema().Tables["peeringdb_facility"]
	since := s.getLastSyncDate(table.Name)
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed facilities objects since the given timestamp
	facilities, err := s.API.GetFacility(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*facilities) < 1 {
		fmt.Printf("No facilities to sync since %s.\n", time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	tx, err := s.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}

	bar.SetTotal(int64(len(*facilities)), false)
	for _, facility := range *facilities {
		// Put values in the database
		err = s.executeInsertOrUpdate(
			tx, (since == 0), table.Name, facility.ID, table.GetColumnsNames(), facility.Created, facility.Updated,
			facility.Status, facility.Name, facility.AKA, facility.NameLong, facility.Website,
			marshalJSON(facility.SocialMedia), facility.CLLI, facility.Rencode, facility.Npanxx, facility.Notes,
			facility.SalesEmail, facility.SalesPhone, facility.TechEmail, facility.TechPhone,
			marshalJSON(facility.AvailableVoltageServices), facility.DiverseServingSubstations, facility.Property,
			facility.RegionContinent, facility.StatusDashboard, facility.Address1, facility.Address2, facility.City,
			facility.Country, facility.State, facility.Zipcode, facility.Floor, facility.Suite, facility.Latitude,
			facility.Longitude, facility.OrganizationID, facility.CampusID,
		)
		if err != nil {
			log.Fatal(err)
		}

		bar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, table.Name)

	tx.Commit()
	bar.SetTotal(-1, true)
}

func (s *Synchronization) SynchronizeCarriers(bar *mpb.Bar) {
	table := GetSchema().Tables["peeringdb_carrier"]
	since := s.getLastSyncDate(table.Name)
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed carriers objects since the given timestamp
	carriers, err := s.API.GetCarrier(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*carriers) < 1 {
		fmt.Printf("No carriers to sync since %s.\n", time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	tx, err := s.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}

	bar.SetTotal(int64(len(*carriers)), false)
	for _, carrier := range *carriers {
		// Put values in the database
		err = s.executeInsertOrUpdate(
			tx, (since == 0), table.Name, carrier.ID, table.GetColumnsNames(), carrier.Created, carrier.Updated,
			carrier.Status, carrier.Name, carrier.AKA, carrier.NameLong, carrier.Website,
			marshalJSON(carrier.SocialMedia), carrier.Notes, carrier.OrganizationID,
		)
		if err != nil {
			log.Fatal(err)
		}

		bar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, table.Name)

	tx.Commit()
	bar.SetTotal(-1, true)
}
func (s *Synchronization) SynchronizeNetworks(bar *mpb.Bar) {
	table := GetSchema().Tables["peeringdb_network"]
	since := s.getLastSyncDate(table.Name)
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed networks objects since the given timestamp
	networks, err := s.API.GetNetwork(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*networks) < 1 {
		fmt.Printf("No networks to sync since %s.\n", time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	tx, err := s.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}

	bar.SetTotal(int64(len(*networks)), false)
	for _, network := range *networks {
		// Put values in the database
		err = s.executeInsertOrUpdate(
			tx, (since == 0), table.Name, network.ID, table.GetColumnsNames(), network.Created, network.Updated,
			network.Status, network.Name, network.AKA, network.NameLong, network.Website,
			marshalJSON(network.SocialMedia), network.ASN, network.LookingGlass, network.RouteServer,
			network.IRRASSet, network.InfoType, marshalJSON(network.InfoTypes), network.InfoPrefixes4,
			network.InfoPrefixes6, network.InfoTraffic, network.InfoRatio, network.InfoScope, network.InfoUnicast,
			network.InfoMulticast, network.InfoIPv6, network.InfoNeverViaRouteServers, network.Notes,
			network.PolicyURL, network.PolicyGeneral, network.PolicyLocations, network.PolicyRatio,
			network.PolicyContracts, network.AllowIXPUpdate, network.StatusDashboard, network.RIRStatus,
			network.RIRStatusUpdated, network.OrganizationID,
		)
		if err != nil {
			log.Fatal(err)
		}

		bar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, "peeringdb_network")

	tx.Commit()
	bar.SetTotal(-1, true)
}

func (s *Synchronization) SynchronizeInternetExchanges(bar *mpb.Bar) {
	table := GetSchema().Tables["peeringdb_ix"]
	since := s.getLastSyncDate(table.Name)
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed internet exchanges objects since the given timestamp
	ixs, err := s.API.GetInternetExchange(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*ixs) < 1 {
		fmt.Printf("No internet exchanges to sync since %s.\n",
			time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	tx, err := s.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}

	bar.SetTotal(int64(len(*ixs)), false)
	for _, ix := range *ixs {
		// Put values in the database
		err = s.executeInsertOrUpdate(
			tx, (since == 0), table.Name, ix.ID, table.GetColumnsNames(), ix.Created, ix.Updated, ix.Status, ix.Name,
			ix.AKA, ix.NameLong, ix.City, ix.Country, ix.RegionContinent, ix.Media, ix.Notes, ix.ProtoUnicast,
			ix.ProtoMulticast, ix.ProtoIPv6, ix.Website, marshalJSON(ix.SocialMedia), ix.URLStats, ix.TechEmail,
			ix.TechPhone, ix.PolicyEmail, ix.PolicyPhone, ix.SalesEmail, ix.SalesPhone, ix.IxfNetCount,
			ix.IxfLastImport, ix.IxfImportRequest, ix.IxfImportRequestStatus, ix.ServiceLevel, ix.Terms,
			ix.StatusDashboard, ix.OrganizationID,
		)
		if err != nil {
			log.Fatal(err)
		}

		bar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, table.Name)

	tx.Commit()
	bar.SetTotal(-1, true)
}

func (s *Synchronization) SynchronizeInternetExchangeFacilities(bar *mpb.Bar) {
	table := GetSchema().Tables["peeringdb_ix_facility"]
	since := s.getLastSyncDate(table.Name)
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed internet exchange facilities objects since the given
	// timestamp
	ixfacilities, err := s.API.GetInternetExchangeFacility(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*ixfacilities) < 1 {
		fmt.Printf("No internet exchange facilities to sync since %s.\n",
			time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	tx, err := s.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}

	bar.SetTotal(int64(len(*ixfacilities)), false)
	for _, ixfacility := range *ixfacilities {
		// Put values in the database
		err = s.executeInsertOrUpdate(
			tx, (since == 0), table.Name, ixfacility.ID, table.GetColumnsNames(), ixfacility.Created,
			ixfacility.Updated, ixfacility.Status, ixfacility.Name, ixfacility.City, ixfacility.Country,
			ixfacility.InternetExchangeID, ixfacility.FacilityID,
		)
		if err != nil {
			log.Fatal(err)
		}

		bar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, table.Name)

	tx.Commit()
	bar.SetTotal(-1, true)
}

func (s *Synchronization) SynchronizeInternetExchangeLANs(bar *mpb.Bar) {
	table := GetSchema().Tables["peeringdb_ixlan"]
	since := s.getLastSyncDate(table.Name)
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed internet exchange LANs objects since the given timestamp
	ixlans, err := s.API.GetInternetExchangeLAN(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*ixlans) < 1 {
		fmt.Printf("No internet exchange LANs to sync since %s.\n",
			time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	tx, err := s.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}

	bar.SetTotal(int64(len(*ixlans)), false)
	for _, ixlan := range *ixlans {
		// Put values in the database
		err = s.executeInsertOrUpdate(
			tx, (since == 0), table.Name, ixlan.ID, table.GetColumnsNames(), ixlan.Created, ixlan.Updated,
			ixlan.Status, ixlan.Name, ixlan.Description, ixlan.MTU, ixlan.Dot1QSupport, ixlan.RouteServerASN,
			ixlan.ARPSponge, ixlan.IXFIXPMemberListURL, ixlan.IXFIXPMemberListURLVisible, ixlan.IXFIXPImportEnabled,
			ixlan.InternetExchangeID,
		)
		if err != nil {
			log.Fatal(err)
		}

		bar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, table.Name)

	tx.Commit()
	bar.SetTotal(-1, true)
}

func (s *Synchronization) SynchronizeInternetExchangePrefixes(bar *mpb.Bar) {
	table := GetSchema().Tables["peeringdb_ix_prefix"]
	since := s.getLastSyncDate(table.Name)
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed internet exchange prefixes objects since the given timestamp
	ixpfxs, err := s.API.GetInternetExchangePrefix(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*ixpfxs) < 1 {
		fmt.Printf("No internet exchange prefixes to sync since %s.\n",
			time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	tx, err := s.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}

	bar.SetTotal(int64(len(*ixpfxs)), false)
	for _, ixpfx := range *ixpfxs {
		// Put values in the database
		err = s.executeInsertOrUpdate(
			tx, (since == 0), table.Name, ixpfx.ID, table.GetColumnsNames(), ixpfx.Created, ixpfx.Updated,
			ixpfx.Status, ixpfx.Protocol, ixpfx.Prefix, ixpfx.InDFZ, ixpfx.InternetExchangeLANID,
		)
		if err != nil {
			log.Fatal(err)
		}

		bar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, table.Name)

	tx.Commit()
	bar.SetTotal(-1, true)
}

func (s *Synchronization) SynchronizeNetworkContacts(bar *mpb.Bar) {
	table := GetSchema().Tables["peeringdb_network_contact"]
	since := s.getLastSyncDate(table.Name)
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed network contacts objects since the given timestamp
	contacts, err := s.API.GetNetworkContact(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*contacts) < 1 {
		fmt.Printf("No contacts to sync since %s.\n", time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	tx, err := s.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}

	bar.SetTotal(int64(len(*contacts)), false)
	for _, netcontact := range *contacts {
		// Put values in the database
		err = s.executeInsertOrUpdate(
			tx, (since == 0), table.Name, netcontact.ID, table.GetColumnsNames(), netcontact.Created,
			netcontact.Updated, netcontact.Status, netcontact.Role, netcontact.Visible, netcontact.Name,
			netcontact.Phone, netcontact.Email, netcontact.URL, netcontact.NetworkID,
		)
		if err != nil {
			log.Fatal(err)
		}

		bar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, table.Name)

	tx.Commit()
	bar.SetTotal(-1, true)
}

func (s *Synchronization) SynchronizeNetworkFacilities(bar *mpb.Bar) {
	table := GetSchema().Tables["peeringdb_network_facility"]
	since := s.getLastSyncDate(table.Name)
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed network facilities objects since the given timestamp
	netfacilities, err := s.API.GetNetworkFacility(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*netfacilities) < 1 {
		fmt.Printf("No network facilities to sync since %s.\n",
			time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	tx, err := s.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}

	bar.SetTotal(int64(len(*netfacilities)), false)
	for _, netfacility := range *netfacilities {
		// Put values in the database
		err = s.executeInsertOrUpdate(
			tx, (since == 0), table.Name, netfacility.ID, table.GetColumnsNames(), netfacility.Created,
			netfacility.Updated, netfacility.Status, netfacility.Name, netfacility.City, netfacility.Country,
			netfacility.LocalASN, netfacility.NetworkID, netfacility.FacilityID,
		)
		if err != nil {
			log.Fatal(err)
		}

		bar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, table.Name)

	tx.Commit()
	bar.SetTotal(-1, true)
}

func (s *Synchronization) SynchronizeNetworkInternetExchangeLANs(bar *mpb.Bar) {
	table := GetSchema().Tables["peeringdb_network_ixlan"]
	since := s.getLastSyncDate(table.Name)
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed network internet exchange LANs objects since the given
	// timestamp
	netixlans, err := s.API.GetNetworkInternetExchangeLAN(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*netixlans) < 1 {
		fmt.Printf("No network internet exchange LANs to sync since %s.\n",
			time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	tx, err := s.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}

	bar.SetTotal(int64(len(*netixlans)), false)
	for _, netixlan := range *netixlans {
		// Put values in the database
		err = s.executeInsertOrUpdate(
			tx, (since == 0), table.Name, netixlan.ID, table.GetColumnsNames(), netixlan.Created, netixlan.Updated,
			netixlan.Status, netixlan.Name, netixlan.Notes, netixlan.Speed, netixlan.ASN, netixlan.IPAddr4,
			netixlan.IPAddr6, netixlan.IsRSPeer, netixlan.BFDSupport, netixlan.Operational, netixlan.NetworkID,
			netixlan.InternetExchangeID, netixlan.InternetExchangeLANID, netixlan.NetworkSideID,
			netixlan.InternetExchangeSideID,
		)
		if err != nil {
			log.Fatal(err)
		}

		bar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, table.Name)

	tx.Commit()
	bar.SetTotal(-1, true)
}
