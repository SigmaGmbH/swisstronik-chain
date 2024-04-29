package types

import "time"

const (
	// ModuleName defines the module name
	ModuleName = "vesting"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_vesting"
)

var (
	SecondsOfDay   = int64((time.Hour * 24).Seconds())      // 1 day
	SecondsOfMonth = int64((time.Hour * 24 * 30).Seconds()) // 30 days
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
