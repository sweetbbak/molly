package tor

import (
	"fmt"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/dustin/go-humanize"
)

const megabyte = 1024 * 1024

// Download retrieves the download information for a torrent and
// returns it as a string.
func Downloaded(t *torrent.Torrent, showPercent bool) string {
	var (
		done    = t.BytesCompleted()
		total   = t.Length()
		percent = float64(done) / float64(total) * 100

		tail string
	)

	if showPercent {
		tail = fmt.Sprintf(" (%d%%)", uint64(percent))
	}

	return fmt.Sprintf(
		"%s/%s%s ↓",
		humanize.Bytes(uint64(done)),
		humanize.Bytes(uint64(total)),
		tail,
	)
}

// calculateDownloadRate calculates the download rate in MB/s.
func calculateDownloadRate(bytesCompleted, startSize int64, elapsedTime time.Duration) float64 {
	return float64(bytesCompleted-startSize) / elapsedTime.Seconds() / megabyte
}

// Peers retrieves the peer information for a torrent and returns it as
// a string.
func Peers(t *torrent.Torrent) string {
	stats := t.Stats()

	return fmt.Sprintf(
		"%d/%d peers",
		stats.ActivePeers,
		stats.TotalPeers,
	)
}

// Upload retrieves the amount of data seeded for a torrent and returns
// it as a string.
func Upload(t *torrent.Torrent) string {
	var (
		stats  = t.Stats()
		upload = stats.BytesWritten.Int64()
	)

	return fmt.Sprintf(
		"%s ↑",
		humanize.Bytes(uint64(upload)),
	)
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
		blue   = "\x1b[33m|" // these are opposite
		clear  = "\x1b[0m"
	)

	for i := int(0); i < t.NumPieces(); i++ {
		p := t.Piece(i)
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
	}
	return sb.String()
}

// returns a seed ratio compared to the entire torrent
func TorrentSeedRatio(t *torrent.Torrent) float64 {
	stats := t.Stats()
	seedratio := float64(stats.BytesWrittenData.Int64()) / float64(stats.BytesReadData.Int64())
	return seedratio
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
