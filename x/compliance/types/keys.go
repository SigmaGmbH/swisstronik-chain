package types

const (
	// ModuleName defines the module name
	ModuleName = "compliance"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_compliance"
)

const (
	prefixIssuerDetails = iota + 1
	prefixAddressDetails
	prefixVerificationDetails
)

var (
	KeyPrefixIssuerDetails       = []byte{prefixIssuerDetails}
	KeyPrefixAddressDetails      = []byte{prefixAddressDetails}
	KeyPrefixVerificationDetails = []byte{prefixVerificationDetails}
)
