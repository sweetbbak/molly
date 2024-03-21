package tor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type TorFunc func(string)

func (c *Client) WatchDir(paths []string) context.Context {
	var tfunc TorFunc
	var added []string

	tfunc = func(tor string) {
		for _, r := range added {
			if tor == r {
				return
			}
		}

		_, err := c.AddTorrentFile(tor)
		if err != nil {
		}

		added = append(added, tor)
	}

	ctx, _ := context.WithCancel(context.Background())
	go watcher(ctx, paths, tfunc)
	return ctx
}

func watcher(ctx context.Context, paths []string, fn TorFunc) {
	for {
		select {
		case <-ctx.Done():
			break
		default:
			ts := findTorrents(paths)
			for _, tx := range ts {
				fn(tx)
			}
		}

		time.Sleep(time.Second * 1)
	}
}

// watch a set of directories for torrents
func findTorrents(paths []string) []string {
	var matches []string
	for _, p := range paths {
		p = os.ExpandEnv(p)
		if p[len(p)-1] == byte('/') {
			p = p[:len(p)-1]
		}

		p = fmt.Sprintf("%s/*.torrent", p)

		mz, _ := filepath.Glob(p)
		if len(mz) == 0 {
			continue
		}

		for _, m := range mz {
			fmt.Println(m)
		}

		matches = append(matches, mz...)
	}

	return matches
}
