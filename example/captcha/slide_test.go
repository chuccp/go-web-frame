package main

import (
	"encoding/json"
	"testing"

	"github.com/wenlng/go-captcha-assets/resources/imagesv2"
	"github.com/wenlng/go-captcha-assets/resources/tiles"
	"github.com/wenlng/go-captcha/v2/slide"
)

func TestName(t *testing.T) {
	builder := slide.NewBuilder()
	imgs, _ := imagesv2.GetImages()
	graphs, _ := tiles.GetTiles()
	graph2s := make([]*slide.GraphImage, len(graphs))
	for i, graph := range graphs {
		graph2s[i] = &slide.GraphImage{
			MaskImage:    graph.MaskImage,
			OverlayImage: graph.OverlayImage,
			ShadowImage:  graph.ShadowImage,
		}
	}
	builder.SetResources(slide.WithBackgrounds(imgs), slide.WithGraphImages(graph2s))
	slideTileCapt := builder.Make()
	captData, _ := slideTileCapt.Generate()
	blockData := captData.GetData()
	block, _ := json.Marshal(blockData)
	t.Log(string(block))
}
