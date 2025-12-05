package util

import (
	"io"

	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

func GenerateQrcode(content string, writeCloser io.WriteCloser, opts ...standard.ImageOption) error {

	qrc, err := qrcode.New(content)
	if err != nil {
		return err
	}
	err = qrc.Save(standard.NewWithWriter(writeCloser, opts...))
	if err != nil {
		return err
	}
	return nil
}

type ShapeRoundedSquare struct {
}

func (s *ShapeRoundedSquare) Draw(ctx *standard.DrawContext) {
	w, h := ctx.Edge()
	fw0, fh0 := float64(w), float64(h)
	x0, y0 := ctx.UpperLeft()
	ctx.SetColor(ctx.Color())
	ctx.DrawRoundedRectangle(x0+1, y0+1, fw0-2, fh0-2, fw0/3)
	ctx.Fill()
}

func (s *ShapeRoundedSquare) DrawFinder(ctx *standard.DrawContext) {
	w, h := ctx.Edge()
	fw0, fh0 := float64(w), float64(h)
	x0, y0 := ctx.UpperLeft()
	ctx.SetColor(ctx.Color())
	ctx.DrawRectangle(x0, y0, fw0, fh0)
	ctx.Fill()
}

func WithRoundedSquareShape() standard.ImageOption {
	return standard.WithCustomShape(&ShapeRoundedSquare{})

}
