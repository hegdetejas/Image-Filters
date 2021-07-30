// Package png allows for loading png images and applying
// image flitering effects on them.
package png

import (
	"image/color"
)

// Grayscale applies a grayscale filtering effect to the image
func (img *Image) Grayscale(last bool, start int, end int) {
	bounds := img.Out.Bounds()

	if start == -1 && end == -1 {
		start = bounds.Min.X
		end = bounds.Max.X
	}

	for x := start; x < end; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			r, g, b, a := img.In.At(x, y).RGBA()

			greyC := clamp(float64(r+g+b) / 3)

			img.Out.Set(x, y, color.RGBA64{greyC, greyC, greyC, uint16(a)})
		}
	}
}

func (img *Image) Sharpen(last bool, start int, end int) {
	kernelMatrix := [][]float64{
		{0, -1, 0},
		{-1, 5, -1},
		{0, -1, 0},
	}

	convolute(kernelMatrix, img, last, start, end)
}

func (img *Image) Blur(last bool, start int, end int) {
	kernelMatrix := [][]float64{
		{1 / 9.0, 1 / 9.0, 1 / 9.0},
		{1 / 9.0, 1 / 9.0, 1 / 9.0},
		{1 / 9.0, 1 / 9.0, 1 / 9.0},
	}

	convolute(kernelMatrix, img, last, start, end)
}

func (img *Image) EdgeDetection(last bool, start int, end int) {
	kernelMatrix := [][]float64{
		{-1, -1, -1},
		{-1, 8, -1},
		{-1, -1, -1},
	}

	convolute(kernelMatrix, img, last, start, end)
}

func convolute(kernelMatrix [][]float64, img *Image, last bool, start int, end int) {
	bounds := img.Out.Bounds()

	if start == -1 && end == -1 {
		start = bounds.Min.X
		end = bounds.Max.X
	}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := start; x < end; x++ {
			r, g, b, _ := img.In.At(x-1, y-1).RGBA()
			rLeftUp := float64(r) * kernelMatrix[0][0]
			gLeftUp := float64(g) * kernelMatrix[0][0]
			bLeftUp := float64(b) * kernelMatrix[0][0]

			r, g, b, _ = img.In.At(x, y-1).RGBA()
			rUp := float64(r) * kernelMatrix[1][0]
			gUp := float64(g) * kernelMatrix[1][0]
			bUp := float64(b) * kernelMatrix[1][0]

			r, g, b, _ = img.In.At(x+1, y-1).RGBA()
			rRightUp := float64(r) * kernelMatrix[2][0]
			gRightUp := float64(g) * kernelMatrix[2][0]
			bRightUp := float64(b) * kernelMatrix[2][0]

			r, g, b, _ = img.In.At(x-1, y).RGBA()
			rLeft := float64(r) * kernelMatrix[0][1]
			gLeft := float64(g) * kernelMatrix[0][1]
			bLeft := float64(b) * kernelMatrix[0][1]

			r, g, b, _ = img.In.At(x+1, y).RGBA()
			rRight := float64(r) * kernelMatrix[2][1]
			gRight := float64(g) * kernelMatrix[2][1]
			bRight := float64(b) * kernelMatrix[2][1]

			r, g, b, _ = img.In.At(x-1, y+1).RGBA()
			rLeftDown := float64(r) * kernelMatrix[0][2]
			gLeftDown := float64(g) * kernelMatrix[0][2]
			bLeftDown := float64(b) * kernelMatrix[0][2]

			r, g, b, _ = img.In.At(x, y+1).RGBA()
			rDown := float64(r) * kernelMatrix[1][2]
			gDown := float64(g) * kernelMatrix[1][2]
			bDown := float64(b) * kernelMatrix[1][2]

			r, g, b, _ = img.In.At(x+1, y+1).RGBA()
			rRightDown := float64(r) * kernelMatrix[2][2]
			gRightDown := float64(g) * kernelMatrix[2][2]
			bRightDown := float64(b) * kernelMatrix[2][2]

			r, g, b, a := img.In.At(x, y).RGBA()
			rCentre := float64(r) * kernelMatrix[1][1]
			gCentre := float64(g) * kernelMatrix[1][1]
			bCentre := float64(b) * kernelMatrix[1][1]
			aCentre := clamp(float64(a))

			rNew := clamp(float64(rCentre) + float64(rRightDown) + float64(rDown) + float64(rLeftDown) + float64(rRight) + float64(rLeft) + float64(rRightUp) + float64(rUp) + float64(rLeftUp))
			gNew := clamp(float64(gCentre) + float64(gRightDown) + float64(gDown) + float64(gLeftDown) + float64(gRight) + float64(gLeft) + float64(gRightUp) + float64(gUp) + float64(gLeftUp))
			bNew := clamp(float64(bCentre) + float64(bRightDown) + float64(bDown) + float64(bLeftDown) + float64(bRight) + float64(bLeft) + float64(bRightUp) + float64(bUp) + float64(bLeftUp))

			img.Out.Set(x, y, color.RGBA64{rNew, gNew, bNew, uint16(aCentre)})
		}
	}
}
