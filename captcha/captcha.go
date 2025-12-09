package captcha

import (
	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/wenlng/go-captcha-assets/resources/imagesv2"
	"github.com/wenlng/go-captcha-assets/resources/tiles"
	"github.com/wenlng/go-captcha/v2/slide"
)

const Name = "captcha_component"

type Component struct {
	captcha slide.Captcha
}

func (c *Component) Init(config *config2.Config) error {
	builder := slide.NewBuilder()
	images, err := imagesv2.GetImages()
	if err != nil {
		return err
	}
	graphs, err := tiles.GetTiles()
	if err != nil {
		return err
	}
	graph2s := make([]*slide.GraphImage, len(graphs))
	for i, graph := range graphs {
		graph2s[i] = &slide.GraphImage{
			MaskImage:    graph.MaskImage,
			OverlayImage: graph.OverlayImage,
			ShadowImage:  graph.ShadowImage,
		}
	}
	builder.SetResources(slide.WithBackgrounds(images), slide.WithGraphImages(graph2s))
	c.captcha = builder.Make()
	return nil
}
func (c *Component) GetCaptcha() slide.Captcha {
	return c.captcha
}
func (c *Component) Name() string {
	return Name
}
