package nft

import "github.com/google/uuid"

type Attribute struct {
	TraitType string `bson:"trait_type" json:"trait_type"`
	Value     string `bson:"value" json:"value"`
}

type NftItemMetadata struct {
	Name        string      `bson:"name" json:"name"`
	Image       string      `bson:"image" json:"image"`
	Attributes  []Attribute `bson:"attributes" json:"attributes"`
	Description string      `bson:"description" json:"description"`
	ExternalUrl string      `bson:"external_url" json:"external_url"`
}

type NftItem struct {
	Address           string          `bson:"_id" json:"address"`
	Index             int64           `bson:"index" json:"index"`
	CollectionAddress string          `bson:"collection_address" json:"collection_address"`
	CollectionName    string          `bson:"collection_name" json:"collection_name"`
	Owner             uuid.UUID          `bson:"owner" json:"owner"`
	Metadata          NftItemMetadata `bson:"metadata" json:"metadata"` // под вопросом как метадата будет приходить
}
