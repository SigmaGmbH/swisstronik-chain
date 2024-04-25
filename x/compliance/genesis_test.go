package compliance_test

import (
	"sort"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/status-im/keycard-go/hexutils"
	"github.com/stretchr/testify/require"

	testkeeper "swisstronik/testutil/keeper"
	"swisstronik/x/compliance"
	"swisstronik/x/compliance/types"
)

func init() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("swtr", "swtrpub")
}

func TestInitGenesis_Validation(t *testing.T) {
	testCases := []struct {
		name     string
		genState *types.GenesisState
		expPanic bool
	}{
		{
			name: "invalid operators",
			genState: &types.GenesisState{
				Operators: []*types.OperatorDetails{
					{Operator: "wrong address"},
				},
			},
			expPanic: true,
		},
		{
			name: "invalid issuers",
			genState: &types.GenesisState{
				Issuers: []*types.IssuerGenesisAccount{
					{Address: "wrong address"},
				},
			},
			expPanic: true,
		},
		{
			name: "invalid issuer details",
			genState: &types.GenesisState{
				Issuers: []*types.IssuerGenesisAccount{
					{Address: "swtr1tpvqt6zfl9yef58gl7jcdpkw88thgrkf38d5zx"},
				},
			},
			expPanic: true,
		},
		{
			name: "invalid issuer in verification data",
			genState: &types.GenesisState{
				VerificationDetails: []*types.GenesisVerificationDetails{
					{
						Id: hexutils.HexToBytes("0273FBBAFFC58F732199B20833643248C213C5DBA8F4A05DF505713FD36B8CE2"),
						Details: &types.VerificationDetails{
							IssuerAddress: "invalid issuer",
						},
					},
				},
			},
			expPanic: true,
		},
		{
			name: "invalid timestamp in verification data",
			genState: &types.GenesisState{
				VerificationDetails: []*types.GenesisVerificationDetails{
					{
						Id: hexutils.HexToBytes("0273FBBAFFC58F732199B20833643248C213C5DBA8F4A05DF505713FD36B8CE2"),
						Details: &types.VerificationDetails{
							IssuerAddress:       "swtr199wynlfwhj6ytkvujjf6mel5z7fl0mwzqck8l6",
							OriginChain:         "test chain",
							IssuanceTimestamp:   1715018692,
							ExpirationTimestamp: 1712018692,
							OriginalData:        hexutils.HexToBytes("B639DF194671CDE06EFAA368A404F72E3306DF0359117AC7E78EC2BE04B7629D"),
						},
					},
				},
			},
			expPanic: true,
		},
		{
			name: "no proof in verification data",
			genState: &types.GenesisState{
				VerificationDetails: []*types.GenesisVerificationDetails{
					{
						Id: hexutils.HexToBytes("0273FBBAFFC58F732199B20833643248C213C5DBA8F4A05DF505713FD36B8CE2"),
						Details: &types.VerificationDetails{
							IssuerAddress:       "swtr199wynlfwhj6ytkvujjf6mel5z7fl0mwzqck8l6",
							OriginChain:         "test chain",
							IssuanceTimestamp:   1712018692,
							ExpirationTimestamp: 1715018692,
							OriginalData:        nil,
						},
					},
				},
			},
			expPanic: true,
		},
		{
			name: "invalid account address",
			genState: &types.GenesisState{
				AddressDetails: []*types.GenesisAddressDetails{
					{Address: "wrong address"},
				},
			},
			expPanic: true,
		},
		{
			name: "issuer of verification not found for verified account ",
			genState: &types.GenesisState{
				AddressDetails: []*types.GenesisAddressDetails{
					{
						Address: "swtr1996rrzmj36jjd6hmfenluhxs664pdg3aewe3le",
						Details: &types.AddressDetails{
							IsVerified: true,
							IsRevoked:  false,
							Verifications: []*types.Verification{{
								Type:           types.VerificationType_VT_KYC,
								VerificationId: nil,
								IssuerAddress:  "swtr199wynlfwhj6ytkvujjf6mel5z7fl0mwzqck8l6",
							}},
						},
					},
				},
			},
			expPanic: true,
		},
		{
			name: "verification id for verified account is nil", // there's no verification data with verification_id
			genState: &types.GenesisState{
				Issuers: []*types.IssuerGenesisAccount{
					{
						Address: "swtr199wynlfwhj6ytkvujjf6mel5z7fl0mwzqck8l6",
						Details: &types.IssuerDetails{
							Name: "test issuer",
						},
					},
				},
				AddressDetails: []*types.GenesisAddressDetails{
					{
						Address: "swtr1996rrzmj36jjd6hmfenluhxs664pdg3aewe3le",
						Details: &types.AddressDetails{
							IsVerified: true,
							IsRevoked:  false,
							Verifications: []*types.Verification{{
								Type:           types.VerificationType_VT_KYC,
								VerificationId: nil,
								IssuerAddress:  "swtr199wynlfwhj6ytkvujjf6mel5z7fl0mwzqck8l6",
							}},
						},
					},
				},
			},
			expPanic: true,
		},
		{
			// There's no verification data with verification_id
			name: "not found verification data for verified account",
			genState: &types.GenesisState{
				Issuers: []*types.IssuerGenesisAccount{
					{
						Address: "swtr199wynlfwhj6ytkvujjf6mel5z7fl0mwzqck8l6",
						Details: &types.IssuerDetails{
							Name: "test issuer",
						},
					},
				},
				AddressDetails: []*types.GenesisAddressDetails{
					{
						Address: "swtr1996rrzmj36jjd6hmfenluhxs664pdg3aewe3le",
						Details: &types.AddressDetails{
							IsVerified: true,
							IsRevoked:  false,
							Verifications: []*types.Verification{{
								Type:           types.VerificationType_VT_KYC,
								VerificationId: hexutils.HexToBytes("1075ee73240c62b820651c22f22f9371dccde1963dec74afffa493902439def2"),
								IssuerAddress:  "swtr199wynlfwhj6ytkvujjf6mel5z7fl0mwzqck8l6",
							}},
						},
					},
				},
				VerificationDetails: []*types.GenesisVerificationDetails{
					{
						Id: hexutils.HexToBytes("0273FBBAFFC58F732199B20833643248C213C5DBA8F4A05DF505713FD36B8CE2"),
						Details: &types.VerificationDetails{
							IssuerAddress:       "swtr199wynlfwhj6ytkvujjf6mel5z7fl0mwzqck8l6",
							OriginChain:         "test chain",
							IssuanceTimestamp:   1712018692,
							ExpirationTimestamp: 1715018692,
							OriginalData:        hexutils.HexToBytes("B639DF194671CDE06EFAA368A404F72E3306DF0359117AC7E78EC2BE04B7629D"),
						},
					},
				},
			},
			expPanic: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			k, ctx := testkeeper.ComplianceKeeper(t)

			if tc.expPanic {
				require.Panics(t,
					func() {
						compliance.InitGenesis(ctx, *k, *tc.genState)
					},
				)
				return
			}
			compliance.InitGenesis(ctx, *k, *tc.genState)
		})
	}
}

