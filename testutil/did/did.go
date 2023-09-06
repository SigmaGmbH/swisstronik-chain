package did

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"math/big"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/mr-tron/base58"
	"github.com/multiformats/go-multibase"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"swisstronik/x/did/keeper"
	"swisstronik/x/did/types"
)

const (
	Base58_16bytes   IDType = iota
	Base58_16symbols
	UUID            
)

var letters = []rune("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

type IDType int

type KeyPair struct {
	Private ed25519.PrivateKey
	Public  ed25519.PublicKey
}

type SignInput struct {
	VerificationMethodID string
	Key                  ed25519.PrivateKey
}

type DIDDocumentInfo struct {
	Msg          *types.MsgCreateDIDDocumentPayload
	Did          string
	CollectionID string
	KeyPair      KeyPair
	KeyID        string
	SignInput    SignInput
}

type CreatedDIDDocumentInfo struct {
	DIDDocumentInfo
	VersionID string
}

func GenerateDID(idtype IDType) string {
	prefix := "did:swtr:"

	switch idtype {
	case Base58_16bytes:
		return prefix + RandBase58String(16)
	case Base58_16symbols:
		return prefix + RandString(16)
	case UUID:
		return prefix + uuid.NewString()
	default:
		panic("Unknown ID type")
	}
}

func GenerateKeyPair() KeyPair {
	PublicKey, PrivateKey, _ := ed25519.GenerateKey(rand.Reader)
	return KeyPair{PrivateKey, PublicKey}
}

func GenerateEd25519VerificationKey2020VerificationMaterial(publicKey ed25519.PublicKey) string {
	publicKeyMultibaseBytes := []byte{0xed, 0x01}
	publicKeyMultibaseBytes = append(publicKeyMultibaseBytes, publicKey...)
	keyStr, _ := multibase.Encode(multibase.Base58BTC, publicKeyMultibaseBytes)
	return keyStr
}

func GenerateJSONWebKey2020VerificationMaterial(publicKey ed25519.PublicKey) string {
	pubKeyJwk, err := jwk.New(publicKey)
	if err != nil {
		panic(err)
	}

	pubKeyJwkJSON, err := json.Marshal(pubKeyJwk)
	if err != nil {
		panic(err)
	}

	return string(pubKeyJwkJSON)
}

func GenerateEd25519VerificationKey2018VerificationMaterial(publicKey ed25519.PublicKey) string {
	return base58.Encode(publicKey)
}

func RandBase58String(bytes int) string {
	b := make([]byte, bytes)

	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return base58.Encode(b)
}

func RandString(lenght int) string {
	b := make([]rune, lenght)
	for i := range b {
		letterIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}
		b[i] = letters[letterIndex.Int64()]
	}
	return string(b)
}

func CreateDID(ctx sdk.Context, didKeeper keeper.Keeper, payload *types.MsgCreateDIDDocumentPayload, signInputs []SignInput) (*types.MsgCreateDIDDocumentResponse, error) {
	signBytes := payload.GetSignBytes()
	signatures := make([]*types.SignInfo, 0, len(signInputs))

	for _, input := range signInputs {
		signature := ed25519.Sign(input.Key, signBytes)

		signatures = append(signatures, &types.SignInfo{
			VerificationMethodId: input.VerificationMethodID,
			Signature:            signature,
		})
	}

	msg := &types.MsgCreateDIDDocument{
		Payload:    payload,
		Signatures: signatures,
	}

	return didKeeper.CreateDIDDocument(sdk.WrapSDKContext(ctx), msg)
}

func DefaultDIDDocumentWithDID(did string) DIDDocumentInfo {
	_, _, collectionID := types.MustSplitDID(did)

	keyPair := GenerateKeyPair()
	keyID := did + "#key-1"

	payload := &types.MsgCreateDIDDocumentPayload{
		Id: did,
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     keyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             did,
				VerificationMaterial:   GenerateEd25519VerificationKey2020VerificationMaterial(keyPair.Public),
			},
		},
		Authentication: []string{keyID},
		VersionId:      uuid.NewString(),
	}

	signInput := SignInput{
		VerificationMethodID: keyID,
		Key:                  keyPair.Private,
	}

	return DIDDocumentInfo{
		Did:          did,
		CollectionID: collectionID,
		KeyPair:      keyPair,
		KeyID:        keyID,
		Msg:          payload,
		SignInput:    signInput,
	}
}

