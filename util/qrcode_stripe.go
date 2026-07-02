package util

import (
	"image/color"

	qrcode "github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

// StripeQRCode 条纹胶囊风格二维码
type StripeQRCode struct {
	moduleSize int
	dotRadius  float64
	bgColor    color.RGBA
	darkColor  color.RGBA
}

func NewStripeQRCode() *StripeQRCode {
	return &StripeQRCode{
		moduleSize: 10,
		dotRadius:  4.8,
		bgColor:    color.RGBA{R: 255, G: 255, B: 255, A: 255},
		darkColor:  color.RGBA{R: 0, G: 0, B: 0, A: 255},
	}
}

func (s *StripeQRCode) WithModuleSize(size int) *StripeQRCode {
	s.moduleSize = size
	s.dotRadius = float64(size) * 0.46
	return s
}

// SetDotRadius 直接设置圆半径
func (s *StripeQRCode) SetDotRadius(r float64) *StripeQRCode {
	s.dotRadius = r
	return s
}
func (s *StripeQRCode) WithColors(bg, dark, _ color.RGBA) *StripeQRCode {
	s.bgColor = bg
	s.darkColor = dark
	return s
}

func (s *StripeQRCode) GenerateFile(content, filePath string) error {
	qr, err := qrcode.New(content)
	if err != nil {
		return err
	}
	w, err := standard.New(filePath, standard.WithCustomShape(s))
	if err != nil {
		return err
	}
	return qr.Save(w)
}

// Draw 实现 IShape 接口
func (s *StripeQRCode) Draw(ctx *standard.DrawContext) {
	if IsWhite(ctx.Color()) {
		return
	}
	w, _ := ctx.Edge()
	x0, y0 := ctx.UpperLeft()
	fw := float64(w)
	cx := float64(x0) + fw/2
	cy := float64(y0) + fw/2
	r := s.dotRadius
	mod := float64(w)
	nei := ctx.Neighbours()

	col := int(x0) / w
	row := int(y0) / w

	hasH := nei&(standard.NLeft|standard.NRight) != 0
	hasV := nei&(standard.NTop|standard.NBot) != 0

	ctx.SetColor(s.darkColor)
	ctx.DrawCircle(cx, cy, r)

	// 第一步：横条
	drawH := hasH && nei&standard.NRight != 0 &&
		(nei&standard.NLeft == 0 || col%3 != 2)

	// 第二步：竖条（原始：无横邻 + 无对角 + %3）
	drawV := !hasH && hasV && nei&standard.NBot != 0 &&
		nei&(standard.NBotLeft|standard.NBotRight) == 0 &&
		(nei&standard.NTop == 0 || row%3 != 2)

	// 第三步：相交（横竖都要 → 横赢），相接（NLeft → 竖让位）
	// （已由 !hasH 和 if/else 结构处理）

	if drawH {
		ctx.DrawRectangle(cx, cy-r, mod, r*2)
	} else if drawV {
		ctx.DrawRectangle(cx-r, cy, r*2, mod)
	} else if nei&standard.NLeft == 0 &&
		nei&standard.NBot != 0 &&
		nei&(standard.NBotLeft|standard.NBotRight) == 0 &&
		(nei&standard.NTop == 0 || row%3 != 2) {
		// 第四步：独立圆点连接——本格只画圆，下方有暗色格，无任何横条冲突
		ctx.DrawRectangle(cx-r, cy, r*2, mod)
	}

	ctx.Fill()
}

// DrawFinder 实现 IShape 接口
func (s *StripeQRCode) DrawFinder(ctx *standard.DrawContext) {
	w, h := ctx.Edge()
	fw0, fh0 := float64(w), float64(h)
	x0, y0 := ctx.UpperLeft()
	if s.darkColor != (color.RGBA{}) {
		if IsWhite(ctx.Color()) {
			ctx.SetColor(ctx.Color())
		} else {
			ctx.SetColor(s.darkColor)
		}
	} else {
		ctx.SetColor(ctx.Color())
	}
	ctx.DrawRectangle(x0, y0, fw0, fh0)
	ctx.Fill()
}

func WithStripeShape() standard.ImageOption {
	return standard.WithCustomShape(NewStripeQRCode())
}

func WithCustomStripeShape(dark color.RGBA) standard.ImageOption {
	return standard.WithCustomShape(
		NewStripeQRCode().WithColors(
			color.RGBA{R: 255, G: 255, B: 255, A: 255},
			dark,
			color.RGBA{},
		),
	)
}
