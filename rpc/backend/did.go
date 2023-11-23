package backend

import (
	rpctypes "swisstronik/rpc/types"
	didtypes "swisstronik/x/did/types"
)

// DIDResolve return DID Document metadata
func (b *Backend) DIDResolve(blockNrOrHash rpctypes.BlockNumberOrHash, Id string) (*didtypes.DIDDocumentWithMetadata, error) {
	blockNum, err := b.BlockNumberFromTendermint(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	ctx := rpctypes.ContextWithHeight(blockNum.Int64())
	req := didtypes.QueryDIDDocumentRequest{Id: Id}
	res, err := b.queryClient.DidQueryClient.DIDDocument(ctx, &req)
	if err != nil {
		return nil, err
	}

	return res.Value, nil
}

// DocumentsControlledBy returns list of DID Documents controlled by provided verification method
func (b *Backend) DocumentsControlledBy(blockNrOrHash rpctypes.BlockNumberOrHash, VerificationMaterial string) ([]string, error) {
	blockNum, err := b.BlockNumberFromTendermint(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	ctx := rpctypes.ContextWithHeight(blockNum.Int64())
	req := didtypes.QueryAllControlledDIDDocumentsRequest{ VerificationMaterial: VerificationMaterial }
	res, err := b.queryClient.DidQueryClient.AllControlledDIDDocuments(ctx, &req)
	if err != nil {
		return nil, err
	}

	return res.ControlledDocuments, nil
}