func DefaultDIDDocumentWithRandomDID() DIDDocumentInfo {
	did := GenerateDID(Base58_16bytes)
	return DefaultDIDDocumentWithDID(did)
}

func CreateCustomDIDDocument(ctx sdk.Context, didKeeper keeper.Keeper, info DIDDocumentInfo) (CreatedDIDDocumentInfo, error) {
	created, err := CreateDID(ctx, didKeeper, info.Msg, []SignInput{info.SignInput})
	if err != nil {
		return CreatedDIDDocumentInfo{}, err
	}

	return CreatedDIDDocumentInfo{
		DIDDocumentInfo: info,
		VersionID:  created.Value.Metadata.VersionId,
	}, nil
}

func CreateDefaultDID(ctx sdk.Context, didKeeper keeper.Keeper) (CreatedDIDDocumentInfo, error) {
	did := DefaultDIDDocumentWithRandomDID()
	return CreateCustomDIDDocument(ctx, didKeeper, did)
}

func CreateDIDDocumentWithExternalControllers(ctx sdk.Context, didKeeper keeper.Keeper, controllers []string, signInputs []SignInput) (CreatedDIDDocumentInfo, error) {
	did := DefaultDIDDocumentWithRandomDID()
	did.Msg.Controller = append(did.Msg.Controller, controllers...)

	created, err := CreateDID(ctx, didKeeper, did.Msg, append(signInputs, did.SignInput))
	if err != nil {
		return CreatedDIDDocumentInfo{}, err
	}

	return CreatedDIDDocumentInfo{
		DIDDocumentInfo: did,
		VersionID:  created.Value.Metadata.VersionId,
	}, nil
}

func DeactivateDIDDocument(ctx sdk.Context, keeper keeper.Keeper, payload *types.MsgDeactivateDIDDocumentPayload, signInputs []SignInput) (*types.MsgDeactivateDIDDocumentResponse, error) {
	signBytes := payload.GetSignBytes()
	signatures := make([]*types.SignInfo, 0, len(signInputs))

	for _, input := range signInputs {
		signature := ed25519.Sign(input.Key, signBytes)

		signatures = append(signatures, &types.SignInfo{
			VerificationMethodId: input.VerificationMethodID,
			Signature:            signature,
		})
	}

	msg := &types.MsgDeactivateDIDDocument{
		Payload:    payload,
		Signatures: signatures,
	}

	return keeper.DeactivateDIDDocument(sdk.WrapSDKContext(ctx), msg)
}

func UpdateDIDDocument(ctx sdk.Context, keeper keeper.Keeper, payload *types.MsgUpdateDIDDocumentPayload, signInputs []SignInput) (*types.MsgUpdateDIDDocumentResponse, error) {
	signBytes := payload.GetSignBytes()
	signatures := make([]*types.SignInfo, 0, len(signInputs))

	for _, input := range signInputs {
		signature := ed25519.Sign(input.Key, signBytes)

		signatures = append(signatures, &types.SignInfo{
			VerificationMethodId: input.VerificationMethodID,
			Signature:            signature,
		})
	}

	msg := &types.MsgUpdateDIDDocument{
		Payload:    payload,
		Signatures: signatures,
	}

	return keeper.UpdateDIDDocument(sdk.WrapSDKContext(ctx), msg)
}

func GetDIDDocument(ctx sdk.Context, keeper keeper.Keeper, did string) (*types.QueryDIDDocumentResponse, error) {
	return keeper.DIDDocument(sdk.WrapSDKContext(ctx), &types.QueryDIDDocumentRequest{Id: did})
} 

func GetDIDDocumentVersion(ctx sdk.Context, keeper keeper.Keeper, did string, version string) (*types.QueryDIDDocumentVersionResponse, error) {
	return keeper.DIDDocumentVersion(sdk.WrapSDKContext(ctx), &types.QueryDIDDocumentVersionRequest{Id: did, Version: version})
}

func GetAllDIDVersionsMetadata(ctx sdk.Context, keeper keeper.Keeper, did string) (*types.QueryAllDIDDocumentVersionsMetadataResponse, error) {
	return keeper.AllDIDDocumentVersionsMetadata(sdk.WrapSDKContext(ctx), &types.QueryAllDIDDocumentVersionsMetadataRequest{Id: did})
}