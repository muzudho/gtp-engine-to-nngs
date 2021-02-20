package phase

import (
	"fmt"

	"github.com/muzudho/kifuwarabe-gtp/entities/phase"
)

// FlipColorString - 色を反転
func FlipColorString(color string) string {
	switch color {
	case "B":
		return "W"
	case "W":
		return "B"
	case "b":
		return "w"
	case "w":
		return "b"
	default:
		return color
	}
}

// ToString - 色を大文字アルファベットに変換
func ToString(ph phase.Phase) string {
	switch ph {
	case phase.Black:
		return "B"
	case phase.White:
		return "W"
	default:
		panic(fmt.Sprintf("Unexpected phase=[%d]", ph))
	}
}

// ToNum - アルファベットを色に変換
func ToNum(color string) phase.Phase {
	switch color {
	case "B":
		return phase.Black
	case "W":
		return phase.White
	case "b":
		return phase.Black
	case "w":
		return phase.White
	default:
		panic(fmt.Sprintf("Unexpected color=[%s]", color))
	}
}
