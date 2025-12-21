package captcha

import (
	"encoding/json"

	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/util"
	"github.com/wenlng/go-captcha-assets/resources/imagesv2"
	"github.com/wenlng/go-captcha-assets/resources/tiles"
	"github.com/wenlng/go-captcha/v2/slide"
)

const Name = "captcha_component"

type Component struct {
	captcha slide.Captcha
	key     string
	iv      string
}
type SlideCaptchaData struct {
	TileImage   string `json:"tileImage"`
	MasterImage string `json:"masterImage"`
	ThumbX      int    `json:"thumbX"`
	ThumbCode   string `json:"thumbCode"`
	ThumbY      int    `json:"thumbY"`
	ThumbWidth  int    `json:"thumbWidth"`
	ThumbHeight int    `json:"thumbHeight"`
	ThumbAngle  int    `json:"thumbAngle"`
}

type Config struct {
	CodeKey string
	CodeIv  string
}

func (c *Config) Key() string {
	return "captcha"
}

type SlideCaptcha struct {
}

func (c *Component) Init(config config2.IConfig) error {
	var cfg Config
	err := config.Unmarshal(cfg.Key(), &cfg)
	if err != nil {
		return err
	}
	c.key = util.SubStringAndPadSpace(cfg.CodeKey, 32)
	c.iv = util.SubStringAndPadSpace(cfg.CodeIv, 16)
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
func (c *Component) GetCaptchaData() (*SlideCaptchaData, error) {
	captchaData, err := c.captcha.Generate()
	if err != nil {
		return nil, err
	}
	tile, err := captchaData.GetTileImage().ToBase64()
	if err != nil {
		return nil, err
	}
	master, err := captchaData.GetMasterImage().ToBase64()
	if err != nil {
		return nil, err
	}
	block := captchaData.GetData()
	data := util.OfMap2("time", util.NowDateTime(), "thumbX", block.X)
	js, _ := json.Marshal(data)
	v := util.EncryptByCBC(string(js), c.key, c.iv)
	return &SlideCaptchaData{
		TileImage:   tile,
		MasterImage: master,
		ThumbX:      0,
		ThumbY:      block.Y,
		ThumbWidth:  block.Width,
		ThumbHeight: block.Height,
		ThumbAngle:  block.Angle,
		ThumbCode:   v,
	}, nil
}
func (c *Component) Name() string {
	return Name
}
