package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/respawner/peeringdb"
	"gopkg.in/cheggaaa/pb.v1"
)

const dbSchema = `
CREATE TABLE last_sync (
  id      integer  NOT NULL PRIMARY KEY AUTOINCREMENT,
  updated datetime NOT NULL
);
INSERT INTO last_sync VALUES (0, CURRENT_TIMESTAMP);

CREATE TABLE peeringdb_facility (
  id       integer      NOT NULL PRIMARY KEY AUTOINCREMENT,
  status   varchar(255) NOT NULL,
  created  datetime     NOT NULL,
  updated  datetime     NOT NULL,
  version  integer      NOT NULL,
  address1 varchar(255) NOT NULL,
  address2 varchar(255) NOT NULL,
  city     varchar(255) NOT NULL,
  state    varchar(255) NOT NULL,
  zipcode  varchar(48)  NOT NULL,
  country  varchar(2)   NOT NULL,
  name     varchar(255) NOT NULL UNIQUE,
  website  varchar(255) NOT NULL,
  clli     varchar(18)  NOT NULL,
  rencode  varchar(18)  NOT NULL,
  npanxx   varchar(21)  NOT NULL,
  notes    text         NOT NULL,
  org_id   integer      NOT NULL REFERENCES peeringdb_organization (id)
);

CREATE TABLE peeringdb_ix (
  id               integer      NOT NULL PRIMARY KEY AUTOINCREMENT,
  status           varchar(255) NOT NULL,
  created          datetime     NOT NULL,
  updated          datetime     NOT NULL,
  version          integer      NOT NULL,
  name             varchar(64)  NOT NULL UNIQUE,
  name_long        varchar(254) NOT NULL,
  city             varchar(192) NOT NULL,
  country          varchar(2)   NOT NULL,
  notes            text         NOT NULL,
  region_continent varchar(255) NOT NULL,
  media            varchar(128) NOT NULL,
  proto_unicast    bool         NOT NULL,
  proto_multicast  bool         NOT NULL,
  proto_ipv6       bool         NOT NULL,
  website          varchar(255) NOT NULL,
  url_stats        varchar(255) NOT NULL,
  tech_email       varchar(254) NOT NULL,
  tech_phone       varchar(192) NOT NULL,
  policy_email     varchar(254) NOT NULL,
  policy_phone     varchar(192) NOT NULL,
  org_id           integer      NOT NULL REFERENCES peeringdb_organization (id)
);

CREATE TABLE peeringdb_ix_facility (
  id      integer      NOT NULL PRIMARY KEY AUTOINCREMENT,
  status  varchar(255) NOT NULL,
  created datetime     NOT NULL,
  updated datetime     NOT NULL,
  version integer      NOT NULL,
  ix_id   integer      NOT NULL REFERENCES peeringdb_ix       (id),
  fac_id  integer      NOT NULL REFERENCES peeringdb_facility (id),
  UNIQUE (ix_id, fac_id)
);

CREATE TABLE peeringdb_ixlan (
  id            integer      NOT NULL PRIMARY KEY AUTOINCREMENT,
  status        varchar(255) NOT NULL,
  created       datetime     NOT NULL,
  updated       datetime     NOT NULL,
  version       integer      NOT NULL,
  name          varchar(255) NOT NULL,
  descr         text         NOT NULL,
  mtu           integer unsigned NULL,
  vlan          integer unsigned NULL,
  dot1q_support bool         NOT NULL,
  rs_asn        integer unsigned NULL,
  arp_sponge    varchar(17)      NULL,
  ix_id         integer      NOT NULL REFERENCES peeringdb_ix (id)
);

CREATE TABLE peeringdb_ixlan_prefix (
  id       integer      NOT NULL PRIMARY KEY AUTOINCREMENT,
  status   varchar(255) NOT NULL,
  created  datetime     NOT NULL,
  updated  datetime     NOT NULL,
  version  integer      NOT NULL,
  notes    varchar(255) NOT NULL,
  protocol varchar(64)  NOT NULL,
  prefix   varchar(43)  NOT NULL UNIQUE,
  ixlan_id integer      NOT NULL REFERENCES peeringdb_ixlan (id)
);

CREATE TABLE peeringdb_network (
  id               integer          NOT NULL PRIMARY KEY AUTOINCREMENT,
  status           varchar(255)     NOT NULL,
  created          datetime         NOT NULL,
  updated          datetime         NOT NULL,
  version          integer          NOT NULL,
  asn              integer unsigned NOT NULL UNIQUE,
  name             varchar(255)     NOT NULL UNIQUE,
  aka              varchar(255)     NOT NULL,
  irr_as_set       varchar(255)     NOT NULL,
  website          varchar(255)     NOT NULL,
  looking_glass    varchar(255)     NOT NULL,
  route_server     varchar(255)     NOT NULL,
  notes            text             NOT NULL,
  notes_private    text             NOT NULL,
  info_traffic     varchar(39)      NOT NULL,
  info_ratio       varchar(45)      NOT NULL,
  info_scope       varchar(39)      NOT NULL,
  info_type        varchar(60)      NOT NULL,
  info_prefixes4   integer unsigned     NULL,
  info_prefixes6   integer unsigned     NULL,
  info_unicast     bool             NOT NULL,
  info_multicast   bool             NOT NULL,
  info_ipv6        bool             NOT NULL,
  policy_url       varchar(255)     NOT NULL,
  policy_general   varchar(72)      NOT NULL,
  policy_locations varchar(72)      NOT NULL,
  policy_ratio     bool             NOT NULL,
  policy_contracts varchar(36)      NOT NULL,
  org_id           integer          NOT NULL REFERENCES peeringdb_organization (id)
);

CREATE TABLE peeringdb_network_contact (
  id      integer      NOT NULL PRIMARY KEY AUTOINCREMENT,
  status  varchar(255) NOT NULL,
  created datetime     NOT NULL,
  updated datetime     NOT NULL,
  version integer      NOT NULL,
  role    varchar(27)  NOT NULL,
  visible varchar(64)  NOT NULL,
  name    varchar(254) NOT NULL,
  phone   varchar(100) NOT NULL,
  email   varchar(254) NOT NULL,
  url     varchar(255) NOT NULL,
  net_id  integer      NOT NULL REFERENCES peeringdb_network (id)
);

CREATE TABLE peeringdb_network_facility (
  id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  status         varchar(255) NOT NULL,
  created        datetime     NOT NULL,
  updated        datetime     NOT NULL,
  version        integer      NOT NULL,
  local_asn      integer unsigned NULL,
  avail_sonet    bool         NOT NULL,
  avail_ethernet bool         NOT NULL,
  avail_atm      bool         NOT NULL,
  net_id         integer      NOT NULL REFERENCES peeringdb_network  (id),
  fac_id         integer      NOT NULL REFERENCES peeringdb_facility (id),
  UNIQUE (net_id, fac_id, local_asn)
);

CREATE TABLE peeringdb_network_ixlan (
  id         integer          NOT NULL PRIMARY KEY AUTOINCREMENT,
  status     varchar(255)     NOT NULL,
  created    datetime         NOT NULL,
  updated    datetime         NOT NULL,
  version    integer          NOT NULL,
  asn        integer unsigned NOT NULL,
  ipaddr4    varchar(39)          NULL,
  ipaddr6    varchar(39)          NULL,
  is_rs_peer bool             NOT NULL,
  notes      varchar(255)     NOT NULL,
  speed      integer unsigned NOT NULL,
  net_id     integer          NOT NULL REFERENCES peeringdb_network (id),
  ixlan_id   integer          NOT NULL REFERENCES peeringdb_ixlan   (id)
);

CREATE TABLE peeringdb_organization (
  id       integer      NOT NULL PRIMARY KEY AUTOINCREMENT,
  status   varchar(255) NOT NULL,
  created  datetime     NOT NULL,
  updated  datetime     NOT NULL,
  version  integer      NOT NULL,
  address1 varchar(255) NOT NULL,
  address2 varchar(255) NOT NULL,
  city     varchar(255) NOT NULL,
  state    varchar(255) NOT NULL,
  zipcode  varchar(48)  NOT NULL,
  country  varchar(2)   NOT NULL,
  name     varchar(255) NOT NULL UNIQUE,
  website  varchar(255) NOT NULL,
  notes    text         NOT NULL
);`

