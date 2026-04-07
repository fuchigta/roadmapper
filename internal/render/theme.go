package render

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// BrandColors はブランドカラーから派生した色セット。
type BrandColors struct {
	Base  string
	Light string
}

// DeriveColors は "#rrggbb" 形式のカラーコードから派生色を計算する。
func DeriveColors(hex string) BrandColors {
	r, g, b := parseHex(hex)
	// ライトバリアントは彩度維持で明度を上げる
	h, s, l := rgbToHSL(r, g, b)
	lightL := math.Min(l+0.42, 0.95)
	lr, lg, lb := hslToRGB(h, s, lightL)
	return BrandColors{
		Base:  hex,
		Light: fmt.Sprintf("#%02x%02x%02x", lr, lg, lb),
	}
}

func parseHex(hex string) (r, g, b float64) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) < 6 {
		return 0, 0, 0
	}
	ri, _ := strconv.ParseInt(hex[0:2], 16, 32)
	gi, _ := strconv.ParseInt(hex[2:4], 16, 32)
	bi, _ := strconv.ParseInt(hex[4:6], 16, 32)
	return float64(ri) / 255, float64(gi) / 255, float64(bi) / 255
}

func rgbToHSL(r, g, b float64) (h, s, l float64) {
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	l = (max + min) / 2
	if max == min {
		return 0, 0, l
	}
	d := max - min
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}
	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	case b:
		h = (r-g)/d + 4
	}
	h /= 6
	return
}

func hslToRGB(h, s, l float64) (r, g, b uint8) {
	if s == 0 {
		v := uint8(l * 255)
		return v, v, v
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	rf := hue(p, q, h+1.0/3)
	gf := hue(p, q, h)
	bf := hue(p, q, h-1.0/3)
	return uint8(rf * 255), uint8(gf * 255), uint8(bf * 255)
}

func hue(p, q, t float64) float64 {
	if t < 0 {
		t++
	}
	if t > 1 {
		t--
	}
	switch {
	case t < 1.0/6:
		return p + (q-p)*6*t
	case t < 1.0/2:
		return q
	case t < 2.0/3:
		return p + (q-p)*(2.0/3-t)*6
	default:
		return p
	}
}
