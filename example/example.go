package main

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"
	"log"
	"time"

	"github.com/remko/go-sni"
)

// "github.com/remko/go-sni"

var (
	//go:embed icon.png
	icon_png []byte

	//go:embed icon-attention.png
	attentionicon_png []byte
)

func main() {
	icon, _, err := image.Decode(bytes.NewReader(icon_png))
	if err != nil {
		panic(err)
	}
	attentionicon, _, err := image.Decode(bytes.NewReader(attentionicon_png))
	if err != nil {
		panic(err)
	}

	item, err := sni.NewItem(sni.ItemConfig{
		Category:      "ApplicationStatus",
		ID:            "GoSNIExample",
		Title:         "Go SNI Example",
		Status:        "Active",
		Icon:          sni.Icon{Name: "go-sni-example-icon", Pixmaps: []sni.Pixmap{sni.ImagePixmap(icon)}},
		AttentionIcon: sni.Icon{Name: "go-sni-example-icon-attention", Pixmaps: []sni.Pixmap{sni.ImagePixmap(attentionicon)}},
	})
	if err != nil {
		panic(err)
	}
	defer item.Close()

	item.Activate = func(x, y int) {
		log.Printf("Item was clicked at %d, %d", x, y)
	}
	item.SecondaryActivate = func(x, y int) {
		log.Printf("Item was middle-clicked at %d, %d", x, y)
	}
	item.ContextMenu = func(x, y int) {
		log.Printf("Item was right-clicked at %d, %d", x, y)
	}
	item.Scroll = func(delta int, direction string) {
		log.Printf("Item was scrolled with delta %d, direction %s", delta, direction)
	}

	attention := false
	for {
		time.Sleep(3 * time.Second)
		attention = !attention
		if attention {
			item.SetStatus("NeedsAttention")
		} else {
			item.SetStatus("Active")
		}
	}
}
