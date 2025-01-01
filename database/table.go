package database

import (
	"fmt"
	"strings"
)

// Column represents a single column in a table schema.
type Column struct {
	Name        string // Column name
	Type        string // Data type (e.g., INTEGER, VARCHAR, etc.)
	Constraints string // Constraints (e.g., NOT NULL, PRIMARY KEY, etc.)
}

// Table represents a schema of a table.
type Table struct {
	Name                  string   // Table name
	Columns               []Column // List of columns
	UniquenessConstraints []string // List of uniqueness constraints
}

// Schema represents a schema of a database.
type Schema struct {
	Tables  map[string]Table // List of tables
	Indexes []string         // List of indexes
}

// GetColumnsNames returns the list of column names in the table without the "id" one.
func (t *Table) GetColumnsNames() []string {
	names := make([]string, 0, len(t.Columns))
	for _, column := range t.Columns {
		if column.Name != "id" {
			names = append(names, column.Name)
		}
	}
	return names
}

// generateCreateTableQuery generates the SQL CREATE TABLE statement from the Table struct.
func (t *Table) generateCreateTableQuery() string {
	query := fmt.Sprintf("CREATE TABLE %s (\n", t.Name)
	for i, column := range t.Columns {
		query += fmt.Sprintf("  %s %s %s", column.Name, column.Type, column.Constraints)
		if i < len(t.Columns)-1 {
			query += ",\n"
		}
	}

	if t.UniquenessConstraints != nil {
		query += fmt.Sprintf(",\n  UNIQUE (%s)", strings.Join(t.UniquenessConstraints, ", "))
	}

	query += "\n);"
	return query
}

// GenerateSchemaQuery generates the SQL CREATE TABLE statements from the Schema struct.
func (s *Schema) GenerateSchemaQuery() string {
	query := ""
	for _, table := range s.Tables {
		query += table.generateCreateTableQuery() + "\n\n"
	}

	for _, index := range s.Indexes {
		query += index + "\n"
	}

	return query
}

// GetTableNames returns the list of table names in the schema.
func (s *Schema) GetTableNames() []string {
	names := make([]string, 0, len(s.Tables))
	for name := range s.Tables {
		names = append(names, name)
	}
	return names
}
