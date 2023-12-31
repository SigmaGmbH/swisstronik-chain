package types

import (
	"fmt"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		VersionSets: []*DIDDocumentVersionSet{},
		Resources: []*ResourceWithMetadata{},
	}
}

func (gs GenesisState) ValidateNoDuplicates() error {
	// Check for duplicates in version set
	didCache := make(map[string]bool)

	for _, versionSet := range gs.VersionSets {
		did := versionSet.DidDocs[0].DidDoc.Id
		if _, ok := didCache[did]; ok {
			return fmt.Errorf("duplicated DID document found with id %s", did)
		}

		didCache[did] = true

		// Check for duplicates in didDoc versions
		versionCache := make(map[string]bool)

		for _, didDoc := range versionSet.DidDocs {
			version := didDoc.Metadata.VersionId
			if _, ok := versionCache[version]; ok {
				return fmt.Errorf("duplicated DID document version found with id %s and version %s", did, version)
			}

			versionCache[version] = true
		}

		// Check that latest version is present
		if _, ok := versionCache[versionSet.LatestVersion]; !ok {
			return fmt.Errorf("latest version not found in DID document with id %s", did)
		}
	}

	// Group resources by collection
	resourcesByCollection := make(map[string][]*ResourceWithMetadata)

	for _, resource := range gs.Resources {
		existing := resourcesByCollection[resource.Metadata.CollectionId]
		resourcesByCollection[resource.Metadata.CollectionId] = append(existing, resource)
	}

	// Check that there are no collisions within each collection
	for _, resources := range resourcesByCollection {
		resourceIDMap := make(map[string]bool)

		for _, resource := range resources {
			if _, ok := resourceIDMap[resource.Metadata.Id]; ok {
				return fmt.Errorf("duplicated id for resource within the same collection. collection: %s, id: %s", resource.Metadata.CollectionId, resource.Metadata.Id)
			}

			resourceIDMap[resource.Metadata.Id] = true
		}
	}

	return nil
}

func (gs GenesisState) ValidateVersionSets() error {
	for _, versionSet := range gs.VersionSets {
		did := versionSet.DidDocs[0].DidDoc.Id

		for _, didDoc := range versionSet.DidDocs {
			if did != didDoc.DidDoc.Id {
				return fmt.Errorf("DID document %s does not belong to version set %s", didDoc.DidDoc.Id, did)
			}
		}
	}

	return nil
}

func (gs GenesisState) ValidateBasic() error {
	for _, versionSet := range gs.VersionSets {
		for _, didDoc := range versionSet.DidDocs {
			err := didDoc.DidDoc.Validate(nil)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	if err := gs.ValidateBasic(); err != nil {
		return err
	}

	if err := gs.ValidateNoDuplicates(); err != nil {
		return err
	}

	if err := gs.ValidateVersionSets(); err != nil {
		return err
	}

	return nil
}
