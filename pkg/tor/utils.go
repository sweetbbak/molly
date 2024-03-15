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
