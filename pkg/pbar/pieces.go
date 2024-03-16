package pbar

import (
	"strings"

	"github.com/anacrolix/torrent"
)

var (
	bar   = "\u2590"
	G     = "\x1b[32m"
	Y     = "\x1b[34m"
	B     = "\x1b[33m" // these are opposite
	BG    = "\x1b[42m" // these are opposite
	BB    = "\x1b[43m" // these are opposite
	BY    = "\x1b[44m" // these are opposite
	clear = "\x1b[0m"
)

func PieceBar(t *torrent.Torrent) string {
	var sb strings.Builder

	for i := int(0); i < t.NumPieces(); i++ {
		p := t.Piece(i)
		state := p.State()

		if state.Complete {
			sb.WriteString(G)
			sb.WriteString(clear)
		} else if state.Partial {
			sb.WriteString(Y)
			sb.WriteString(clear)
		} else {
			sb.WriteString(B)
			sb.WriteString(clear)
		}
	}
	return sb.String()
}
