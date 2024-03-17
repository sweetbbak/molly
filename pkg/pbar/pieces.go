package pbar

import (
	"strings"

	"github.com/anacrolix/torrent"
)

var (
	bar   = "\u2590"
	R     = "\x1b[31m"
	G     = "\x1b[32m"
	B     = "\x1b[33m" // these are opposite
	Y     = "\x1b[34m"
	BR    = "\x1b[41m"
	BG    = "\x1b[42m" // these are opposite
	BB    = "\x1b[43m" // these are opposite
	BY    = "\x1b[44m" // these are opposite
	clear = "\x1b[0m"
)

func chunk(t *torrent.Torrent, start, end int) int {
	x, y, z := 0, 0, 0

	for i := start; i < end; i++ {
		state := t.Piece(i).State()
		if state.Complete {
			x++
		} else if state.Partial {
			y++
		} else {
			z++
		}
	}
	if x > y && x > z {
		return 0
	} else if y > x && y > z {
		return 1
	} else {
		return 2
	}
}

func PieceBar2(t *torrent.Torrent) string {
	var sb strings.Builder
	total := t.NumPieces()
	size := 12

	for i := 0; i <= total; i = i + size {
		var x int
		if i+size > total {
			x = total - 1
		} else {
			x = i + size
		}

		col := chunk(t, i, x)

		switch col {
		case 0:
			sb.WriteString(G)
			sb.WriteString(bar)
			sb.WriteString(clear)
		case 1:
			sb.WriteString(Y)
			sb.WriteString(bar)
			sb.WriteString(clear)
		case 2:
			sb.WriteString(B)
			sb.WriteString(bar)
			sb.WriteString(clear)
		}

	}
	sb.WriteString("\n")
	return sb.String()
}

func PieceBar(t *torrent.Torrent) string {
	var sb strings.Builder

	for i := int(0); i < t.NumPieces(); i++ {
		p := t.Piece(i)
		state := p.State()

		if state.Complete {
			sb.WriteString(G)
			sb.WriteString(bar)
			sb.WriteString(clear)
		} else if state.Partial {
			sb.WriteString(Y)
			sb.WriteString(bar)
			sb.WriteString(clear)
		} else {
			sb.WriteString(B)
			sb.WriteString(bar)
			sb.WriteString(clear)
		}
	}
	return sb.String()
}
