package util

import (
	"image/color"
	"os"
	"testing"

	"github.com/yeqown/go-qrcode/writer/standard"
)

func TestColor(t *testing.T) {
	v := color.Color(color.RGBA{R: 255, G: 255, B: 255, A: 255}) == color.Color(color.RGBA{R: 255, G: 255, B: 255, A: 255})
	t.Log(v)
}

func TestStripeQRCode_GenerateFile(t *testing.T) {
	// 1) 默认黑白条纹风格，直接写文件
	err := GenerateStripeQRCode("https://github.com/chuccp/go-web-frame", "test_stripe_default.png")
	if err != nil {
		t.Fatalf("默认条纹生成失败: %v", err)
	}
	info, err := os.Stat("test_stripe_default.png")
	if err != nil {
		t.Fatalf("文件未生成: %v", err)
	}
	t.Logf("生成: %s, %d bytes", info.Name(), info.Size())
	defer os.Remove("test_stripe_default.png")

	// 2) 自定义红色 + 大模块
	s := NewStripeQRCode().
		WithModuleSize(14).
		WithColors(
			color.RGBA{R: 240, G: 240, B: 240, A: 255},
			color.RGBA{R: 180, G: 30, B: 50, A: 255},
			color.RGBA{R: 240, G: 240, B: 240, A: 255},
		)
	err = s.GenerateFile("https://github.com/chuccp/go-web-frame", "test_stripe_red.png")
	if err != nil {
		t.Fatalf("红色条纹生成失败: %v", err)
	}
	t.Logf("生成: test_stripe_red.png")
	//defer os.Remove("test_stripe_red.png")

	// 3) 蓝色小模块 + 短内容
	s2 := NewStripeQRCode().WithModuleSize(8)
	err = s2.GenerateFile("Hello 条纹 QR!", "test_stripe_blue.png")
	if err != nil {
		t.Fatalf("蓝色条纹生成失败: %v", err)
	}
	t.Logf("生成: test_stripe_blue.png")
	//defer os.Remove("test_stripe_blue.png")
}

func TestStripeQRCode_ISHapeAPI(t *testing.T) {
	// 通过 GenerateQrcode + WithStripeShape（走 IShape，无预扫描回退）
	buf := CreateBufferWriteCloser()
	err := GenerateQrcode("https://github.com/chuccp/go-web-frame", buf, WithStripeShape())
	if err != nil {
		t.Fatalf("IShape 接口方式生成失败: %v", err)
	}
	t.Logf("IShape 接口 → %d bytes", len(buf.Bytes()))

	buf2 := CreateBufferWriteCloser()
	err = GenerateQrcode("Hello IShape!", buf2, WithCustomStripeShape(color.RGBA{R: 0, G: 100, B: 80, A: 255}))
	if err != nil {
		t.Fatalf("自定义颜色 IShape 生成失败: %v", err)
	}
	t.Logf("自定义颜色 IShape → %d bytes", len(buf2.Bytes()))
}

func TestOriginQRCodeStyles(t *testing.T) {
	// 圆角方块
	buf1 := CreateBufferWriteCloser()
	err := GenerateQrcode("test-rounded", buf1, WithRoundedSquareShape())
	if err != nil {
		t.Fatalf("圆角方块失败: %v", err)
	}
	t.Logf("圆角方块 → %d bytes", len(buf1.Bytes()))

	// 圆形
	buf2 := CreateBufferWriteCloser()
	err = GenerateQrcode("test-circle", buf2, WithCircleShape())
	if err != nil {
		t.Fatalf("圆形失败: %v", err)
	}
	t.Logf("圆形 → %d bytes", len(buf2.Bytes()))
}

// TestPreviewQRStyles 模拟预览接口生成图片，保存文件供人工检查。
// 对应: /admin/api/qrcode/style/preview-qr?color=%23000000&style=XXX
func TestPreviewQRStyles(t *testing.T) {
	content := "http://localhost:8082"

	makeOpts := func(shapeOpt standard.ImageOption) []standard.ImageOption {
		opts := []standard.ImageOption{
			shapeOpt,
			standard.WithBorderWidth(4),
			standard.WithBgTransparent(),
			standard.WithBuiltinImageEncoder(standard.PNG_FORMAT),
		}
		return opts
	}

	tests := []struct {
		name    string
		shape   standard.ImageOption
		outFile string
	}{
		{"circle", WithCircleShape(), "test_preview_circle.png"},
		{"rounded_square", WithRoundedSquareShape(), "test_preview_rounded_square.png"},
		{"stripe", WithStripeShape(), "test_preview_stripe.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := CreateBufferWriteCloser()
			err := GenerateQrcode(content, buf, makeOpts(tt.shape)...)
			if err != nil {
				t.Fatalf("%s 生成失败: %v", tt.name, err)
			}
			if err := os.WriteFile(tt.outFile, buf.Bytes(), 0644); err != nil {
				t.Fatalf("写文件失败: %v", err)
			}
			t.Logf("%s → %s (%d bytes)", tt.name, tt.outFile, len(buf.Bytes()))
		})
	}
}
