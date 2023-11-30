package types

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	SplitDIDRegexp     = regexp.MustCompile(`^did:([^:]+?)(:([^:]+?))?:([^:]+)$`)
	DidNamespaceRegexp = regexp.MustCompile(`^[a-zA-Z0-9]*$`)
)

// TrySplitDID Validates generic format of DID. It doesn't validate method, name and id content.
// Call ValidateDID for further validation.
func TrySplitDID(did string) (method string, id string, err error) {
	// Example: did:swtr:base58str1ng1111
	// match [0] - the whole string
	// match [1] - swtr                 - method
	// match [4] - base58str1ng1111     - id
	matches := SplitDIDRegexp.FindAllStringSubmatch(did, -1)
	if len(matches) != 1 {
		return "", "", errors.New("unable to split did into method, namespace and id")
	}

	match := matches[0]
	return match[1], match[4], nil
}

func MustSplitDID(did string) (method string, namespace string, id string) {
	method, id, err := TrySplitDID(did)
	if err != nil {
		panic(err.Error())
	}
	return
}

func JoinDID(method, id string) string {
	return "did:" + method + ":" + id
}

func ReplaceDIDInDIDURL(didURL string, oldDid string, newDid string) string {
	did, path, query, fragment := MustSplitDIDUrl(didURL)
	if did == oldDid {
		did = newDid
	}

	return JoinDIDUrl(did, path, query, fragment)
}

func ReplaceDIDInDIDURLList(didURLList []string, oldDid string, newDid string) []string {
	res := make([]string, len(didURLList))

	for i := range didURLList {
		res[i] = ReplaceDIDInDIDURL(didURLList[i], oldDid, newDid)
	}

	return res
}

// ValidateDID checks method and allowed namespaces only when the corresponding parameters are specified.
func ValidateDID(did string, method string) error {
	sMethod, sUniqueID, err := TrySplitDID(did)
	if err != nil {
		return err
	}

	// check method
	if method != "" && method != sMethod {
		return fmt.Errorf("did method must be: %s", method)
	}

	// check unique-id
	err = ValidateID(sUniqueID)
	if err != nil {
		return err
	}

	return err
}

func IsValidDID(did string, method string) bool {
	err := ValidateDID(did, method)
	return err == nil
}

func NormalizeDID(did string) string {
	method, _, id := MustSplitDID(did)
	id = NormalizeID(id)
	return JoinDID(method, id)
}

func NormalizeDIDList(didList []string) []string {
	if didList == nil {
		return nil
	}
	newDIDs := []string{}
	for _, did := range didList {
		newDIDs = append(newDIDs, NormalizeDID(did))
	}
	return newDIDs
}