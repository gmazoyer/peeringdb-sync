package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/respawner/peeringdb"
)

const dbSchema = `
CREATE TABLE last_sync (
  id      integer  NOT NULL PRIMARY KEY AUTOINCREMENT,
  updated datetime NOT NULL
);
INSERT INTO last_sync VALUES (0, 0);

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
);

CREATE INDEX peeringdb_facility_org_id         ON peeringdb_facility         (org_id);
CREATE INDEX peeringdb_ix_org_id               ON peeringdb_ix               (org_id);
CREATE INDEX peeringdb_ix_facility_ix_id       ON peeringdb_ix_facility      (ix_id);
CREATE INDEX peeringdb_ix_facility_fac_id      ON peeringdb_ix_facility      (fac_id);
CREATE INDEX peeringdb_ixlan_ix_id             ON peeringdb_ixlan            (ix_id);
CREATE INDEX peeringdb_ixlan_prefix_ixlan_id   ON peeringdb_ixlan_prefix     (ixlan_id);
CREATE INDEX peeringdb_network_org_id          ON peeringdb_network          (org_id);
CREATE INDEX peeringdb_network_contact_net_id  ON peeringdb_network_contact  (net_id);
CREATE INDEX peeringdb_network_facility_net_id ON peeringdb_network_facility (net_id);
CREATE INDEX peeringdb_network_facility_fac_id ON peeringdb_network_facility (fac_id);
CREATE INDEX peeringdb_network_ixlan_ixlan_id  ON peeringdb_network_ixlan    (ixlan_id);
CREATE INDEX peeringdb_network_ixlan_net_id    ON peeringdb_network_ixlan    (net_id);`

func main() {
	databaseFile := "./peeringdb.db"
	initialSynchronization := true
	fullRequested := flag.Bool("full", false,
		"Request a full synchronization (needed to remove old entries)")
	flag.Parse()

	// The database already exists, assume it is not the first synchronization
	if _, err := os.Stat(databaseFile); err == nil {
		if *fullRequested {
			err = os.Remove(databaseFile)
			if err != nil {
				log.Fatal(err)
			} else {
				fmt.Println("Full synchronization requested, the local database has been removed.")
			}
		} else {
			initialSynchronization = false
		}
	}

	// Open the SQLite database, will create it if needed
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Prepare to query the API and the synchronization
	sync := synchronization{peeringdb.NewAPI(), db}

	var lastSync int64
	if !initialSynchronization {
		// Found the last database sync
		lastSync = sync.getLastSyncDate()
	} else {
		// First time, create the database schema
		_, err = db.Exec(dbSchema)
		if err != nil {
			log.Printf("%q: %s\n", err, dbSchema)
			return
		}
	}

	fmt.Println("Starting PeeringDB synchronization...")

	// Synchronize all objects
	sync.synchronizeOrganizations(lastSync)
	sync.synchronizeFacilities(lastSync)
	sync.synchronizeNetworks(lastSync)
	sync.synchronizeInternetExchanges(lastSync)
	sync.synchronizeInternetExchangeFacilities(lastSync)
	sync.synchronizeInternetLANs(lastSync)
	sync.synchronizeInternetPrefixes(lastSync)
	sync.synchronizeNetworkContacts(lastSync)
	sync.synchronizeNetworkFacilities(lastSync)
	sync.synchronizeNetworkInternetExchangeLANs(lastSync)

	// Set the last sync date
	sync.setLastSyncDate(time.Now().Unix())

	fmt.Println("PeeringDB fully synchronized.")
}
