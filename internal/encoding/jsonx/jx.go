package jsonx

import (
	"fmt"
	"log"

	"github.com/go-faster/jx"
	"github.com/tonkeeper/tonapi-go"
)

type DecodeJxError struct {
	errorStr string
}

func (c *DecodeJxError) Error() string {
	return c.errorStr
}

func DecodeAndPackNftItemMetadata(metadata tonapi.NftItemMetadata) (map[string]any, error) {
	metadataName, nameErr := jx.DecodeBytes(metadata["name"]).Str()
	metadataImage, imageErr := jx.DecodeBytes(metadata["image"]).Str()
	metadataDescription, descriptionErr := jx.DecodeBytes(metadata["description"]).Str()
	metadataExternalUrl, urlErr := jx.DecodeBytes(metadata["external_url"]).Str()

	if nameErr != nil || imageErr != nil || descriptionErr != nil || urlErr != nil && fmt.Sprint(urlErr) != "unexpected EOF" {
		log.Printf("‼️Error decoding NFTs metadata:\nName: %v\nImage: %v\nDesc: %v\nURL: %v\n", nameErr, imageErr, descriptionErr, urlErr)
		return nil, &DecodeJxError{errorStr: fmt.Sprintf("Error decoding NFTs metadata:\nName: %v\nImage: %v\nDesc: %v\nURL: %v\n", nameErr, imageErr, descriptionErr, urlErr)}

	}

	clearMetadata := make(map[string]any)
	clearMetadata["name"] = metadataName
	clearMetadata["image"] = metadataImage
	clearMetadata["description"] = metadataDescription
	clearMetadata["external_url"] = metadataExternalUrl

	return clearMetadata, nil
}
