package types

import (
	"errors"
	"fmt"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/multiformats/go-multibase"
)

// Helper enums

type ValidationType int

const (
	Optional ValidationType = iota
	Required ValidationType = iota
	Empty    ValidationType = iota
)

// Custom error rule

var _ validation.Rule = &CustomErrorRule{}

type CustomErrorRule struct {
	fn func(value interface{}) error
}

func NewCustomErrorRule(fn func(value interface{}) error) *CustomErrorRule {
	return &CustomErrorRule{fn: fn}
}

func (c CustomErrorRule) Validate(value interface{}) error {
	return c.fn(value)
}

// Validation helpers

func IsID() *CustomErrorRule {
	return NewCustomErrorRule(func(value interface{}) error {
		casted, ok := value.(string)
		if !ok {
			panic("IsID must be only applied on string properties")
		}

		return ValidateID(casted)
	})
}

func IsDID() *CustomErrorRule {
	return NewCustomErrorRule(func(value interface{}) error {
		casted, ok := value.(string)
		if !ok {
			panic("IsDID must be only applied on string properties")
		}

		return ValidateDID(casted, DIDMethod)
	})
}

func IsDIDUrl(pathRule, queryRule, fragmentRule ValidationType) *CustomErrorRule {
	return NewCustomErrorRule(func(value interface{}) error {
		casted, ok := value.(string)
		if !ok {
			panic("IsDIDUrl must be only applied on string properties")
		}

		if err := ValidateDIDUrl(casted, DIDMethod); err != nil {
			return err
		}

		_, path, query, fragment, err := TrySplitDIDUrl(casted)
		if err != nil {
			return err
		}

		if pathRule == Required && path == "" {
			return errors.New("path is required")
		}

		if pathRule == Empty && path != "" {
			return errors.New("path must be empty")
		}

		if queryRule == Required && query == "" {
			return errors.New("query is required")
		}

		if queryRule == Empty && query != "" {
			return errors.New("query must be empty")
		}

		if fragmentRule == Required && fragment == "" {
			return errors.New("fragment is required")
		}

		if fragmentRule == Empty && fragment != "" {
			return errors.New("fragment must be empty")
		}

		return nil
	})
}

func IsURI() *CustomErrorRule {
	return NewCustomErrorRule(func(value interface{}) error {
		casted, ok := value.(string)
		if !ok {
			panic("IsURI must be only applied on string properties")
		}

		return ValidateURI(casted)
	})
}

func IsMultibase() *CustomErrorRule {
	return NewCustomErrorRule(func(value interface{}) error {
		casted, ok := value.(string)
		if !ok {
			panic("IsMultibase must be only applied on string properties")
		}

		return ValidateMultibase(casted)
	})
}

func IsMultibaseEd25519VerificationKey2020() *CustomErrorRule {
	return NewCustomErrorRule(func(value interface{}) error {
		casted, ok := value.(string)
		if !ok {
			panic("IsMultibaseEd25519VerificationKey2020 must be only applied on string properties")
		}

		return ValidateMultibaseEd25519VerificationKey2020(casted)
	})
}

func IsBase58Ed25519VerificationKey2018() *CustomErrorRule {
	return NewCustomErrorRule(func(value interface{}) error {
		casted, ok := value.(string)
		if !ok {
			panic("IsBase58Ed25519VerificationKey2018 must be only applied on string properties")
		}

		return ValidateBase58Ed25519VerificationKey2018(casted)
	})
}

func IsMultibaseEncodedEd25519PubKey() *CustomErrorRule {
	return NewCustomErrorRule(func(value interface{}) error {
		casted, ok := value.(string)
		if !ok {
			panic("IsMultibaseEncodedEd25519PubKey must be only applied on string properties")
		}

		_, keyBytes, err := multibase.Decode(casted)
		if err != nil {
			return err
		}

		err = ValidateEd25519PubKey(keyBytes)
		if err != nil {
			return err
		}

		return nil
	})
}

func IsJWK() *CustomErrorRule {
	return NewCustomErrorRule(func(value interface{}) error {
		casted, ok := value.(string)
		if !ok {
			panic("IsJWK must be only applied on string properties")
		}

		return ValidateJWK(casted)
	})
}

func HasPrefix(prefix string) *CustomErrorRule {
	return NewCustomErrorRule(func(value interface{}) error {
		casted, ok := value.(string)
		if !ok {
			panic("HasPrefix must be only applied on string properties")
		}

		if !strings.HasPrefix(casted, prefix) {
			return fmt.Errorf("must have prefix: %s", prefix)
		}

		return nil
	})
}

func IsUniqueStrList() *CustomErrorRule {
	return NewCustomErrorRule(func(value interface{}) error {
		casted, ok := value.([]string)
		if !ok {
			panic("IsSet must be only applied on string array properties")
		}

		if !IsUnique(casted) {
			return errors.New("there should be no duplicates")
		}

		return nil
	})
}

func IsUUID() *CustomErrorRule {
	return NewCustomErrorRule(func(value interface{}) error {
		casted, ok := value.(string)
		if !ok {
			panic("IsDID must be only applied on string properties")
		}

		return ValidateUUID(casted)
	})
}

func (au AlternativeUri) Validate() error {
	return validation.ValidateStruct(&au,
		validation.Field(&au.Uri, validation.Required, validation.Length(1, 256)),
		validation.Field(&au.Description, validation.Length(1, 128)),
	)
}

func ValidAlternativeURI() *CustomErrorRule {
	return NewCustomErrorRule(func(value interface{}) error {
		casted, ok := value.(AlternativeUri)
		if !ok {
			panic("ValidAlternativeUri must be only applied on AlternativeUri properties")
		}

		return casted.Validate()
	})
}