package db

import (
	"testing"
)

func TestGetTableName(t *testing.T) {
	name := "rainbow"
	tName := GetTableNameHash("balance_hash", name)
	t.Log(tName)
}
