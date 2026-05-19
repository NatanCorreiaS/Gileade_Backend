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

// Uint64Array mapeia []uint64 para bigint[] no Postgres.
type Uint64Array []uint64

// Value converte Uint64Array para o formato esperado pelo driver.
func (a Uint64Array) Value() (driver.Value, error) {
	arr := make(pq.Int64Array, len(a))
	for i, v := range a {
		arr[i] = int64(v)
	}
	return arr.Value()
}

// Scan le valores do banco e popula a Uint64Array.
func (a *Uint64Array) Scan(src any) error {
	var arr pq.Int64Array
	if err := arr.Scan(src); err != nil {
		return err
	}
	out := make(Uint64Array, len(arr))
	for i, v := range arr {
		out[i] = uint64(v)
	}
	*a = out
	return nil
}

// MarshalJSON garante serializacao como array JSON.
func (a Uint64Array) MarshalJSON() ([]byte, error) {
	return json.Marshal([]uint64(a))
}

// UnmarshalJSON garante desserializacao como array JSON.
func (a *Uint64Array) UnmarshalJSON(data []byte) error {
	var items []uint64
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	*a = Uint64Array(items)
	return nil
}
