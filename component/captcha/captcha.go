package captcha

import (
	"encoding/json"
	"math"

	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/spf13/cast"
	"github.com/wenlng/go-captcha-assets/resources/imagesv2"
	"github.com/wenlng/go-captcha-assets/resources/tiles"
	"github.com/wenlng/go-captcha/v2/slide"
	"go.uber.org/zap"
)

const Name = "captcha_component"

type Captcha struct {
	captcha slide.Captcha
	key     string
	iv      string
}
type SlideCaptchaData struct {
	Type        string `json:"type"`
	TileImage   string `json:"tileImage"`
	MasterImage string `json:"masterImage"`
	ThumbX      int    `json:"thumbX"`
	ThumbCode   string `json:"thumbCode"`
	ThumbY      int    `json:"thumbY"`
	ThumbWidth  int    `json:"thumbWidth"`
	ThumbHeight int    `json:"thumbHeight"`
	ThumbAngle  int    `json:"thumbAngle"`
}
type Data struct {
	Type        string `json:"type"`
	CaptchaCode string `json:"captchaCode"`
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

func (c *Captcha) Init(config config2.IConfig) error {
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
func (c *Captcha) GetCaptcha() slide.Captcha {
	return c.captcha
}
func (c *Captcha) GetCaptchaData() (*SlideCaptchaData, error) {
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

	v, err := c.generateCode(cast.ToString(block.X))
	if err != nil {
		return nil, err
	}
	return &SlideCaptchaData{
		Type:        "slide",
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
func (c *Captcha) generateCode(value string) (string, error) {
	data := util.OfMap2("time", util.NowDateFormatTime(util.TimestampFormat), "thumbX", value)
	js, _ := json.Marshal(data)
	return util.EncryptByCBC(string(js), c.key, c.iv)
}
func (c *Captcha) ValidateThumb(code string, x string) (*Data, bool) {
	v, err := util.DecryptByCBC(code, c.key, c.iv)
	if err != nil {
		log.Errors("ValidateThumb", err)
		return nil, false
	}
	var data map[string]interface{}
	err = json.Unmarshal([]byte(v), &data)
	if err != nil {
		return nil, false
	}
	thumbX := cast.ToString(data["thumbX"])
	x0 := cast.ToInt(thumbX)
	x1 := cast.ToInt(x)
	n := x0 - x1
	log.Debug("ValidateThumb", zap.String("thumbX", thumbX), zap.String("x", x), zap.Int("n", n))
	if math.Abs(float64(n)) < 3 {
		v, err := c.generateCode(util.CRC(6, c.key[:6]))
		if err != nil {
			log.Errors("ValidateThumb", err)
			return nil, false
		}
		return &Data{CaptchaCode: v, Type: "code"}, true
	}
	return nil, false
}
func (c *Captcha) ValidateCode(code string) bool {
	v, err := util.DecryptByCBC(code, c.key, c.iv)
	if err != nil {
		log.Errors("ValidateCode", err)
		return false
	}
	var data map[string]interface{}
	err = json.Unmarshal([]byte(v), &data)
	if err != nil {
		return false
	}
	time := cast.ToString(data["time"])
	if util.IsBlank(time) {
		return false
	}
	return util.IsAfter(time, util.GetNowTime(), util.TimestampFormat)

}
func (c *Captcha) Destroy() error {
	return nil
}
