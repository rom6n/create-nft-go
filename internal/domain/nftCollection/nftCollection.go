package nftcollection

import "github.com/google/uuid"

type NftCollectionMetadata struct {
}

type NftCollection struct {
	Address       string                `bson:"_id" json:"address"`
	NextItemIndex int64                 `bson:"next_item_index" json:"next_item_index"`
	Owner         uuid.UUID             `bson:"owner" json:"owner"`
	Metadata      NftCollectionMetadata `bson:"metadata" json:"metadata"` // под вопросом как метадата будет приходить
}

type MintCollectionCfg struct {
	Owner             string
	CommonContent     string
	CollectionContent string
	RoyaltyDividend   uint16
	RoyaltyDivisor    uint16
	// next item index always is 1
	// nft item code
}
