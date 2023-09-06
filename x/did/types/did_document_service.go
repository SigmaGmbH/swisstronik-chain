package types

import (
	"errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func NewService(id string, serviceType string, serviceEndpoint []string) *Service {
	return &Service{
		Id:              id,
		ServiceType:     serviceType,
		ServiceEndpoint: serviceEndpoint,
	}
}

// ReplaceDIDs replaces ids in all fields
func (s *Service) ReplaceDIDs(old, new string) {
	// Id
	s.Id = ReplaceDIDInDIDURL(s.Id, old, new)
}

// Helpers

func GetServiceIds(vms []*Service) []string {
	res := make([]string, len(vms))

	for i := range vms {
		res[i] = vms[i].Id
	}

	return res
}

// Validation

func (s Service) Validate(baseDID string, allowedNamespaces []string) error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Id, validation.Required, IsDIDUrl(allowedNamespaces, Empty, Empty, Required), HasPrefix(baseDID)),
		validation.Field(&s.ServiceType, validation.Required, validation.Length(1, 255)),
		validation.Field(&s.ServiceEndpoint, validation.Each(validation.Required)),
	)
}

func ValidServiceRule(baseDID string, allowedNamespaces []string) *CustomErrorRule {
	return NewCustomErrorRule(func(value interface{}) error {
		casted, ok := value.(Service)
		if !ok {
			panic("ValidVerificationMethodRule must be only applied on verification methods")
		}

		return casted.Validate(baseDID, allowedNamespaces)
	})
}

func IsUniqueServiceListByIDRule() *CustomErrorRule {
	return NewCustomErrorRule(func(value interface{}) error {
		casted, ok := value.([]*Service)
		if !ok {
			panic("IsUniqueServiceListByIdRule must be only applied on service lists")
		}

		ids := GetServiceIds(casted)
		if !IsUnique(ids) {
			return errors.New("there are service duplicates")
		}

		return nil
	})
}