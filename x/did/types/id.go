package types

import (
	"errors"
	"strconv"
	"strings"
	"github.com/google/uuid"
	"github.com/mr-tron/base58"
)

const (
	IndyIDLength = 16
	StandardUUIDLength = 36
)

func ValidateID(id string) error {
	isValidID := IsValidIndyID(id) || IsValidUUID(id)

	if !isValidID {
		return errors.New("unique id should be one of: 16 bytes of decoded base58 string or UUID")
	}

	return nil
}

func IsValidID(id string) bool {
	err := ValidateID(id)
	return err == nil
}

func IsValidIndyID(data string) bool {
	bytes, err := base58.Decode(data)
	if err != nil {
		return false
	}
	return len(bytes) == IndyIDLength
}

func NormalizeID(id string) string {
	if IsValidUUID(id) {
		return NormalizeUUID(id)
	}
	return id
}

func NormalizeIDList(keys []string) []string {
	if keys == nil {
		return nil
	}
	newKeys := []string{}
	for _, id := range keys {
		newKeys = append(newKeys, NormalizeID(id))
	}
	return newKeys
}

func ValidateUUID(u string) error {
	if len(u) != StandardUUIDLength {
		return errors.New("uuid must be of length " + strconv.Itoa(StandardUUIDLength) + " (in form of xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)")
	}

	_, err := uuid.Parse(u)
	return err
}

func IsValidUUID(u string) bool {
	return ValidateUUID(u) == nil
}

func NormalizeUUID(uuid string) string {
	return strings.ToLower(uuid)
}