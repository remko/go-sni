package sni

import "image"

type Pixmap struct {
	Width  int
	Height int
	Data   []byte
}

func ImagePixmap(img image.Image) Pixmap {
	iconpm := Pixmap{
		Width:  img.Bounds().Dx(),
		Height: img.Bounds().Dy(),
	}
	data := make([]byte, 0, iconpm.Width*iconpm.Height*4)
	for y := 0; y < iconpm.Height; y++ {
		for x := 0; x < iconpm.Width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			data = append(data, byte(a>>8))
			data = append(data, byte(r>>8))
			data = append(data, byte(g>>8))
			data = append(data, byte(b>>8))
		}
	}
	iconpm.Data = data
	return iconpm
}
