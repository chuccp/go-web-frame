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

	// 预扫描方向矩阵：0=圆 1=横 2=竖，索引 [row][col]
	dirs [][]int
}

// stripeWriter 第一遍：原始前三步；第二遍：用 dirs 精确连独立圆点。
type stripeWriter struct {
	inner *standard.Writer
	shape *StripeQRCode
}

func (w *stripeWriter) Write(mat qrcode.Matrix) error {
	w.shape.preScan(mat)
	return w.inner.Write(mat)
}

func (w *stripeWriter) Close() error { return w.inner.Close() }

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
	inner, err := standard.New(filePath, standard.WithCustomShape(s))
	if err != nil {
		return err
	}
	return qr.Save(&stripeWriter{inner: inner, shape: s})
}

// preScan 第一遍：原始前三步（col%3 对齐：padding=40,blockWidth=10 → (c+1)%3）。
// 第二遍：精确连接独立圆点。
func (s *StripeQRCode) preScan(mat qrcode.Matrix) {
	w, h := mat.Width(), mat.Height()
	dark := make([][]bool, h)
	for r := 0; r < h; r++ {
		dark[r] = make([]bool, w)
	}
	mat.Iterate(qrcode.IterDirection_ROW, func(x, y int, v qrcode.QRValue) {
		dark[y][x] = v.IsSet()
	})

	dirs := make([][]int, h)
	for r := 0; r < h; r++ {
		dirs[r] = make([]int, w)
	}

	neiOf := func(c, r int) uint16 {
		var n uint16
		if r > 0 && c > 0 && dark[r-1][c-1] { n |= standard.NTopLeft }
		if r > 0 && dark[r-1][c]           { n |= standard.NTop }
		if r > 0 && c < w-1 && dark[r-1][c+1] { n |= standard.NTopRight }
		if c > 0 && dark[r][c-1]           { n |= standard.NLeft }
		if c < w-1 && dark[r][c+1]         { n |= standard.NRight }
		if r < h-1 && c > 0 && dark[r+1][c-1] { n |= standard.NBotLeft }
		if r < h-1 && dark[r+1][c]         { n |= standard.NBot }
		if r < h-1 && c < w-1 && dark[r+1][c+1] { n |= standard.NBotRight }
		return n
	}

	// 偏移：padding=40, blockWidth=20 → col = 2+c, col%3 = (c+2)%3
	c3 := func(c int) int { return (c + 2) % 3 }
	r3 := func(r int) int { return (r + 2) % 3 }

	dirAt := func(c, r int) int {
		if c >= 0 && c < w && r >= 0 && r < h {
			return dirs[r][c]
		}
		return 0
	}

	// 单遍：当前行算前三步，延迟一行算第四步（下方格方向已出）
	for r := 0; r < h; r++ {
		for c := 0; c < w; c++ {
			if !dark[r][c] {
				continue
			}
			nei := neiOf(c, r)
			hasH := nei&(standard.NLeft|standard.NRight) != 0
			hasV := nei&(standard.NTop|standard.NBot) != 0

			// 第一步：横条
			wantH := hasH && nei&standard.NRight != 0 &&
				(nei&standard.NLeft == 0 || c3(c) != 2)
			// 第二步：竖条（ld!=1 精确替代 !hasH）
			wantV := hasV && nei&standard.NBot != 0 &&
				dirAt(c-1, r) != 1 &&
				nei&(standard.NBotLeft|standard.NBotRight) == 0 &&
				(nei&standard.NTop == 0 || r3(r) != 2)

			// 第三步：相交 → 竖赢保持均匀
			if wantH && wantV {
				wantH = false
			}

			if wantH {
				dirs[r][c] = 1
			} else if wantV {
				dirs[r][c] = 2
			}
		}

		// 第四步（延迟一行：处理 r-1 行，此时 r 行 dirs 已出）
		if r > 0 {
			for c := 0; c < w; c++ {
				if !dark[r-1][c] || dirs[r-1][c] != 0 {
					continue
				}
				nei := neiOf(c, r-1)
				if nei&standard.NBot == 0 {
					continue
				}
				if dirAt(c-1, r-1) == 1 { // 左邻画横
					continue
				}
				if dirAt(c, r) == 1 { // 下方格画横（本次已算）
					continue
				}
				if dirAt(c-1, r) == 1 { // 左下画横（本次已算）
					continue
				}
				dirs[r-1][c] = 2
			}
		}
	}
	// 末行第四步（无下一行）
	last := h - 1
	for c := 0; c < w; c++ {
		if !dark[last][c] || dirs[last][c] != 0 {
			continue
		}
		nei := neiOf(c, last)
		if nei&standard.NBot == 0 {
			continue
		}
		if dirAt(c-1, last) == 1 {
			continue
		}
		dirs[last][c] = 2
	}

	s.dirs = dirs
}

// Draw 实现 IShape 接口 —— 有预扫描 dirs 时查表，否则走逐格逻辑。
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

	var dir int
	// dirs 存矩阵坐标 [mr][mc]；图像坐标 (row,col) = 2 + (mr,mc)
	mr := row - 2
	mc := col - 2
	if s.dirs != nil && mr >= 0 && mr < len(s.dirs) && mc >= 0 && mc < len(s.dirs[mr]) {
		dir = s.dirs[mr][mc]
	} else {
		drawH := hasH && nei&standard.NRight != 0 &&
			(nei&standard.NLeft == 0 || col%3 != 2)
		drawV := !hasH && hasV && nei&standard.NBot != 0 &&
			nei&(standard.NBotLeft|standard.NBotRight) == 0 &&
			(nei&standard.NTop == 0 || row%3 != 2)

		if drawH {
			dir = 1
		} else if drawV {
			dir = 2
		} else if nei&standard.NLeft == 0 &&
			nei&standard.NBot != 0 &&
			nei&(standard.NBotLeft|standard.NBotRight) == 0 &&
			(nei&standard.NTop == 0 || row%3 != 2) {
			dir = 2
		}
	}

	if dir == 1 {
		ctx.DrawRectangle(cx, cy-r, mod, r*2)
	} else if dir == 2 {
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
