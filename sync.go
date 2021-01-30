package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/respawner/peeringdb"
	"gopkg.in/cheggaaa/pb.v1"
)

// synchronization is a structure holding pointers to the PeeringDB API and
// database being used.
type synchronization struct {
	api *peeringdb.API
	db  *sql.DB
}

// idExistsInTable returns true of the given ID exists in the given database
// table. It returns false if the ID cannot be found.
func (s *synchronization) idExistsInTable(table string, id int) bool {
	// Look for the given ID in the given table
	result, err := s.db.Query(fmt.Sprintf("SELECT id FROM %s WHERE id = %d",
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
func (s *synchronization) insertOrUpdateStatement(forceInsert bool, table string, id int, columns []string) string {
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

func (s *synchronization) removeDeleted(tx *sql.Tx, table string) error {
	// Prepare the statement.
	result, err := tx.Query(fmt.Sprintf("DELETE FROM %s WHERE status == 'deleted'", table))
	if err != nil {
		return err
	}
	return result.Close()
}

// getLastSyncDate retrieves the timestamp at which the last synchronization
// has occured. The returned value is an int64.
func (s *synchronization) getLastSyncDate(table string) int64 {
	switch {
	case table == "":
		log.Fatal("table not supplied")
	case strings.Contains(table, " "):
		log.Fatal("table malformed")
	}

	updated := time.Unix(0, 0)

	// Query for last sync date
	result, err := s.db.Query(fmt.Sprintf("SELECT updated FROM %s ORDER BY updated DESC LIMIT 1", table))
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
func (s *synchronization) executeInsertOrUpdate(tx *sql.Tx, forceInsert bool, table string, id int, columns []string, values ...interface{}) error {
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

func (s *synchronization) synchronizeOrganizations() {
	since := s.getLastSyncDate("peeringdb_organization")
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed organizations objects since the given timestamp
	organizations, err := s.api.GetOrganization(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*organizations) < 1 {
		fmt.Printf("No organizations to sync since %s.\n", time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	tx, err := s.db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	columns := []string{"status", "created", "updated", "version", "address1",
		"address2", "city", "state", "zipcode", "country", "name", "website",
		"notes"}

	// Put values in the database
	progressbar := pb.StartNew(len(*organizations))
	for _, organization := range *organizations {
		err = s.executeInsertOrUpdate(tx, (since == 0),
			"peeringdb_organization", organization.ID, columns,
			organization.Status, organization.Created, organization.Updated, 0,
			organization.Address1, organization.Address2, organization.City,
			organization.State, organization.Zipcode, organization.Country,
			organization.Name, organization.Website, organization.Notes)
		if err != nil {
			log.Fatal(err)
		}

		progressbar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, "peeringdb_organization")

	tx.Commit()
	progressbar.FinishPrint("Organizations sync done.")
}

func (s *synchronization) synchronizeFacilities() {
	since := s.getLastSyncDate("peeringdb_facility")
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed facilities objects since the given timestamp
	facilities, err := s.api.GetFacility(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*facilities) < 1 {
		fmt.Printf("No facilities to sync since %s.\n", time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	tx, err := s.db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	columns := []string{"status", "created", "updated", "version", "address1",
		"address2", "city", "state", "zipcode", "country", "name", "website",
		"clli", "rencode", "npanxx", "notes", "org_id"}

	progressbar := pb.StartNew(len(*facilities))
	for _, facility := range *facilities {
		// Put values in the database
		err = s.executeInsertOrUpdate(tx, (since == 0), "peeringdb_facility",
			facility.ID, columns, facility.Status, facility.Created,
			facility.Updated, 0, facility.Address1, facility.Address2,
			facility.City, facility.State, facility.Zipcode, facility.Country,
			facility.Name, facility.Website, facility.CLLI, facility.Rencode,
			facility.Npanxx, facility.Notes, facility.OrganizationID)
		if err != nil {
			log.Fatal(err)
		}

		progressbar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, "peeringdb_facility")

	tx.Commit()
	progressbar.FinishPrint("Facilities sync done.")
}

func (s *synchronization) synchronizeNetworks() {
	since := s.getLastSyncDate("peeringdb_network")
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed networks objects since the given timestamp
	networks, err := s.api.GetNetwork(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*networks) < 1 {
		fmt.Printf("No networks to sync since %s.\n", time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	tx, err := s.db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	columns := []string{"status", "created", "updated", "version", "asn",
		"name", "aka", "irr_as_set", "website", "looking_glass",
		"route_server", "notes", "notes_private", "info_traffic", "info_ratio",
		"info_scope", "info_type", "info_prefixes4", "info_prefixes6",
		"info_unicast", "info_multicast", "info_ipv6", "policy_url",
		"policy_general", "policy_locations", "policy_ratio",
		"policy_contracts", "org_id"}

	progressbar := pb.StartNew(len(*networks))
	for _, network := range *networks {
		// Put values in the database
		err = s.executeInsertOrUpdate(tx, (since == 0), "peeringdb_network",
			network.ID, columns, network.Status, network.Created,
			network.Updated, 0, network.ASN, network.Name, network.AKA,
			network.IRRASSet, network.Website, network.LookingGlass,
			network.RouteServer, network.Notes, "", network.InfoTraffic,
			network.InfoRatio, network.InfoScope, network.InfoType,
			network.InfoPrefixes4, network.InfoPrefixes6, network.InfoUnicast,
			network.InfoMulticast, network.InfoIPv6, network.PolicyURL,
			network.PolicyGeneral, network.PolicyLocations,
			network.PolicyRatio, network.PolicyContracts,
			network.OrganizationID)
		if err != nil {
			log.Fatal(err)
		}

		progressbar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, "peeringdb_network")

	tx.Commit()
	progressbar.FinishPrint("Networks sync done.")
}

func (s *synchronization) synchronizeInternetExchanges() {
	since := s.getLastSyncDate("peeringdb_ix")
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed internet exchanges objects since the given timestamp
	ixs, err := s.api.GetInternetExchange(search)
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
	tx, err := s.db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	columns := []string{"status", "created", "updated", "version", "name",
		"name_long", "city", "country", "notes", "region_continent", "media",
		"proto_unicast", "proto_multicast", "proto_ipv6", "website",
		"url_stats", "tech_email", "tech_phone", "policy_email",
		"policy_phone", "org_id"}

	progressbar := pb.StartNew(len(*ixs))
	for _, ix := range *ixs {
		// Put values in the database
		err = s.executeInsertOrUpdate(tx, (since == 0), "peeringdb_ix",
			ix.ID, columns, ix.Status, ix.Created, ix.Updated, 0, ix.Name,
			ix.NameLong, ix.City, ix.Country, ix.Notes, ix.RegionContinent,
			ix.Media, ix.ProtoUnicast, ix.ProtoMulticast, ix.ProtoIPv6,
			ix.Website, ix.URLStats, ix.TechEmail, ix.TechPhone,
			ix.PolicyEmail, ix.PolicyPhone, ix.OrganizationID)
		if err != nil {
			log.Fatal(err)
		}

		progressbar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, "peeringdb_ix")

	tx.Commit()
	progressbar.FinishPrint("Internet exchanges sync done.")
}

func (s *synchronization) synchronizeInternetExchangeFacilities() {
	since := s.getLastSyncDate("peeringdb_ix_facility")
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed internet exchange facilities objects since the given
	// timestamp
	ixfacilities, err := s.api.GetInternetExchangeFacility(search)
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
	tx, err := s.db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	columns := []string{"status", "created", "updated", "version", "ix_id",
		"fac_id"}

	progressbar := pb.StartNew(len(*ixfacilities))
	for _, ixfacility := range *ixfacilities {
		// Put values in the database
		err = s.executeInsertOrUpdate(tx, (since == 0),
			"peeringdb_ix_facility", ixfacility.ID, columns, ixfacility.Status,
			ixfacility.Created, ixfacility.Updated, 0,
			ixfacility.InternetExchangeID, ixfacility.FacilityID)
		if err != nil {
			log.Fatal(err)
		}

		progressbar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, "peeringdb_ix_facility")

	tx.Commit()
	progressbar.FinishPrint("Internet exchange facilities sync done.")
}

func (s *synchronization) synchronizeInternetLANs() {
	since := s.getLastSyncDate("peeringdb_ixlan")
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed internet exchange LANs objects since the given timestamp
	ixlans, err := s.api.GetInternetExchangeLAN(search)
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
	tx, err := s.db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	columns := []string{"status", "created", "updated", "version", "name",
		"descr", "mtu", "vlan", "dot1q_support", "rs_asn", "arp_sponge",
		"ix_id"}

	progressbar := pb.StartNew(len(*ixlans))
	for _, ixlan := range *ixlans {
		// Put values in the database
		err = s.executeInsertOrUpdate(tx, (since == 0), "peeringdb_ixlan",
			ixlan.ID, columns, ixlan.Status, ixlan.Created, ixlan.Updated, 0,
			ixlan.Name, ixlan.Description, ixlan.MTU, 0, ixlan.Dot1QSupport,
			ixlan.RouteServerASN, ixlan.ARPSponge, ixlan.InternetExchangeID)
		if err != nil {
			log.Fatal(err)
		}

		progressbar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, "peeringdb_ixlan")

	tx.Commit()
	progressbar.FinishPrint("Internet exchange LANs sync done.")
}

func (s *synchronization) synchronizeInternetPrefixes() {
	since := s.getLastSyncDate("peeringdb_ixlan_prefix")
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed internet exchange prefixes objects since the given timestamp
	ixpfxs, err := s.api.GetInternetExchangePrefix(search)
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
	tx, err := s.db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	columns := []string{"status", "created", "updated", "version", "notes",
		"protocol", "prefix", "ixlan_id"}

	progressbar := pb.StartNew(len(*ixpfxs))
	for _, ixpfx := range *ixpfxs {
		// Put values in the database
		err = s.executeInsertOrUpdate(tx, (since == 0),
			"peeringdb_ixlan_prefix", ixpfx.ID, columns, ixpfx.Status,
			ixpfx.Created, ixpfx.Updated, 0, "", ixpfx.Protocol, ixpfx.Prefix,
			ixpfx.InternetExchangeLANID)
		if err != nil {
			log.Fatal(err)
		}

		progressbar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, "peeringdb_ixlan_prefix")

	tx.Commit()
	progressbar.FinishPrint("Internet exchange prefixes sync done.")
}

func (s *synchronization) synchronizeNetworkContacts() {
	since := s.getLastSyncDate("peeringdb_network_contact")
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed network contacts objects since the given timestamp
	netcontacts, err := s.api.GetNetworkContact(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*netcontacts) < 1 {
		fmt.Printf("No network contacts to sync since %s.\n",
			time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	tx, err := s.db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	columns := []string{"status", "created", "updated", "version", "role",
		"visible", "name", "phone", "email", "url", "net_id"}

	progressbar := pb.StartNew(len(*netcontacts))
	for _, netcontact := range *netcontacts {
		// Put values in the database
		err = s.executeInsertOrUpdate(tx, (since == 0),
			"peeringdb_network_contact", netcontact.ID, columns,
			netcontact.Status, netcontact.Created, netcontact.Updated, 0,
			netcontact.Role, netcontact.Visible, netcontact.Name,
			netcontact.Phone, netcontact.Email, netcontact.URL,
			netcontact.NetworkID)
		if err != nil {
			log.Fatal(err)
		}

		progressbar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, "peeringdb_network_contact")

	tx.Commit()
	progressbar.FinishPrint("Network contacts sync done.")
}

func (s *synchronization) synchronizeNetworkFacilities() {
	since := s.getLastSyncDate("peeringdb_network_facility")
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed network facilities objects since the given timestamp
	netfacilities, err := s.api.GetNetworkFacility(search)
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
	tx, err := s.db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	columns := []string{"status", "created", "updated", "version", "local_asn",
		"avail_sonet", "avail_ethernet", "avail_atm", "net_id", "fac_id"}

	progressbar := pb.StartNew(len(*netfacilities))
	for _, netfacility := range *netfacilities {
		// Put values in the database
		err = s.executeInsertOrUpdate(tx, (since == 0),
			"peeringdb_network_facility", netfacility.ID, columns,
			netfacility.Status, netfacility.Created, netfacility.Updated, 0,
			netfacility.LocalASN, 0, 0, 0, netfacility.NetworkID,
			netfacility.FacilityID)
		if err != nil {
			log.Fatal(err)
		}

		progressbar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, "peeringdb_network_facility")

	tx.Commit()
	progressbar.FinishPrint("Network facilities sync done.")
}

func (s *synchronization) synchronizeNetworkInternetExchangeLANs() {
	since := s.getLastSyncDate("peeringdb_network_ixlan")
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed network internet exchange LANs objects since the given
	// timestamp
	netixlans, err := s.api.GetNetworkInternetExchangeLAN(search)
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
	tx, err := s.db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	columns := []string{"status", "created", "updated", "version", "asn",
		"ipaddr4", "ipaddr6", "is_rs_peer", "notes", "speed", "net_id",
		"ixlan_id"}

	progressbar := pb.StartNew(len(*netixlans))
	for _, netixlan := range *netixlans {
		// Put values in the database
		err = s.executeInsertOrUpdate(tx, (since == 0),
			"peeringdb_network_ixlan", netixlan.ID, columns, netixlan.Status,
			netixlan.Created, netixlan.Updated, 0, netixlan.ASN,
			netixlan.IPAddr4, netixlan.IPAddr6, netixlan.IsRSPeer,
			netixlan.Notes, netixlan.Speed, netixlan.NetworkID,
			netixlan.InternetExchangeLANID)
		if err != nil {
			log.Fatal(err)
		}

		progressbar.Increment()
	}

	// Remove the entries marked as deleted.
	s.removeDeleted(tx, "peeringdb_network_ixlan")

	tx.Commit()
	progressbar.FinishPrint("Network internet exchange LANs sync done.")
}
