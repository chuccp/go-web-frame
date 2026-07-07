package qrcode

import (
	"image/color"
	"io"

	qrcode "github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	"go.uber.org/zap/buffer"
)

// ============================================================
// 通用二维码生成
// ============================================================

func GenerateQrcode(content string, writeCloser io.WriteCloser, opts ...standard.ImageOption) error {
	qrc, err := qrcode.New(content)
	if err != nil {
		return err
	}
	w := standard.NewWithWriter(writeCloser, opts...)
	if s := NewStripeQRCode(); s != nil {
		err = qrc.Save(&stripeWriter{inner: w, shape: s})
	} else {
		err = qrc.Save(w)
	}
	if err != nil {
		return err
	}
	return nil
}

// ============================================================
// ShapeRoundedSquare: 圆角方块风格
// ============================================================

type ShapeRoundedSquare struct {
	Color     color.Color
	dim       int
	borderMod int
}

func IsWhite(c color.Color) bool {
	r, g, b, a := c.RGBA()
	// alpha 为 0 视为空白（透明），RGB 全满为纯白
	return a == 0 || (r == 0xffff && g == 0xffff && b == 0xffff)
}

func (s *ShapeRoundedSquare) Draw(ctx *standard.DrawContext) {
	w, h := ctx.Edge()
	fw0, fh0 := float64(w), float64(h)
	x0, y0 := ctx.UpperLeft()

	if s.Color != nil {
		if IsWhite(ctx.Color()) {
			ctx.SetColor(ctx.Color())
		} else {
			ctx.SetColor(s.Color)
		}
	} else {
		ctx.SetColor(ctx.Color())
	}

	ctx.DrawRoundedRectangle(x0+1, y0+1, fw0-2, fh0-2, fw0/3)
	ctx.Fill()
}

func (s *ShapeRoundedSquare) DrawFinder(ctx *standard.DrawContext) {
	s.ensureDim(ctx)
	drawFinderWhole(ctx, ctx.Color(), color.RGBA{R: 255, G: 255, B: 255, A: 255}, s.dim, s.borderMod)
}

func (s *ShapeRoundedSquare) ensureDim(ctx *standard.DrawContext) {
	if s.dim > 0 {
		return
	}
	w, _ := ctx.Edge()
	x0, _ := ctx.UpperLeft()
	s.borderMod = int(x0) / w
	imgW := ctx.Width()
	s.dim = (imgW - 2*s.borderMod*w) / w
	if s.dim <= 0 {
		s.dim = 25
	}
}

func WithRoundedSquareShape() standard.ImageOption {
	return standard.WithCustomShape(&ShapeRoundedSquare{})
}

// ============================================================
// ShapeCircle: 圆形风格
// ============================================================

type ShapeCircle struct {
	Color     color.Color
	dim       int
	borderMod int
}

func (s *ShapeCircle) Draw(ctx *standard.DrawContext) {
	w, h := ctx.Edge()
	fw0, fh0 := float64(w), float64(h)
	x0, y0 := ctx.UpperLeft()

	if s.Color != nil {
		if IsWhite(ctx.Color()) {
			ctx.SetColor(ctx.Color())
		} else {
			ctx.SetColor(s.Color)
		}
	} else {
		ctx.SetColor(ctx.Color())
	}

	radius := fw0 / 2
	if fh0/2 < radius {
		radius = fh0 / 2
	}
	cx := float64(x0) + fw0/2
	cy := float64(y0) + fh0/2
	ctx.DrawCircle(cx, cy, radius-1)
	ctx.Fill()
}

func (s *ShapeCircle) DrawFinder(ctx *standard.DrawContext) {
	s.ensureDim(ctx)
	drawFinderWhole(ctx, ctx.Color(), color.RGBA{R: 255, G: 255, B: 255, A: 255}, s.dim, s.borderMod)
}

func (s *ShapeCircle) ensureDim(ctx *standard.DrawContext) {
	if s.dim > 0 {
		return
	}
	w, _ := ctx.Edge()
	x0, _ := ctx.UpperLeft()
	s.borderMod = int(x0) / w
	imgW := ctx.Width()
	s.dim = (imgW - 2*s.borderMod*w) / w
	if s.dim <= 0 {
		s.dim = 25
	}
}

// drawFinderWhole 7×7 定位图案整体绘制。dim 为矩阵维度，borderMod 为边框模块数，用于确定各定位图案的左上角坐标。
func drawFinderWhole(ctx *standard.DrawContext, dark, bg color.Color, dim, borderMod int) {
	if dark == nil {
		dark = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	}
	if bg == nil {
		bg = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	w, _ := ctx.Edge()
	fw := float64(w)
	x0, y0 := ctx.UpperLeft()

	col := int(x0) / w
	row := int(y0) / w
	mr := row - borderMod // 矩阵坐标
	mc := col - borderMod

	// 确定本格所属定位图案的左上角矩阵坐标
	var fx, fy int
	if mr < 7 && mc < 7 {
		fx, fy = 0, 0 // 左上
	} else if mr < 7 && mc >= dim-7 {
		fx, fy = dim-7, 0 // 右上
	} else if mr >= dim-7 && mc < 7 {
		fx, fy = 0, dim-7 // 左下
	} else {
		return // 非定位图案角格
	}

	// 仅该定位图案的左上角格触发绘制
	if mr != fy || mc != fx {
		return
	}

	fx0 := x0 - float64(mc-fx)*fw
	fy0 := y0 - float64(mr-fy)*fw
	fsize := fw * 7

	ctx.SetColor(dark)
	ctx.DrawRoundedRectangle(fx0, fy0, fsize, fsize, fw)
	ctx.Fill()

	ctx.SetColor(bg)
	ctx.DrawRoundedRectangle(fx0+fw, fy0+fw, fw*5, fw*5, fw)
	ctx.Fill()

	ctx.SetColor(dark)
	ctx.DrawRoundedRectangle(fx0+fw*2, fy0+fw*2, fw*3, fw*3, fw)
	ctx.Fill()
}

func WithCircleShape() standard.ImageOption {
	return standard.WithCustomShape(&ShapeCircle{})
}

// ============================================================
// BufferWriteCloser: 写入 buffer.Buffer 的便捷 Writer
// ============================================================

type BufferWriteCloser struct {
	b *buffer.Buffer
}

func (w *BufferWriteCloser) Write(p []byte) (n int, err error) {
	return w.b.Write(p)
}
func (w *BufferWriteCloser) Close() error {
	return nil
}

func (w *BufferWriteCloser) Bytes() []byte {
	return w.b.Bytes()
}

func CreateBufferWriteCloser() *BufferWriteCloser {
	return &BufferWriteCloser{
		b: new(buffer.Buffer),
	}
}

// GenerateStripeQRCode 便捷函数：用默认样式直接生成条纹二维码文件
func GenerateStripeQRCode(content, filePath string) error {
	return NewStripeQRCode().GenerateFile(content, filePath)
}
