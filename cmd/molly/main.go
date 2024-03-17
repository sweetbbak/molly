package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"molly/pkg/pbar"
	"molly/pkg/tor"
)

func molly() error {
	client := tor.NewClient()
	client.Name = "molly"

	err := client.NewSession()
	if err != nil {
		return err
	}

	t, err := client.AddTorrent(os.Args[1])
	if err != nil {
		return err
	}

	t.DownloadAll()
	// 1465 or 4262 or 945
	for !t.Complete.Bool() {
		// vv := tor.Veri(t)
		p := pbar.PieceBar2(t)
		println("\x1b[2J\x1b[H")
		fmt.Printf("\n%s\n", p)

		r := tor.TorrentSeedRatio(t)
		pr := tor.TorrentPercentage(t)
		rd := tor.TorrentRatioFromDownload(t)

		fmt.Printf("\n\nratio [ %f ] (%f) (%f)\n", r, pr, rd)
		time.Sleep(time.Millisecond * 99)
	}
	return nil
}

func main() {
	if err := molly(); err != nil {
		log.Fatal(err)
	}
}
