package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"molly/pkg/pbar"
	"molly/pkg/tor"

	"golang.org/x/net/context"
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

func findTorrents() {
	matches, _ := filepath.Glob("/home/sweet/Downloads/*.torrent")
	for _, m := range matches {
		fmt.Println(m)
	}
}

func testWatch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			break
		default:
			findTorrents()
		}

		time.Sleep(time.Second * 1)
	}
}

func breaker(cancel context.CancelFunc) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	for {
		<-c
		cancel()
		println("canceling...")
		os.Exit(0)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go breaker(cancel)
	testWatch(ctx)

	if err := molly(); err != nil {
		log.Fatal(err)
	}
}
