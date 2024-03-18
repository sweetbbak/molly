package tor

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

type State struct {
	infohash       string
	seeding        bool
	BytesCompleted int64
	BytesMissing   int64
	Length         int64
	State          string
	complete       bool
	hash           string
}

func (c *Client) RecoverState(infohash metainfo.Hash) error {
	c.AddTorrentFromSpec(&torrent.TorrentSpec{InfoHash: infohash}, true)
	return nil
}

func (c *Client) SaveState() error {
	ts := c.ShowTorrents()
	if ts == nil {
		return fmt.Errorf("no torrents found")
	}

	var states []State
	for _, t := range ts {
		var s State
		s.seeding = t.Seeding()
		s.complete = t.Complete.Bool()
		s.hash = t.InfoHash().AsString()

		states = append(states, s)
	}

	b, err := json.Marshal(states)
	if err != nil {
		return err
	}
	out := path.Join(c.DataDir, "state.json")
	p, err := os.Create(out)
	if err != nil {
		return err
	}

	_, err = p.Write(b)
	return err
}
