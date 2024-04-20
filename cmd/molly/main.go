package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	// "molly/pkg/pbar"
	"molly/pkg/mydb"
	"molly/pkg/tor"

	"github.com/anacrolix/torrent/metainfo"
	"golang.org/x/net/context"
)

func restartTorrent(ih metainfo.Hash, client *tor.Client) {
	client.TorrentClient.AddTorrentInfoHash(ih)
}

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

	if !mydb.Exists(t.InfoHash()) {
		fmt.Printf("torrent doesnt exist, adding it now: %s", t.InfoHash().String())
		err = mydb.Add(t.InfoHash())
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("torrent exists, starting: %s", t.InfoHash().String())
		if mydb.HasStarted(t.InfoHash().String()) {
			restartTorrent(t.InfoHash(), client)
		}
	}

	for !t.Complete.Bool() {
		r := tor.TorrentSeedRatio(t)
		pr := tor.TorrentPercentage(t)
		rd := tor.TorrentRatioFromDownload(t)

		println("\x1b[2J\x1b[H")
		fmt.Printf("\n\nratio [ %.1f ] percentage (%.1f) ratio_from_dl (%.1f)\n", r, pr, rd)
		time.Sleep(time.Millisecond * 99)
	}

	// // 1465 or 4262 or 945
	// for !t.Complete.Bool() {
	// 	// vv := tor.Veri(t)
	// 	p := pbar.PieceBar2(t)
	// 	println("\x1b[2J\x1b[H")
	// 	fmt.Printf("\n%s\n", p)
	//
	// 	r := tor.TorrentSeedRatio(t)
	// 	pr := tor.TorrentPercentage(t)
	// 	rd := tor.TorrentRatioFromDownload(t)
	//
	// 	fmt.Printf("\n\nratio [ %f ] (%f) (%f)\n", r, pr, rd)
	// 	time.Sleep(time.Millisecond * 99)
	// }
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
	if err := Start(); err != nil {
		log.Fatal(err)
	}

	if err := molly(); err != nil {
		log.Fatal(err)
	}
}
