package database

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func GetSchema() *Schema {
	return &Schema{
		Tables: map[string]Table{
			"peeringdb_organization": {
				Name: "peeringdb_organization",
				Columns: []Column{
					{Name: "id", Type: "integer", Constraints: "NOT NULL PRIMARY KEY AUTOINCREMENT"},
					{Name: "created", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "updated", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "status", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "name", Type: "varchar(255)", Constraints: "NOT NULL UNIQUE"},
					{Name: "aka", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "name_long", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "website", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "social_media", Type: "text", Constraints: "NOT NULL"},
					{Name: "notes", Type: "text", Constraints: "NOT NULL"},
					{Name: "address1", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "address2", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "city", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "country", Type: "varchar(7)", Constraints: "NOT NULL"},
					{Name: "state", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "zipcode", Type: "varchar(48)", Constraints: "NOT NULL"},
					{Name: "floor", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "suite", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "latitude", Type: "float", Constraints: "NULL"},
					{Name: "longitude", Type: "float", Constraints: "NULL"},
				},
			},
			"peeringdb_campus": {
				Name: "peeringdb_campus",
				Columns: []Column{
					{Name: "id", Type: "integer", Constraints: "NOT NULL PRIMARY KEY AUTOINCREMENT"},
					{Name: "created", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "updated", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "status", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "name", Type: "varchar(255)", Constraints: "NOT NULL UNIQUE"},
					{Name: "name_long", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "aka", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "website", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "social_media", Type: "text", Constraints: "NOT NULL"},
					{Name: "notes", Type: "text", Constraints: "NOT NULL"},
					{Name: "city", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "country", Type: "varchar(7)", Constraints: "NOT NULL"},
					{Name: "state", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "zipcode", Type: "varchar(48)", Constraints: "NOT NULL"},
					{Name: "org_id", Type: "integer", Constraints: "NOT NULL REFERENCES peeringdb_organization (id)"},
				},
			},
			"peeringdb_facility": {
				Name: "peeringdb_facility",
				Columns: []Column{
					{Name: "id", Type: "integer", Constraints: "NOT NULL PRIMARY KEY AUTOINCREMENT"},
					{Name: "created", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "updated", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "status", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "name", Type: "varchar(255)", Constraints: "NOT NULL UNIQUE"},
					{Name: "aka", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "name_long", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "website", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "social_media", Type: "text", Constraints: "NOT NULL"},
					{Name: "clli", Type: "varchar(18)", Constraints: "NOT NULL"},
					{Name: "rencode", Type: "varchar(18)", Constraints: "NOT NULL"},
					{Name: "npanxx", Type: "varchar(21)", Constraints: "NOT NULL"},
					{Name: "notes", Type: "text", Constraints: "NOT NULL"},
					{Name: "sales_email", Type: "varchar(254)", Constraints: "NOT NULL"},
					{Name: "sales_phone", Type: "varchar(192)", Constraints: "NOT NULL"},
					{Name: "tech_email", Type: "varchar(254)", Constraints: "NOT NULL"},
					{Name: "tech_phone", Type: "varchar(192)", Constraints: "NOT NULL"},
					{Name: "available_voltage_services", Type: "varchar(255)", Constraints: "NULL"},
					{Name: "diverse_serving_substations", Type: "bool", Constraints: "NULL"},
					{Name: "property", Type: "varchar(27)", Constraints: "NULL"},
					{Name: "region_continent", Type: "varchar(255)", Constraints: "NULL"},
					{Name: "status_dashboard", Type: "varchar(255)", Constraints: "NULL"},
					{Name: "address1", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "address2", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "city", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "country", Type: "varchar(7)", Constraints: "NOT NULL"},
					{Name: "state", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "zipcode", Type: "varchar(48)", Constraints: "NOT NULL"},
					{Name: "floor", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "suite", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "latitude", Type: "float", Constraints: "NULL"},
					{Name: "longitude", Type: "float", Constraints: "NULL"},
					{Name: "org_id", Type: "integer", Constraints: "NOT NULL REFERENCES peeringdb_organization (id)"},
					{Name: "campus_id", Type: "integer", Constraints: "NULL REFERENCES peeringdb_campus (id)"},
				},
			},
			"peeringdb_carrier": {
				Name: "peeringdb_carrier",
				Columns: []Column{
					{Name: "id", Type: "integer", Constraints: "NOT NULL PRIMARY KEY AUTOINCREMENT"},
					{Name: "created", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "updated", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "status", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "name", Type: "varchar(255)", Constraints: "NOT NULL UNIQUE"},
					{Name: "aka", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "name_long", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "website", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "social_media", Type: "text", Constraints: "NULL"},
					{Name: "notes", Type: "text", Constraints: "NOT NULL"},
					{Name: "org_id", Type: "integer", Constraints: "NOT NULL REFERENCES peeringdb_organization (id)"},
				},
			},
			"peeringdb_carrier_facility": {
				Name: "peeringdb_carrier_facility",
				Columns: []Column{
					{Name: "id", Type: "integer", Constraints: "NOT NULL PRIMARY KEY AUTOINCREMENT"},
					{Name: "carrier_id", Type: "integer", Constraints: "NOT NULL REFERENCES peeringdb_carrier (id)"},
					{Name: "fac_id", Type: "integer", Constraints: "NOT NULL REFERENCES peeringdb_facility (id)"},
				},
				UniquenessConstraints: []string{"carrier_id", "fac_id"},
			},
			"peeringdb_network": {
				Name: "peeringdb_network",
				Columns: []Column{
					{Name: "id", Type: "integer", Constraints: "NOT NULL PRIMARY KEY AUTOINCREMENT"},
					{Name: "created", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "updated", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "status", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "name", Type: "varchar(255)", Constraints: "NOT NULL UNIQUE"},
					{Name: "aka", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "name_long", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "website", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "social_media", Type: "text", Constraints: "NULL"},
					{Name: "asn", Type: "integer unsigned", Constraints: "NOT NULL UNIQUE"},
					{Name: "looking_glass", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "route_server", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "irr_as_set", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "info_type", Type: "varchar(60)", Constraints: "NOT NULL"},
					{Name: "info_types", Type: "varchar(255)", Constraints: "NULL"},
					{Name: "info_prefixes4", Type: "integer unsigned", Constraints: "NULL"},
					{Name: "info_prefixes6", Type: "integer unsigned", Constraints: "NULL"},
					{Name: "info_traffic", Type: "varchar(39)", Constraints: "NOT NULL"},
					{Name: "info_ratio", Type: "varchar(45)", Constraints: "NOT NULL"},
					{Name: "info_scope", Type: "varchar(39)", Constraints: "NOT NULL"},
					{Name: "info_unicast", Type: "bool", Constraints: "NOT NULL"},
					{Name: "info_multicast", Type: "bool", Constraints: "NOT NULL"},
					{Name: "info_ipv6", Type: "bool", Constraints: "NOT NULL"},
					{Name: "info_never_via_route_servers", Type: "bool", Constraints: "NOT NULL"},
					{Name: "notes", Type: "text", Constraints: "NOT NULL"},
					{Name: "policy_url", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "policy_general", Type: "varchar(72)", Constraints: "NOT NULL"},
					{Name: "policy_locations", Type: "varchar(72)", Constraints: "NOT NULL"},
					{Name: "policy_ratio", Type: "bool", Constraints: "NOT NULL"},
					{Name: "policy_contracts", Type: "varchar(36)", Constraints: "NOT NULL"},
					{Name: "allow_ixp_update", Type: "bool", Constraints: "NOT NULL"},
					{Name: "status_dashboard", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "rir_status", Type: "varchar(255)", Constraints: "NULL"},
					{Name: "rir_status_updated", Type: "datetime", Constraints: "NULL"},
					{Name: "org_id", Type: "integer", Constraints: "NOT NULL REFERENCES peeringdb_organization (id)"},
				},
			},
			"peeringdb_ix": {
				Name: "peeringdb_ix",
				Columns: []Column{
					{Name: "id", Type: "integer", Constraints: "NOT NULL PRIMARY KEY AUTOINCREMENT"},
					{Name: "created", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "updated", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "status", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "name", Type: "varchar(64)", Constraints: "NOT NULL UNIQUE"},
					{Name: "aka", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "name_long", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "city", Type: "varchar(192)", Constraints: "NOT NULL"},
					{Name: "country", Type: "varchar(7)", Constraints: "NOT NULL"},
					{Name: "region_continent", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "media", Type: "varchar(128)", Constraints: "NOT NULL"},
					{Name: "notes", Type: "text", Constraints: "NULL"},
					{Name: "proto_unicast", Type: "bool", Constraints: "NOT NULL"},
					{Name: "proto_multicast", Type: "bool", Constraints: "NOT NULL"},
					{Name: "proto_ipv6", Type: "bool", Constraints: "NOT NULL"},
					{Name: "website", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "social_media", Type: "text", Constraints: "NULL"},
					{Name: "url_stats", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "tech_email", Type: "varchar(254)", Constraints: "NOT NULL"},
					{Name: "tech_phone", Type: "varchar(192)", Constraints: "NOT NULL"},
					{Name: "policy_email", Type: "varchar(254)", Constraints: "NOT NULL"},
					{Name: "policy_phone", Type: "varchar(192)", Constraints: "NOT NULL"},
					{Name: "sales_email", Type: "varchar(254)", Constraints: "NOT NULL"},
					{Name: "sales_phone", Type: "varchar(192)", Constraints: "NOT NULL"},
					{Name: "ixf_net_count", Type: "integer unsigned", Constraints: "NOT NULL"},
					{Name: "ixf_last_import", Type: "datetime", Constraints: "NULL"},
					{Name: "ixf_import_request", Type: "datetime", Constraints: "NULL"},
					{Name: "ixf_import_request_status", Type: "varchar(255)", Constraints: "NULL"},
					{Name: "service_level", Type: "varchar(60)", Constraints: "NOT NULL"},
					{Name: "terms", Type: "varchar(60)", Constraints: "NOT NULL"},
					{Name: "status_dashboard", Type: "varchar(255)", Constraints: "NULL"},
					{Name: "org_id", Type: "integer", Constraints: "NOT NULL REFERENCES peeringdb_organization (id)"},
				},
			},
			"peeringdb_ix_facility": {
				Name: "peeringdb_ix_facility",
				Columns: []Column{
					{Name: "id", Type: "integer", Constraints: "NOT NULL PRIMARY KEY AUTOINCREMENT"},
					{Name: "created", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "updated", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "status", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "name", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "city", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "country", Type: "varchar(7)", Constraints: "NOT NULL"},
					{Name: "ix_id", Type: "integer", Constraints: "NOT NULL REFERENCES peeringdb_ix (id)"},
					{Name: "fac_id", Type: "integer", Constraints: "NOT NULL REFERENCES peeringdb_facility (id)"},
				},
				UniquenessConstraints: []string{"ix_id", "fac_id"},
			},
			"peeringdb_ixlan": {
				Name: "peeringdb_ixlan",
				Columns: []Column{
					{Name: "id", Type: "integer", Constraints: "NOT NULL PRIMARY KEY AUTOINCREMENT"},
					{Name: "created", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "updated", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "status", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "name", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "descr", Type: "text", Constraints: "NOT NULL"},
					{Name: "mtu", Type: "integer unsigned", Constraints: "NULL"},
					{Name: "dot1q_support", Type: "bool", Constraints: "NOT NULL"},
					{Name: "rs_asn", Type: "integer unsigned", Constraints: "NULL"},
					{Name: "arp_sponge", Type: "varchar(17)", Constraints: "NULL"},
					{Name: "ixf_ixp_member_list_url", Type: "varchar(255)", Constraints: "NULL"},
					{Name: "ixf_ixp_member_list_url_visible", Type: "bool", Constraints: "NULL"},
					{Name: "ixf_ixp_import_enabled", Type: "bool", Constraints: "NULL"},
					{Name: "ix_id", Type: "integer", Constraints: "NOT NULL REFERENCES peeringdb_ix (id)"},
				},
			},
			"peeringdb_ix_prefix": {
				Name: "peeringdb_ix_prefix",
				Columns: []Column{
					{Name: "id", Type: "integer", Constraints: "NOT NULL PRIMARY KEY AUTOINCREMENT"},
					{Name: "created", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "updated", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "status", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "protocol", Type: "varchar(64)", Constraints: "NOT NULL"},
					{Name: "prefix", Type: "varchar(43)", Constraints: "NOT NULL UNIQUE"},
					{Name: "in_dfz", Type: "bool", Constraints: "NOT NULL"},
					{Name: "ixlan_id", Type: "integer", Constraints: "NOT NULL REFERENCES peeringdb_ixlan (id)"},
				},
			},
			"peeringdb_network_contact": {
				Name: "peeringdb_network_contact",
				Columns: []Column{
					{Name: "id", Type: "integer", Constraints: "NOT NULL PRIMARY KEY AUTOINCREMENT"},
					{Name: "created", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "updated", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "status", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "role", Type: "varchar(27)", Constraints: "NOT NULL"},
					{Name: "visible", Type: "varchar(64)", Constraints: "NOT NULL"},
					{Name: "name", Type: "varchar(254)", Constraints: "NOT NULL"},
					{Name: "phone", Type: "varchar(100)", Constraints: "NOT NULL"},
					{Name: "email", Type: "varchar(254)", Constraints: "NOT NULL"},
					{Name: "url", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "net_id", Type: "integer", Constraints: "NOT NULL REFERENCES peeringdb_network (id)"},
				},
			},
			"peeringdb_network_facility": {
				Name: "peeringdb_network_facility",
				Columns: []Column{
					{Name: "id", Type: "integer", Constraints: "NOT NULL PRIMARY KEY AUTOINCREMENT"},
					{Name: "created", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "updated", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "status", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "name", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "city", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "country", Type: "varchar(7)", Constraints: "NOT NULL"},
					{Name: "local_asn", Type: "integer unsigned", Constraints: "NULL"},
					{Name: "net_id", Type: "integer", Constraints: "NOT NULL REFERENCES peeringdb_network (id)"},
					{Name: "fac_id", Type: "integer", Constraints: "NOT NULL REFERENCES peeringdb_facility (id)"},
				},
				UniquenessConstraints: []string{"net_id", "fac_id", "local_asn"},
			},
			"peeringdb_network_ixlan": {
				Name: "peeringdb_network_ixlan",
				Columns: []Column{
					{Name: "id", Type: "integer", Constraints: "NOT NULL PRIMARY KEY AUTOINCREMENT"},
					{Name: "created", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "updated", Type: "datetime", Constraints: "NOT NULL"},
					{Name: "status", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "name", Type: "varchar(255)", Constraints: "NOT NULL"},
					{Name: "notes", Type: "text", Constraints: "NULL"},
					{Name: "speed", Type: "integer unsigned", Constraints: "NOT NULL"},
					{Name: "asn", Type: "integer unsigned", Constraints: "NOT NULL"},
					{Name: "ipaddr4", Type: "varchar(39)", Constraints: "NULL"},
					{Name: "ipaddr6", Type: "varchar(39)", Constraints: "NULL"},
					{Name: "is_rs_peer", Type: "bool", Constraints: "NOT NULL"},
					{Name: "bfd_support", Type: "bool", Constraints: "NOT NULL"},
					{Name: "operational", Type: "bool", Constraints: "NOT NULL"},
					{Name: "net_id", Type: "integer", Constraints: "NOT NULL REFERENCES peeringdb_network (id)"},
					{Name: "ix_id", Type: "integer", Constraints: "NOT NULL REFERENCES peeringdb_ix (id)"},
					{Name: "ixlan_id", Type: "integer", Constraints: "NOT NULL REFERENCES peeringdb_ixlan (id)"},
					{Name: "net_side", Type: "integer", Constraints: "NULL REFERENCES peeringdb_facility (id)"},
					{Name: "ix_side", Type: "integer", Constraints: "NULL REFERENCES peeringdb_facility (id)"},
				},
			},
		},
		Indexes: []string{
			"CREATE INDEX peeringdb_campus_org_id ON peeringdb_campus (org_id);",
			"CREATE INDEX peeringdb_facility_org_id ON peeringdb_facility (org_id);",
			"CREATE INDEX peeringdb_carrier_org_id ON peeringdb_carrier (org_id);",
			"CREATE INDEX peeringdb_carrier_facility_carrier_id ON peeringdb_carrier_facility (carrier_id);",
			"CREATE INDEX peeringdb_carrier_facility_fac_id ON peeringdb_carrier_facility (fac_id);",
			"CREATE INDEX peeringdb_network_contact_net_id ON peeringdb_network_contact (net_id);",
			"CREATE INDEX peeringdb_network_org_id ON peeringdb_network (org_id);",
			"CREATE INDEX peeringdb_ix_org_id ON peeringdb_ix (org_id);",
			"CREATE INDEX peeringdb_ix_facility_ix_id ON peeringdb_ix_facility (ix_id);",
			"CREATE INDEX peeringdb_ix_facility_fac_id ON peeringdb_ix_facility (fac_id);",
			"CREATE INDEX peeringdb_ixlan_ix_id ON peeringdb_ixlan (ix_id);",
			"CREATE INDEX peeringdb_ix_prefix_ixlan_id ON peeringdb_ix_prefix (ixlan_id);",
			"CREATE INDEX peeringdb_network_facility_net_id ON peeringdb_network_facility (net_id);",
			"CREATE INDEX peeringdb_network_facility_fac_id ON peeringdb_network_facility (fac_id);",
			"CREATE INDEX peeringdb_network_ixlan_ixlan_id ON peeringdb_network_ixlan (ixlan_id);",
			"CREATE INDEX peeringdb_network_ixlan_net_id ON peeringdb_network_ixlan (net_id);",
			"CREATE INDEX peeringdb_network_ixlan_ix_side ON peeringdb_network_ixlan (ix_side);",
			"CREATE INDEX peeringdb_network_ixlan_net_side ON peeringdb_network_ixlan (net_side);",
		},
	}
}

