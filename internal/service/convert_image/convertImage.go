package convertimage

import (
	"bytes"
	"fmt"
	"image"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
)

type ConvertImageRepository interface {
	ConvertImageWebP(image []byte) ([]byte, error)
}

type ConvertImageRepo struct{}

func New() ConvertImageRepository {
	return &ConvertImageRepo{}
}

func (v *ConvertImageRepo) ConvertImageWebP(img []byte) ([]byte, error) {
	decodedImg, _, decodeErr := image.Decode(bytes.NewReader(img))
	if decodeErr != nil {
		return nil, fmt.Errorf("error decoding image: %v", decodeErr)
	}

	squareImg := imaging.CropCenter(decodedImg, 1000, 1000)

	imgByte, encodeErr := webp.EncodeRGBA(squareImg, 90)
	if encodeErr != nil {
		fmt.Errorf("error converting image to webp: %v", encodeErr)
	}

	return imgByte, nil
}