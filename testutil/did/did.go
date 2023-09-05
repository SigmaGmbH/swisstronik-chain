package did

import (
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/mr-tron/base58"
	"github.com/multiformats/go-multibase"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"math/big"

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

func CreateDIDMessage(payload *types.MsgCreateDIDDocumentPayload, signInputs []SignInput) *types.MsgCreateDIDDocument {
	signBytes := payload.GetSignBytes()
	signatures := make([]*types.SignInfo, 0, len(signInputs))

	for _, input := range signInputs {
		signature := ed25519.Sign(input.Key, signBytes)

		signatures = append(signatures, &types.SignInfo{
			VerificationMethodId: input.VerificationMethodID,
			Signature:            signature,
		})
	}

	return &types.MsgCreateDIDDocument{
		Payload:    payload,
		Signatures: signatures,
	}
}