func synchronizeOrganizations(api *peeringdb.API, since int64, db *sql.DB) {
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed organizations objects since the given timestamp
	organizations, err := api.GetOrganization(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*organizations) < 1 {
		fmt.Printf("No organizations to sync since %s.\n", time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	insert, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	progressbar := pb.StartNew(len(*organizations))
	for _, organization := range *organizations {
		// Prepare the database insertion
		statement, err := insert.Prepare("INSERT INTO peeringdb_organization VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			log.Fatal(err)
		}
		defer statement.Close()

		// Insert the values
		_, err = statement.Exec(organization.ID, organization.Status,
			organization.Created, organization.Updated, 0,
			organization.Address1, organization.Address2, organization.City,
			organization.State, organization.Zipcode, organization.Country,
			organization.Name, organization.Website, organization.Notes)
		if err != nil {
			log.Fatal(err)
		}

		progressbar.Increment()
		time.Sleep(time.Millisecond)
	}

	insert.Commit()
	progressbar.FinishPrint("Syncing organizations done.")
}

func synchronizeFacilities(api *peeringdb.API, since int64, db *sql.DB) {
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed facilities objects since the given timestamp
	facilities, err := api.GetFacility(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*facilities) < 1 {
		fmt.Printf("No facilities to sync since %s.\n", time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	insert, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	progressbar := pb.StartNew(len(*facilities))
	for _, facility := range *facilities {
		// Prepare the database insertion
		statement, err := insert.Prepare("INSERT INTO peeringdb_facility VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			fmt.Println("========================")
			log.Fatal(err)
		}
		defer statement.Close()

		// Insert the values
		_, err = statement.Exec(facility.ID, facility.Status, facility.Created,
			facility.Updated, 0, facility.Address1, facility.Address2,
			facility.City, facility.State, facility.Zipcode, facility.Country,
			facility.Name, facility.Website, facility.CLLI, facility.Rencode,
			facility.Npanxx, facility.Notes, facility.OrganizationID)
		if err != nil {
			fmt.Println("---------------------------")
			log.Fatal(err)
		}

		progressbar.Increment()
		time.Sleep(time.Millisecond)
	}

	insert.Commit()
	progressbar.FinishPrint("Syncing facilities done.")
}

func synchronizeNetworks(api *peeringdb.API, since int64, db *sql.DB) {
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed networks objects since the given timestamp
	networks, err := api.GetNetwork(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*networks) < 1 {
		fmt.Printf("No networks to sync since %s.\n", time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	insert, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	progressbar := pb.StartNew(len(*networks))
	for _, network := range *networks {
		// Prepare the database insertion
		statement, err := insert.Prepare("INSERT INTO peeringdb_network VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			log.Fatal(err)
		}
		defer statement.Close()

		// Insert the values
		_, err = statement.Exec(network.ID, network.Status, network.Created,
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
		time.Sleep(time.Millisecond)
	}

	insert.Commit()
	progressbar.FinishPrint("Syncing networks done.")
}

func synchronizeInternetExchanges(api *peeringdb.API, since int64, db *sql.DB) {
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed internet exchanges objects since the given timestamp
	ixs, err := api.GetInternetExchange(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*ixs) < 1 {
		fmt.Printf("No internet exchanges to sync since %s.\n", time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	insert, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	progressbar := pb.StartNew(len(*ixs))
	for _, ix := range *ixs {
		// Prepare the database insertion
		statement, err := insert.Prepare("INSERT INTO peeringdb_ix VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			log.Fatal(err)
		}
		defer statement.Close()

		// Insert the values
		_, err = statement.Exec(ix.ID, ix.Status, ix.Created, ix.Updated, 0,
			ix.Name, ix.NameLong, ix.City, ix.Country, ix.Notes,
			ix.RegionContinent, ix.Media, ix.ProtoUnicast, ix.ProtoMulticast,
			ix.ProtoIPv6, ix.Website, ix.URLStats, ix.TechEmail, ix.TechPhone,
			ix.PolicyEmail, ix.PolicyPhone, ix.OrganizationID)
		if err != nil {
			log.Fatal(err)
		}

		progressbar.Increment()
		time.Sleep(time.Millisecond)
	}

	insert.Commit()
	progressbar.FinishPrint("Syncing internet exchanges done.")
}

func synchronizeInternetExchangeFacilities(api *peeringdb.API, since int64, db *sql.DB) {
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed internet exchange facilities objects since the given timestamp
	ixfacilities, err := api.GetInternetExchangeFacility(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*ixfacilities) < 1 {
		fmt.Printf("No internet exchange facilities to sync since %s.\n", time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	insert, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	progressbar := pb.StartNew(len(*ixfacilities))
	for _, ixfacility := range *ixfacilities {
		// Prepare the database insertion
		statement, err := insert.Prepare("INSERT INTO peeringdb_ix_facility VALUES (?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			log.Fatal(err)
		}
		defer statement.Close()

		// Insert the values
		_, err = statement.Exec(ixfacility.ID, ixfacility.Status,
			ixfacility.Created, ixfacility.Updated, 0,
			ixfacility.InternetExchangeID, ixfacility.FacilityID)
		if err != nil {
			log.Fatal(err)
		}

		progressbar.Increment()
		time.Sleep(time.Millisecond)
	}

	insert.Commit()
	progressbar.FinishPrint("Syncing internet exchange facilities done.")
}

func synchronizeInternetLANs(api *peeringdb.API, since int64, db *sql.DB) {
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed internet exchange LANs objects since the given timestamp
	ixlans, err := api.GetInternetExchangeLAN(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*ixlans) < 1 {
		fmt.Printf("No internet exchange LANs to sync since %s.\n", time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	insert, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	progressbar := pb.StartNew(len(*ixlans))
	for _, ixlan := range *ixlans {
		// Prepare the database insertion
		statement, err := insert.Prepare("INSERT INTO peeringdb_ixlan VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			log.Fatal(err)
		}
		defer statement.Close()

		// Insert the values
		_, err = statement.Exec(ixlan.ID, ixlan.Status, ixlan.Created,
			ixlan.Updated, 0, ixlan.Name, ixlan.Description, ixlan.MTU, 0,
			ixlan.Dot1QSupport, ixlan.RouteServerASN, ixlan.ARPSponge,
			ixlan.InternetExchangeID)
		if err != nil {
			log.Fatal(err)
		}

		progressbar.Increment()
		time.Sleep(time.Millisecond)
	}

	insert.Commit()
	progressbar.FinishPrint("Syncing internet exchange LANs done.")
}

func synchronizeInternetPrefixes(api *peeringdb.API, since int64, db *sql.DB) {
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed internet exchange prefixes objects since the given timestamp
	ixpfxs, err := api.GetInternetExchangePrefix(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*ixpfxs) < 1 {
		fmt.Printf("No internet exchange prefixes to sync since %s.\n", time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	insert, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	progressbar := pb.StartNew(len(*ixpfxs))
	for _, ixpfx := range *ixpfxs {
		// Prepare the database insertion
		statement, err := insert.Prepare("INSERT INTO peeringdb_ixlan_prefix VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			log.Fatal(err)
		}
		defer statement.Close()

		// Insert the values
		_, err = statement.Exec(ixpfx.ID, ixpfx.Status, ixpfx.Created,
			ixpfx.Updated, 0, "", ixpfx.Protocol, ixpfx.Prefix,
			ixpfx.InternetExchangeLANID)
		if err != nil {
			log.Fatal(err)
		}

		progressbar.Increment()
		time.Sleep(time.Millisecond)
	}

	insert.Commit()
	progressbar.FinishPrint("Syncing internet exchange prefixes done.")
}

func synchronizeNetworkContacts(api *peeringdb.API, since int64, db *sql.DB) {
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed network contacts objects since the given timestamp
	netcontacts, err := api.GetNetworkContact(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*netcontacts) < 1 {
		fmt.Printf("No network contacts to sync since %s.\n", time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	insert, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	progressbar := pb.StartNew(len(*netcontacts))
	for _, netcontact := range *netcontacts {
		// Prepare the database insertion
		statement, err := insert.Prepare("INSERT INTO peeringdb_network_contact VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			log.Fatal(err)
		}
		defer statement.Close()

		// Insert the values
		_, err = statement.Exec(netcontact.ID, netcontact.Status,
			netcontact.Created, netcontact.Updated, 0, netcontact.Role,
			netcontact.Visible, netcontact.Name, netcontact.Phone,
			netcontact.Email, netcontact.URL, netcontact.NetworkID)
		if err != nil {
			log.Fatal(err)
		}

		progressbar.Increment()
		time.Sleep(time.Millisecond)
	}

	insert.Commit()
	progressbar.FinishPrint("Syncing network contacts done.")
}

func synchronizeNetworkFacilities(api *peeringdb.API, since int64, db *sql.DB) {
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed network facilities objects since the given timestamp
	netfacilities, err := api.GetNetworkFacility(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*netfacilities) < 1 {
		fmt.Printf("No network facilities to sync since %s.\n", time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	insert, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	progressbar := pb.StartNew(len(*netfacilities))
	for _, netfacility := range *netfacilities {
		// Prepare the database insertion
		statement, err := insert.Prepare("INSERT INTO peeringdb_network_facility VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			log.Fatal(err)
		}
		defer statement.Close()

		// Insert the values
		_, err = statement.Exec(netfacility.ID, netfacility.Status,
			netfacility.Created, netfacility.Updated, 0, netfacility.LocalASN,
			0, 0, 0, netfacility.NetworkID, netfacility.FacilityID)
		if err != nil {
			log.Fatal(err)
		}

		progressbar.Increment()
		time.Sleep(time.Millisecond)
	}

	insert.Commit()
	progressbar.FinishPrint("Syncing network facilities done.")
}

func synchronizeNetworkInternetExchangeLANs(api *peeringdb.API, since int64, db *sql.DB) {
	search := make(map[string]interface{})
	search["since"] = since

	// Get changed network internet exchange LANs objects since the given
	// timestamp
	netixlans, err := api.GetNetworkInternetExchangeLAN(search)
	if err != nil {
		log.Fatal(err)
	}

	// Slice is empty, nothing to sync
	if len(*netixlans) < 1 {
		fmt.Printf("No network internet exchange LANs to sync since %s.\n", time.Unix(since, 0))
		return
	}

	// Start to work on the local database
	insert, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	progressbar := pb.StartNew(len(*netixlans))
	for _, netixlan := range *netixlans {
		// Prepare the database insertion
		statement, err := insert.Prepare("INSERT INTO peeringdb_network_ixlan VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			log.Fatal(err)
		}
		defer statement.Close()

		// Insert the values
		_, err = statement.Exec(netixlan.ID, netixlan.Status, netixlan.Created,
			netixlan.Updated, 0, netixlan.ASN, netixlan.IPAddr4,
			netixlan.IPAddr6, netixlan.IsRSPeer, netixlan.Notes,
			netixlan.Speed, netixlan.NetworkID, netixlan.InternetExchangeLANID)
		if err != nil {
			log.Fatal(err)
		}

		progressbar.Increment()
		time.Sleep(time.Millisecond)
	}

	insert.Commit()
	progressbar.FinishPrint("Syncing network internet exchange LANs done.")
}

// getLastSyncDate retrieves the timestamp at which the last synchronization
// has occured. The returned value is an int64.
func getLastSyncDate(db *sql.DB) int64 {
	var updated time.Time

	// Query for last sync date
	result, err := db.Query("SELECT updated FROM last_sync WHERE id = 0")
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

// setLastSyncDate sets the timestamp at which the last synchronization has
// been done.
func setLastSyncDate(db *sql.DB, updated int64) {
	update, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// Prepare the database updated
	statement, err := update.Prepare("UPDATE last_sync SET updated = ? WHERE id = 0")
	if err != nil {
		log.Fatal(err)
	}
	defer statement.Close()

	// Set the values
	_, err = statement.Exec(updated)
	if err != nil {
		log.Fatal(err)
	}

	update.Commit()
}

func main() {
	databaseFile := "./peeringdb.db"
	initialSynchronization := true

	// The database already exists, assume it is not the first synchronization
	if _, err := os.Stat(databaseFile); err == nil {
		initialSynchronization = false
	}

	fmt.Println("Starting PeeringDB synchronization...")

	// Open the SQLite database, will create it if needed
	db, err := sql.Open("sqlite3", "./peeringdb.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var lastSync int64
	if !initialSynchronization {
		// Found the last database sync
		lastSync = getLastSyncDate(db)
	} else {
		// First time, create the database schema
		_, err = db.Exec(dbSchema)
		if err != nil {
			log.Printf("%q: %s\n", err, dbSchema)
			return
		}
	}

	// Prepare to query the API
	api := peeringdb.NewAPI()

	// Synchronize all objects
	synchronizeOrganizations(api, lastSync, db)
	synchronizeFacilities(api, lastSync, db)
	synchronizeNetworks(api, lastSync, db)
	synchronizeInternetExchanges(api, lastSync, db)
	synchronizeInternetExchangeFacilities(api, lastSync, db)
	synchronizeInternetLANs(api, lastSync, db)
	synchronizeInternetPrefixes(api, lastSync, db)
	synchronizeNetworkContacts(api, lastSync, db)
	synchronizeNetworkFacilities(api, lastSync, db)
	synchronizeNetworkInternetExchangeLANs(api, lastSync, db)

	// Set the last sync date
	setLastSyncDate(db, time.Now().Unix())

	fmt.Println("PeeringDB fully synchronized.")
}