// GetDatabaseConnection returns a connection to the SQLite database.
func GetDatabaseConnection(filename string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// DeleteDatabase deletes the given SQLite database file.
func DeleteDatabase(filename string) error {
	if _, err := os.Stat(filename); err == nil {
		err = os.Remove(filename)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateDatabase creates a new SQLite database file, removing the existing file if needed.
func CreateDatabase(filename string, removeExisting bool) (*sql.DB, error) {
	if removeExisting {
		if err := DeleteDatabase(filename); err != nil {
			return nil, err
		}
	}

	// Open the SQLite database, will create it if needed
	db, err := GetDatabaseConnection(filename)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// CreateDatabaseSchema creates the given schema in the SQLite database.
func CreateDatabaseSchema(db *sql.DB, schema *Schema) (*sql.Result, error) {
	result, err := db.Exec(schema.GenerateSchemaQuery())
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ClearDatabase removes all data from the SQLite database keeping the schema.
func ClearDatabase(db *sql.DB, schema *Schema) error {
	for _, table := range schema.Tables {
		// Delete data
		_, err := db.Exec("DELETE FROM " + table.Name)
		if err != nil {
			return err
		}

		// Reset auto-increment
		_, err = db.Exec("DELETE FROM SQLITE_SEQUENCE WHERE name= '" + table.Name + "'")
		if err != nil {
			return err
		}
	}
	return nil
}
