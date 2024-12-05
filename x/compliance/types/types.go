package types

func (details *VerificationDetails) IsEmpty() bool {
	if details == nil || details.IssuerAddress == "" {
		return true
	}

	return false
}
