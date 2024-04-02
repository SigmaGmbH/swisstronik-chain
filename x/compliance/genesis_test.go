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

func TestGenesis_Validation(t *testing.T) {
	testCases := []struct {
		name     string
		genState *types.GenesisState
		expPanic bool
	}{
		{
			name:     "default",
			genState: types.DefaultGenesis(),
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
			name: "invalid address",
			genState: &types.GenesisState{
				AddressDetails: []*types.GenesisAddressDetails{
					{Address: "wrong address"},
				},
			},
			expPanic: true,
		},
		{
			name:     "not found verification data", // there's no verification data with verification_id
			genState: &types.GenesisState{},         // todo
			expPanic: true,
		},
		{
			name:     "invalid issuer in verification data",
			genState: &types.GenesisState{}, // todo
			expPanic: true,
		},
		{
			name:     "invalid timestamp in verification data",
			genState: &types.GenesisState{}, // todo
			expPanic: true,
		},
		{
			name:     "issuer not found",
			genState: &types.GenesisState{}, // todo
			expPanic: true,
		},
		{
			name:     "verification id not valid", // Refer: Keeper.AddVerificationDetails
			genState: &types.GenesisState{},       // todo
			expPanic: true,
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

func TestGenesis_Success(t *testing.T) {
	testCases := []struct {
		name     string
		genState *types.GenesisState
		expPanic bool
	}{
		{
			name: "issuers",
			genState: &types.GenesisState{
				Issuers: []*types.IssuerGenesisAccount{
					{
						Address: "cosmos1q9y7yywh0npj2qvgey7geluuhu6j8yyd3gxf2ayk",
						Details: &types.IssuerDetails{
							Name: "test issuer",
						},
					},
					{
						Address: "cosmos1qyzs3crrpxjv2x6j24lc0fykvtj7q2gvcsr8s3nt",
						Details: &types.IssuerDetails{
							Name:        "test issuer2",
							Description: "test description2",
						},
					},
					{
						Address: "cosmos1qgqhhu4rx9yvvdcn7e572njqyhk58swnk2aqgepj2r",
						Details: &types.IssuerDetails{
							Name: "test issuer3",
							Logo: "test logo3",
						},
					},
				},
			},
		},
		{
			name: "addresses",
			genState: &types.GenesisState{
				AddressDetails: []*types.GenesisAddressDetails{
					{
						Address: "cosmos1q9al9ge3frrrwylkd8j5usp9a4pur5ajhgjgry7x",
						Details: &types.AddressDetails{
							IsVerified: true,
							IsRevoked:  false,
							Verifications: []*types.Verification{{
								Type:           types.VerificationType_VT_KYC,
								VerificationId: nil,
								IssuerAddress:  "0x64e739a7d5f9d9e53c7D28be3693Cc0c951d5dC0",
							}},
						},
					},
					{
						Address: "cosmos1qydee7scj8wvd5vemkqtfr6gy7794ul9egn69yfy",
						Details: &types.AddressDetails{
							IsVerified:    true,
							IsRevoked:     false,
							Verifications: nil,
						},
					},
					{
						Address: "cosmos1lytentq4sp4hlrswlwxllppnj5gmhkxvpep445",
						Details: &types.AddressDetails{
							IsVerified: false,
							IsRevoked:  false,
							Verifications: []*types.Verification{{
								Type:           types.VerificationType_VT_KYC,
								VerificationId: nil,
								IssuerAddress:  "0x3826539Cbd8d68DCF119e80B994557B4278CeC9f",
							}},
						},
					},
				},
			},
		},
		{
			name: "verifications",
			genState: &types.GenesisState{
				VerificationDetails: []*types.GenesisVerificationDetails{
					{
						Id: hexutils.HexToBytes("0273FBBAFFC58F732199B20833643248C213C5DBA8F4A05DF505713FD36B8CE2"),
						Details: &types.VerificationDetails{
							IssuerAddress:       "cosmos1qtu30xdvzkqxkluwpmacmluyxw23rw7ces8qtusn",
							OriginChain:         "test chain",
							IssuanceTimestamp:   1712018692,
							ExpirationTimestamp: 1715018692,
							OriginalData:        hexutils.HexToBytes("B639DF194671CDE06EFAA368A404F72E3306DF0359117AC7E78EC2BE04B7629D"),
						},
					},
					{
						Id: hexutils.HexToBytes("1075ee73240c62b820651c22f22f9371dccde1963dec74afffa493902439def2"),
						Details: &types.VerificationDetails{
							IssuerAddress:       "cosmos1qgq3h886rzgae3k3n8wcpdy0fqnmckhnuh9q20czzp",
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
			}
			compliance.InitGenesis(ctx, *k, *tc.genState)

			// Check if issuers were already initialized
			for _, issuerData := range tc.genState.Issuers {
				address, err := sdk.AccAddressFromBech32(issuerData.Address)
				require.NoError(t, err)
				i, err := k.GetIssuerDetails(ctx, address)
				require.NoError(t, err)
				require.NotNil(t, i)
				require.Equal(t, issuerData.Details, i)
			}

			// Check if addresses were already initialized
			for _, addressData := range tc.genState.AddressDetails {
				address, err := sdk.AccAddressFromBech32(addressData.Address)
				require.NoError(t, err)
				i, err := k.GetAddressDetails(ctx, address)
				require.NoError(t, err)
				require.NotNil(t, i)
				require.Equal(t, addressData.Details, i)
			}

			// Check if verification data was already initialized
			for _, verificationData := range tc.genState.VerificationDetails {
				i, err := k.GetVerificationDetails(ctx, verificationData.Id)
				require.NoError(t, err)
				require.NotNil(t, i)
				require.Equal(t, verificationData.Details, i)
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