func TestGenesis_Default(t *testing.T) {
	// Initialize test keeper and context
	k, ctx := testkeeper.ComplianceKeeper(t)

	// Generate a sample genesis state
	defaultGenesis := types.DefaultGenesis()

	// Import genesis state
	compliance.InitGenesis(ctx, *k, *defaultGenesis)

	// Export genesis state
	exportedGenesis := compliance.ExportGenesis(ctx, *k)

	// Ensure exported genesis state matches the sample genesis state
	require.Equal(t, defaultGenesis.Params, exportedGenesis.Params)
}

func TestGenesis_Success(t *testing.T) {
	testCases := []struct {
		name     string
		genState *types.GenesisState
		expPanic bool
	}{
		{
			name: "valid issuers, verifications and addresses",
			genState: &types.GenesisState{
				Operators: []*types.OperatorDetails{
					{
						Operator:     "swtr15srdmqa9934z6utqywsagt456va5xwjpwvmpth",
						OperatorType: types.OperatorType_OT_INITIAL,
					},
					{
						Operator:     "swtr16vgqffr8v0sh3n5qeqdksfpzdkqf3rtk49thun",
						OperatorType: types.OperatorType_OT_REGULAR,
					},
				},
				Issuers: []*types.IssuerGenesisAccount{
					{
						Address: "swtr199wynlfwhj6ytkvujjf6mel5z7fl0mwzqck8l6",
						Details: &types.IssuerDetails{
							Name: "test issuer",
						},
					},
					{
						Address: "swtr13wl63dpe3xdhzvphp32cm9cv2vs9nvhkpaspwu",
						Details: &types.IssuerDetails{
							Name:        "test issuer2",
							Description: "test description2",
						},
					},
				},
				AddressDetails: []*types.GenesisAddressDetails{
					{
						Address: "swtr13yc35xh4r8ap7y440sex4nzxggxdgv7ly0cchg",
						Details: &types.AddressDetails{
							IsVerified:    true,
							IsRevoked:     false,
							Verifications: nil,
						},
					},
					{
						Address: "swtr1996rrzmj36jjd6hmfenluhxs664pdg3aewe3le",
						Details: &types.AddressDetails{
							IsVerified: true,
							IsRevoked:  false,
							Verifications: []*types.Verification{{
								Type:           types.VerificationType_VT_KYC,
								VerificationId: hexutils.HexToBytes("1075ee73240c62b820651c22f22f9371dccde1963dec74afffa493902439def2"),
								IssuerAddress:  "swtr199wynlfwhj6ytkvujjf6mel5z7fl0mwzqck8l6",
							}},
						},
					},
					{
						Address: "swtr1flhu6pdk2ydrjqryn9utq7v5mxsr8ka67fmjj6",
						Details: &types.AddressDetails{
							IsVerified: false,
							IsRevoked:  false,
							Verifications: []*types.Verification{{
								Type:           types.VerificationType_VT_KYC,
								VerificationId: hexutils.HexToBytes("0273FBBAFFC58F732199B20833643248C213C5DBA8F4A05DF505713FD36B8CE2"),
								IssuerAddress:  "swtr199wynlfwhj6ytkvujjf6mel5z7fl0mwzqck8l6",
							}},
						},
					},
				},
				VerificationDetails: []*types.GenesisVerificationDetails{
					{
						Id: hexutils.HexToBytes("0273FBBAFFC58F732199B20833643248C213C5DBA8F4A05DF505713FD36B8CE2"),
						Details: &types.VerificationDetails{
							IssuerAddress:       "swtr199wynlfwhj6ytkvujjf6mel5z7fl0mwzqck8l6",
							OriginChain:         "test chain",
							IssuanceTimestamp:   1712018692,
							ExpirationTimestamp: 1715018692,
							OriginalData:        hexutils.HexToBytes("B639DF194671CDE06EFAA368A404F72E3306DF0359117AC7E78EC2BE04B7629D"),
						},
					},
					{
						Id: hexutils.HexToBytes("1075ee73240c62b820651c22f22f9371dccde1963dec74afffa493902439def2"),
						Details: &types.VerificationDetails{
							IssuerAddress:       "swtr199wynlfwhj6ytkvujjf6mel5z7fl0mwzqck8l6",
							OriginChain:         "test chain",
							IssuanceTimestamp:   1712022843,
							ExpirationTimestamp: 1712052843,
							OriginalData:        hexutils.HexToBytes("0ce39a77d630007ff1b8289d878ec30822a7ee6bfdd1b2d6329edab93d2db2da"),
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			k, ctx := testkeeper.ComplianceKeeper(t)

			if tc.expPanic {
				require.Panics(t,
					func() {
						compliance.InitGenesis(ctx, *k, *tc.genState)
					},
				)
				return
			} else {
				require.NotPanics(t, func() {
					compliance.InitGenesis(ctx, *k, *tc.genState)
				})
			}

			// Check if issuers were already initialized
			for _, operatorData := range tc.genState.Operators {
				address, err := sdk.AccAddressFromBech32(operatorData.Operator)
				require.NoError(t, err)
				details, err := k.GetOperatorDetails(ctx, address)
				require.NoError(t, err)
				require.NotNil(t, details)
				require.Equal(t, operatorData, details)
			}

			// Check if issuers were already initialized
			for _, issuerData := range tc.genState.Issuers {
				address, err := sdk.AccAddressFromBech32(issuerData.Address)
				require.NoError(t, err)
				details, err := k.GetIssuerDetails(ctx, address)
				require.NoError(t, err)
				require.NotNil(t, details)
				require.Equal(t, issuerData.Details, details)
			}

			// Check if addresses were already initialized
			for _, addressData := range tc.genState.AddressDetails {
				address, err := sdk.AccAddressFromBech32(addressData.Address)
				require.NoError(t, err)
				details, err := k.GetAddressDetails(ctx, address)
				require.NoError(t, err)
				require.NotNil(t, details)
				require.Equal(t, addressData.Details, details)
			}

			// Check if verification data was already initialized
			for _, verificationData := range tc.genState.VerificationDetails {
				details, err := k.GetVerificationDetails(ctx, verificationData.Id)
				require.NoError(t, err)
				require.NotNil(t, details)
				require.Equal(t, verificationData.Details, details)
			}

			got := compliance.ExportGenesis(ctx, *k)
			require.NotNil(t, got)

			require.Equal(t, tc.genState.Params, got.Params)
			// Sort by issuer address to check if two issuers are same
			sort.Slice(tc.genState.Issuers, func(i, j int) bool { return tc.genState.Issuers[i].Address < tc.genState.Issuers[j].Address })
			sort.Slice(got.Issuers, func(i, j int) bool { return got.Issuers[i].Address < got.Issuers[j].Address })
			require.Equal(t, tc.genState.Issuers, got.Issuers)
			// Sort by address to check if two address details are same
			sort.Slice(tc.genState.AddressDetails, func(i, j int) bool {
				return tc.genState.AddressDetails[i].Address < tc.genState.AddressDetails[j].Address
			})
			sort.Slice(got.AddressDetails, func(i, j int) bool { return got.AddressDetails[i].Address < got.AddressDetails[j].Address })
			require.Equal(t, tc.genState.AddressDetails, got.AddressDetails)
			// Sort by id to check if two verification details are same
			sort.Slice(tc.genState.VerificationDetails, func(i, j int) bool {
				return hexutils.BytesToHex(tc.genState.VerificationDetails[i].Id) < hexutils.BytesToHex(tc.genState.VerificationDetails[j].Id)
			})
			sort.Slice(got.VerificationDetails, func(i, j int) bool {
				return hexutils.BytesToHex(got.VerificationDetails[i].Id) < hexutils.BytesToHex(got.VerificationDetails[j].Id)
			})
			require.Equal(t, tc.genState.VerificationDetails, got.VerificationDetails)
		})
	}
}
