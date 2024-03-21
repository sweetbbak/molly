package tor

import (
	"net/http"
	"strings"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

// addMetaInfoFile takes a metainfo (.torrent) file and adds the torrent
// to the client. The file can either be local or accessible over
// HTTP(S). If it is a new torrent, the torrent info is added to the
// Bubbles list.
func addMetaInfoFile(path string, dir *storage.ClientImpl, client *torrent.Client) error {
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

	t, new, err := client.AddTorrentSpec(spec)
	if err != nil {
		return err
	}

	if !new {
		return nil
	}

	<-t.GotInfo()
	return nil
}

func (c *Client) Torra(t *torrent.Torrent, magnet string, dir *storage.ClientImpl) error {
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
