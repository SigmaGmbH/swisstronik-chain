package types

import (
	"errors"
	"time"

	"sigs.k8s.io/yaml"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

var (
	_ authtypes.AccountI          = (*MonthlyVestingAccount)(nil)
	_ vestexported.VestingAccount = (*MonthlyVestingAccount)(nil)
	_ authtypes.GenesisAccount    = (*MonthlyVestingAccount)(nil)
)

func NewMonthlyVestingAccountRaw(baseAcc *sdkvesting.BaseVestingAccount, startTime, cliffTime int64, periods []sdkvesting.Period) *MonthlyVestingAccount {
	return &MonthlyVestingAccount{
		BaseVestingAccount: baseAcc,
		StartTime:          startTime,
		CliffTime:          cliffTime,
		VestingPeriods:     periods,
	}
}

func NewMonthlyVestingAccount(baseAcc *authtypes.BaseAccount, originalVesting sdk.Coins, startTime, cliffDays, months int64) *MonthlyVestingAccount {
	afterCliffDays := startTime + SecondsOfDay*cliffDays

	var periods []sdkvesting.Period
	amount := originalVesting.QuoInt(sdk.NewInt(months))
	accVesting := sdk.NewCoins()
	// Every month, vests same amount of coins
	for i := 0; i < int(months)-1; i++ {
		periods = append(periods, sdkvesting.Period{Length: SecondsOfMonth, Amount: amount})
		accVesting = accVesting.Add(amount...)
	}
	if months > 0 {
		// At the last month, includes rest of vesting amount
		periods = append(periods, sdkvesting.Period{
			Length: SecondsOfMonth,
			Amount: originalVesting.Sub(accVesting...),
		})
	}

	endTime := afterCliffDays
	for _, p := range periods {
		endTime += p.Length
	}
	baseVestingAcc := sdkvesting.NewBaseVestingAccount(baseAcc, originalVesting, endTime)

	return &MonthlyVestingAccount{
		BaseVestingAccount: baseVestingAcc,
		StartTime:          startTime,
		CliffTime:          afterCliffDays,
		VestingPeriods:     periods,
	}
}

// GetVestedCoins returns the total number of vested coins. If no coins are vested,
// nil is returned.
func (m MonthlyVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	var vestedCoins sdk.Coins

	if blockTime.Unix() <= m.StartTime || blockTime.Unix() <= m.CliffTime {
		return vestedCoins
	} else if blockTime.Unix() >= m.EndTime {
		return m.OriginalVesting
	}

	// Track the start time of the next period
	currentPeriodCliffTime := m.CliffTime

	// for each period, if the period is over, add those coins as vested and check the next period.
	for _, period := range m.VestingPeriods {
		x := blockTime.Unix() - currentPeriodCliffTime
		if x < period.Length {
			break
		}

		vestedCoins = vestedCoins.Add(period.Amount...)

		// update the start time of the next period
		currentPeriodCliffTime += period.Length
	}

	return vestedCoins
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (m MonthlyVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return m.OriginalVesting.Sub(m.GetVestedCoins(blockTime)...)
}

// LockedCoins returns the set of coins that are not spendable (i.e. locked),
// defined as the vesting coins that are not delegated.
func (m MonthlyVestingAccount) LockedCoins(blockTime time.Time) sdk.Coins {
	return m.BaseVestingAccount.LockedCoinsFromVesting(m.GetVestingCoins(blockTime))
}

// TrackDelegation tracks a desired delegation amount by setting the appropriate
// values for the amount of delegated vesting, delegated free, and reducing the
// overall amount of base coins.
func (m *MonthlyVestingAccount) TrackDelegation(blockTime time.Time, balance, amount sdk.Coins) {
	m.BaseVestingAccount.TrackDelegation(balance, m.GetVestingCoins(blockTime), amount)
}

func (m *MonthlyVestingAccount) TrackUndelegation(amount sdk.Coins) {
	m.BaseVestingAccount.TrackUndelegation(amount)
}

// GetStartTime returns the time when vesting period starts
func (m MonthlyVestingAccount) GetStartTime() int64 {
	return m.StartTime
}

// GetCliffTime returns the time when starts vesting
func (m MonthlyVestingAccount) GetCliffTime() int64 {
	return m.CliffTime
}

// Validate checks for errors on the account fields
func (m MonthlyVestingAccount) Validate() error {
	if m.GetStartTime() >= m.GetEndTime() {
		return errors.New("vesting start-time cannot be before end-time")
	}
	if m.GetStartTime() > m.GetCliffTime() {
		return errors.New("vesting start-time cannot be after cliff-time")
	}
	if m.GetCliffTime() >= m.GetEndTime() {
		return errors.New("vesting cliff-time cannot be after end-time")
	}

	endTime := m.CliffTime
	originalVesting := sdk.NewCoins()
	for _, p := range m.VestingPeriods {
		endTime += p.Length
		originalVesting = originalVesting.Add(p.Amount...)
	}
	if endTime != m.EndTime {
		return errors.New("vesting end time does not match length of all vesting periods")
	}
	if !originalVesting.IsEqual(m.OriginalVesting) {
		return errors.New("original vesting coins does not match the sum of all coins in vesting periods")
	}

	return m.BaseVestingAccount.Validate()
}

func (m MonthlyVestingAccount) String() string {
	out, _ := m.MarshalYAML()
	return out.(string)
}

// MarshalYAML returns the YAML representation of a PeriodicVestingAccount.
func (m MonthlyVestingAccount) MarshalYAML() (interface{}, error) {
	accAddr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return nil, err
	}

	out := vestingAccountYAML{
		Address:          accAddr,
		AccountNumber:    m.AccountNumber,
		PubKey:           getPKString(m),
		Sequence:         m.Sequence,
		OriginalVesting:  m.OriginalVesting,
		DelegatedFree:    m.DelegatedFree,
		DelegatedVesting: m.DelegatedVesting,
		EndTime:          m.EndTime,
		StartTime:        m.StartTime,
		CliffTime:        m.CliffTime,
		VestingPeriods:   m.VestingPeriods,
	}
	return marshalYaml(out)
}

type vestingAccountYAML struct {
	Address          sdk.AccAddress `json:"address"`
	PubKey           string         `json:"public_key"`
	AccountNumber    uint64         `json:"account_number"`
	Sequence         uint64         `json:"sequence"`
	OriginalVesting  sdk.Coins      `json:"original_vesting"`
	DelegatedFree    sdk.Coins      `json:"delegated_free"`
	DelegatedVesting sdk.Coins      `json:"delegated_vesting"`
	EndTime          int64          `json:"end_time"`

	// custom fields based on concrete vesting type which can be omitted
	StartTime      int64              `json:"start_time,omitempty"`
	CliffTime      int64              `json:"cliff_time,omitempty"`
	VestingPeriods sdkvesting.Periods `json:"vesting_periods,omitempty"`
}

type getPK interface {
	GetPubKey() cryptotypes.PubKey
}

func getPKString(g getPK) string {
	if pk := g.GetPubKey(); pk != nil {
		return pk.String()
	}
	return ""
}

func marshalYaml(i interface{}) (interface{}, error) {
	bz, err := yaml.Marshal(i)
	if err != nil {
		return nil, err
	}
	return string(bz), nil
}
