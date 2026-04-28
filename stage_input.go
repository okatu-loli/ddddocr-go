package ddddocr

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	_ "golang.org/x/image/bmp"
)

// ReadInputFile reads an image file into memory.
func ReadInputFile(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("empty image path")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return data, nil
}

// DecodeImage decodes image bytes into an image.Image.
func DecodeImage(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	return img, nil
}

// LoadImageFile reads and decodes an image file.
func LoadImageFile(path string) (image.Image, error) {
	data, err := ReadInputFile(path)
	if err != nil {
		return nil, err
	}
	return DecodeImage(data)
}
