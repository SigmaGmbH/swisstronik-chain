package types

func (query *QueryCollectionResourcesRequest) Normalize() {
	query.CollectionId = NormalizeID(query.CollectionId)
}

func (query *QueryResourceMetadataRequest) Normalize() {
	query.CollectionId = NormalizeID(query.CollectionId)
	query.Id = NormalizeUUID(query.Id)
}

func (query *QueryResourceRequest) Normalize() {
	query.CollectionId = NormalizeID(query.CollectionId)
	query.Id = NormalizeUUID(query.Id)
}