package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONB adalah tipe kustom untuk menangani tipe data JSONB di PostgreSQL.
type JSONB map[string]interface{}

// Value mengimplementasikan interface driver.Valuer untuk GORM.
func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan mengimplementasikan interface sql.Scanner untuk GORM.
func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &j)
}
