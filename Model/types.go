package model

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/lib/pq"
)

// StringArray mapeia []string para text[] no Postgres.
type StringArray []string

// Value converte StringArray para o formato esperado pelo driver.
func (a StringArray) Value() (driver.Value, error) {
	return pq.StringArray(a).Value()
}

// Scan le valores do banco e popula a StringArray.
func (a *StringArray) Scan(src any) error {
	return (*pq.StringArray)(a).Scan(src)
}

// MarshalJSON garante serializacao como array JSON.
func (a StringArray) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string(a))
}

// UnmarshalJSON garante desserializacao como array JSON.
func (a *StringArray) UnmarshalJSON(data []byte) error {
	var items []string
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	*a = StringArray(items)
	return nil
}
