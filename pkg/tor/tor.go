package tor

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

// create a .torrent file from a magnet link and drop the torrent (unless Add2Client is true, then it is added)
func (c *Client) MagnetToTorrent(magnet string, dir *storage.ClientImpl, Add2Client bool) error {
	spec, err := torrent.TorrentSpecFromMagnetUri(magnet)
	if err != nil {
	}
	spec.Storage = *dir

	t, new, err := c.TorrentClient.AddTorrentSpec(spec)
	if err != nil {
		return err
	}
	<-t.GotInfo()

	mi := t.Metainfo()

	p := path.Join(c.DataDir, t.Info().BestName()) + ".torrent"
	f, err := os.Create(p)
	if err != nil {
		return err
	}

	defer f.Close()
	err = bencode.NewEncoder(f).Encode(mi)
	if err != nil {
		return fmt.Errorf("error writing torrent metainfo file: %s", err)
	}

	if !new && !Add2Client {
		t.Drop()
		return nil
	}

	return nil
}

// addMetaInfoFile takes a metainfo (.torrent) file and adds the torrent
// to the client. The file can either be local or accessible over
// HTTP(S). If it is a new torrent, the torrent info is added to the
// Bubbles list.
func (c *Client) addMetaInfoFile(path string, dir *storage.ClientImpl) error {
	var meta *metainfo.MetaInfo

	if strings.HasPrefix(path, "http") {
		response, err := http.Get(path)
		if err != nil {
			return err
		}
		defer response.Body.Close()

		meta, err = metainfo.Load(response.Body)
		if err != nil {
			return err
		}
	} else {
		var err error
		meta, err = metainfo.LoadFromFile(path)
		if err != nil {
			return err
		}
	}

	spec := torrent.TorrentSpecFromMetaInfo(meta)
	spec.Storage = *dir

	t, new, err := c.TorrentClient.AddTorrentSpec(spec)
	if err != nil {
		return err
	}

	if !new {
		return nil
	}

	<-t.GotInfo()
	return nil
}

func (c *Client) AddMagnetFromSpec(magnet string, dir *storage.ClientImpl) error {
	spec, err := torrent.TorrentSpecFromMagnetUri(magnet)
	if err != nil {
		return err
	}
	spec.Storage = *dir

	t, new, err := c.TorrentClient.AddTorrentSpec(spec)
	if !new {
		return nil
	}

	<-t.GotInfo()
	return nil
}
