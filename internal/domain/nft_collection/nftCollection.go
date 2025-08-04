package nftcollection

import (
	"github.com/google/uuid"
	"github.com/xssnick/tonutils-go/address"
)

type NftCollectionMetadata struct {
	Name         string   `bson:"name" json:"name"`
	Image        string   `bson:"image" json:"image"`
	CoverImage   string   `bson:"cover_image" json:"cover_image"`
	Description  string   `bson:"description" json:"description"`
	ExternalUrl  string   `bson:"external_url" json:"external_url"`
	ExternalLink string   `bson:"external_link" json:"external_link"`
	SocialLinks  []string `bson:"social_links" json:"social_links"`
	Marketplace  string   `bson:"marketplace" json:"marketplace"`
}

type NftCollection struct {
	Address       string                `bson:"_id" json:"address"`
	NextItemIndex int64                 `bson:"next_item_index" json:"next_item_index"`
	Owner         uuid.UUID             `bson:"owner" json:"owner"`
	Metadata      NftCollectionMetadata `bson:"metadata" json:"metadata"` // под вопросом как метадата будет приходить
	IsTestnet     bool                  `bson:"is_testnet" json:"is_testnet"`
}

type DeployCollectionCfg struct {
	OwnerAddress      *address.Address
	CommonContent     string
	CollectionContent string
	RoyaltyDividend   uint16
	RoyaltyDivisor    uint16
	// next item index always is 1
	// nft item code
}

func New(address string, ownerUuid uuid.UUID, metadata *NftCollectionMetadata, isTestnet bool) *NftCollection {
	return &NftCollection{
		Address:       address,
		NextItemIndex: 1,
		Owner:         ownerUuid,
		Metadata:      *metadata,
		IsTestnet:     isTestnet,
	}
}
