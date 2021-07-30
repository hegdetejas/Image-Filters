// Package png allows for loading png images and applying
// image flitering effects on them
package png

import (
	"image"
	"image/draw"
	"image/png"
	"math"
	"os"
)

// The Image represents a structure for working with PNG images.
type Image struct {
	In   *image.RGBA64
	Out  *image.RGBA64
	Temp *image.RGBA64
}

//
// Public functions
//

func NewImage(img *image.RGBA64) *Image {
	return &Image{In: img, Out: img}
}

// Load returns a Image that was loaded based on the filePath parameter
func Load(filePath string) (*Image, error) {

	inReader, err := os.Open(filePath)

	if err != nil {
		return nil, err
	}
	defer inReader.Close()

	inImgDecoded, err := png.Decode(inReader)

	if err != nil {
		return nil, err
	}

	inBounds := inImgDecoded.Bounds()

	inImg := image.NewRGBA64(image.Rect(0, 0, inBounds.Dx(), inBounds.Dy()))
	draw.Draw(inImg, inImg.Bounds(), inImgDecoded, inBounds.Min, draw.Src)
	outImg := image.NewRGBA64(inBounds)
	tempImg := image.NewRGBA64(inBounds)

	return &Image{inImg, outImg, tempImg}, nil
}

// Save saves the image to the given file
func (img *Image) Save(filePath string) error {
	outWriter, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outWriter.Close()

	err = png.Encode(outWriter, img.Out)
	if err != nil {
		return err
	}
	return nil
}

//clamp will clamp the comp parameter to zero if it is less than zero or to 65535 if the comp parameter
// is greater than 65535.
func clamp(comp float64) uint16 {
	return uint16(math.Min(65535, math.Max(0, comp)))
}

func (img *Image) FindBounds() []int {
	bounds := img.Out.Bounds()
	boundsList := make([]int, 0)
	boundsList = append(boundsList, bounds.Min.X, bounds.Max.X, bounds.Min.Y, bounds.Max.Y)

	return boundsList
}
