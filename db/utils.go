package db

import (
	"fmt"
	"strings"
)

var (
	BaseKey = [10]uint64{124689, 15, 1835, 6, 123, 51, 9, 13, 81, 156}
)

func getIndex(key string) uint64 {
	d := []byte(key)
	baseKeyLen := len(BaseKey)
	sum := uint64(0)
	for i, v := range d {
		if i >= baseKeyLen {
			break
		}
		sum += uint64(v) * BaseKey[i]
	}
	return sum
}

var (
	shardingByID1 = map[string]uint64{
		"block_id":        100 * 10000,
		"block_tx_rel_id": 100 * 10000,
	}
	shardingByHash = map[string]uint64{
		"txs_hash":                    32,
		"actions_hash":                32,
		"internal_actions_hash":       32,
		"fee_actions_hash":            32,
		"balance_hash":                32,
		"account_action_history_hash": 32,
	}
	shardingByID2 = map[string]uint64{
		"token_fee_history_id": 32,
		// "token_history_id":     32,
	}
)

func GetTableNameID1(table string, id uint64) string {
	var tableName string
	txCount, ok := shardingByID1[table]
	if !ok {
		return tableName
	}

	tableIndex := id / txCount
	tableName = strings.Replace(table, "id", fmt.Sprintf("%05d", tableIndex), -1)
	return tableName
}

func GetTableNameHash(table string, hash string) string {
	var tableName string
	tableCount, ok := shardingByHash[table]
	if !ok {
		return tableName
	}

	tableIndex := getIndex(hash) % tableCount
	tableName = strings.Replace(table, "hash", fmt.Sprintf("%05d", tableIndex), -1)
	return tableName
}

func GetTableNameID2(table string, id uint64) string {
	var tableName string
	txCount, ok := shardingByID2[table]
	if !ok {
		return tableName
	}

	tableIndex := id % txCount
	tableName = strings.Replace(table, "id", fmt.Sprintf("%05d", tableIndex), -1)
	return tableName
}
