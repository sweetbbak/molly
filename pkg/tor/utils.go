package tor

import (
	"strings"

	"github.com/anacrolix/torrent"
)

func escapeUrl(u string) string {
	u = strings.ReplaceAll(u, "'", "")
	u = strings.ReplaceAll(u, "\n", "")
	u = strings.ReplaceAll(u, " ", "_")
	u = strings.ReplaceAll(u, "_-_", "_")
	u = strings.ReplaceAll(u, "__", "_")
	u = strings.ReplaceAll(u, "--", "-")
	return u
}

// Get largest file inside of a Torrent
func GetLargestFile(t *torrent.Torrent) *torrent.File {
	var target *torrent.File
	var maxSize int64

	for _, file := range t.Files() {
		if maxSize < file.Length() {
			maxSize = file.Length()
			target = file
		}
	}
	return target
}

// TODO: add a way to visualize what pieces have been downloaded like the old torrent clients
// [||||||||||||] where the missing pieces are greyed out, bad pieces are red and good pieces are green
func Veri(t *torrent.Torrent) string {
	var sb strings.Builder
	var (
		gween  = "\x1b[32m|"
		yellow = "\x1b[34m|"
		blue   = "\x1b[33m|"
		clear  = "\x1b[0m"
	)

	for i := int(0); i < t.NumPieces(); i++ {
		p := t.Piece(i)
		// pi := p.Info()
		// idx := pi.Index()
		state := p.State()

		if state.Complete {
			sb.WriteString(gween)
			sb.WriteString(clear)
		} else if state.Partial {
			sb.WriteString(yellow)
			sb.WriteString(clear)
		} else {
			sb.WriteString(blue)
			sb.WriteString(clear)
		}

		// t.Piece(i).VerifyData()
	}
	return sb.String()
}

// returns a seed ratio compared to the entire torrent
func TorrentRatio(t *torrent.Torrent) float64 {
	stats := t.Stats()
	upload := stats.BytesWritten.Int64()
	// return float64(t.Length()) / float64(upload)
	return float64(upload) / float64(t.Length())
}

// returns a seed ratio compared to the amount of data the user downloaded
func TorrentRatioFromDownload(t *torrent.Torrent) float64 {
	stats := t.Stats()
	upload := stats.BytesWritten.Int64()
	return float64(t.BytesCompleted()) / float64(upload)
}

// get the downloaded percentage of a torrent
func TorrentPercentage(t *torrent.Torrent) float64 {
	info := t.Info()

	if info == nil {
		return 0
	}

	return float64(t.BytesCompleted()) / float64(info.TotalLength()) * 100
}
