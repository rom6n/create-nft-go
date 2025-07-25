package jsonx

import (
	"fmt"
	"log"

	"github.com/go-faster/jx"
	"github.com/rom6n/create-nft-go/internal/domain/wallet"
	"github.com/tonkeeper/tonapi-go"
)

type DecodeJxError struct {
	errorStr string
}

func (c *DecodeJxError) Error() string {
	return c.errorStr
}

func DecodeAndPackNftItemMetadata(metadata tonapi.NftItemMetadata) (wallet.NftItemMetadata, error) {
	metadataName, nameErr := jx.DecodeBytes(metadata["name"]).Str()
	metadataImage, imageErr := jx.DecodeBytes(metadata["image"]).Str()
	metadataDescription, descriptionErr := jx.DecodeBytes(metadata["description"]).Str()
	metadataExternalUrl, urlErr := jx.DecodeBytes(metadata["external_url"]).Str()

	var attributes []wallet.Attribute

	attributesErr := jx.DecodeBytes(metadata["attributes"]).Arr(
		func(d *jx.Decoder) error {
			attribute := wallet.Attribute{}
			err := d.Obj(func(d *jx.Decoder, key string) error {
				switch key {
				case "trait_type":
					traitType, err := d.Str()
					if err != nil {
						return fmt.Errorf("error decoding trait_type of nft metadata: %w", err)
					}
					attribute.TraitType = traitType
				case "value":
					value, err := d.Str()
					if err != nil {
						return fmt.Errorf("error decoding value of nft metadata: %w", err)
					}
					attribute.Value = value
				default:
					return fmt.Errorf("unexpected key in nft metadata attribute: %s", key)
				}

				return nil
			})

			if err != nil {
				return err
			}

			attributes = append(attributes, attribute)
			return nil
		})

	if nameErr != nil || imageErr != nil || descriptionErr != nil || attributesErr != nil && fmt.Sprint(attributesErr) != "unexpected EOF" || urlErr != nil && fmt.Sprint(urlErr) != "unexpected EOF" {
		log.Printf("‼️Error decoding NFTs metadata:\nName: %v\nImage: %v\nDesc: %v\nURL: %v\n", nameErr, imageErr, descriptionErr, urlErr)
		return wallet.NftItemMetadata{}, &DecodeJxError{errorStr: fmt.Sprintf("Error decoding NFTs metadata:\nName: %v\nImage: %v\nDesc: %v\nURL: %v\nAttributes: %v\n", nameErr, imageErr, descriptionErr, urlErr, attributesErr)}
	}

	nftMetadata := wallet.NftItemMetadata{
		Name:        metadataName,
		Image:       metadataImage,
		Attributes:  attributes,
		Description: metadataDescription,
		ExternalUrl: metadataExternalUrl,
	}

	return nftMetadata, nil
}
