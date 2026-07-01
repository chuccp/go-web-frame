package util

import (
	"image/color"

	qrcode "github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

// StripeQRCode 条纹风格二维码（胶囊短条，横竖不相交）
//
// 实现 standard.IShape 接口，通过 WithStripeShape() 传入 GenerateQrcode 使用。
// 视觉特征：深色模块 = 粗圆点；相邻深色圆由粗矩连成胶囊短条；
// 每 3 个模块为一组，尽量 3 个相连；横竖不相交（横优先）。
type StripeQRCode struct {
	moduleSize int
	dotRadius  float64 // 圆点半径
	bgColor    color.RGBA
	darkColor  color.RGBA
}

func NewStripeQRCode() *StripeQRCode {
	return &StripeQRCode{
		moduleSize: 10,
		dotRadius:  3.8,
		bgColor:    color.RGBA{R: 255, G: 255, B: 255, A: 255},
		darkColor:  color.RGBA{R: 0, G: 0, B: 0, A: 255},
	}
}

func (s *StripeQRCode) WithModuleSize(size int) *StripeQRCode {
	s.moduleSize = size
	s.dotRadius = float64(size) * 0.38
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
// 每模块画粗圆点 + 中心到中心的完整连接条 → 清晰胶囊。
// 横优先：十字路口只画横、竖不画。
// 竖条通过对角邻居检测下邻/上邻是否横穿 → 是则截断到边界，横竖永不交叠。
// %3 分组：每 3 个为一组，组尾不连，尽量 3 个相连。
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
	nei := ctx.Neighbours()

	hasH := nei&(standard.NLeft|standard.NRight) != 0
	hasV := nei&(standard.NTop|standard.NBot) != 0

	col := int(x0) / w
	row := int(y0) / w

	// 中心圆
	s.drawCircle(ctx, cx, cy, r, s.darkColor)

	switch {
	case hasH && hasV:
		// 十字路口：横优先，只画横（中心到中心），竖不画
		if nei&standard.NRight != 0 && col%3 != 2 {
			s.drawBarH(ctx, cx, cy-r, fw, r*2)
		}
		if nei&standard.NLeft != 0 && col%3 != 0 {
			s.drawBarH(ctx, cx-fw, cy-r, fw, r*2)
		}
	case hasH:
		// 纯横向：中心到中心，完整胶囊
		if nei&standard.NRight != 0 && col%3 != 2 {
			s.drawBarH(ctx, cx, cy-r, fw, r*2)
		}
		if nei&standard.NLeft != 0 && col%3 != 0 {
			s.drawBarH(ctx, cx-fw, cy-r, fw, r*2)
		}
	case hasV:
		// 纯纵向：检测对角邻判断下/上邻是否是横穿路口
		if nei&standard.NBot != 0 && row%3 != 2 {
			if nei&(standard.NBotLeft|standard.NBotRight) != 0 {
				// 下邻是横穿 → 截断到边界
				s.drawBarV(ctx, cx-r, cy+r, r*2, fw/2-r)
			} else {
				s.drawBarV(ctx, cx-r, cy, r*2, fw)
			}
		}
		if nei&standard.NTop != 0 && row%3 != 0 {
			if nei&(standard.NTopLeft|standard.NTopRight) != 0 {
				s.drawBarV(ctx, cx-r, float64(y0), r*2, fw/2-r)
			} else {
				s.drawBarV(ctx, cx-r, cy-fw, r*2, fw)
			}
		}
	}
}

func (s *StripeQRCode) drawBarH(ctx *standard.DrawContext, x, y, w, h float64) {
	ctx.SetColor(s.darkColor)
	ctx.DrawRectangle(x, y, w, h)
	ctx.Fill()
}

func (s *StripeQRCode) drawBarV(ctx *standard.DrawContext, x, y, w, h float64) {
	ctx.SetColor(s.darkColor)
	ctx.DrawRectangle(x, y, w, h)
	ctx.Fill()
}

// DrawFinder 实现 IShape 接口 — 标准纯色矩形（与 ShapeRoundedSquare.DrawFinder 一致）
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

// drawCircle 填充圆
func (s *StripeQRCode) drawCircle(ctx *standard.DrawContext, cx, cy, r float64, c color.Color) {
	ctx.SetColor(c)
	ctx.DrawCircle(cx, cy, r)
	ctx.Fill()
}

// WithStripeShape 默认黑白条纹
func WithStripeShape() standard.ImageOption {
	return standard.WithCustomShape(NewStripeQRCode())
}

// WithCustomStripeShape 自定义深色颜色的条纹（背景白色）
func WithCustomStripeShape(dark color.RGBA) standard.ImageOption {
	return standard.WithCustomShape(
		NewStripeQRCode().WithColors(
			color.RGBA{R: 255, G: 255, B: 255, A: 255},
			dark,
			color.RGBA{},
		),
	)
}
