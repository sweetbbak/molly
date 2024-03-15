package bt

import (
	// "fmt"
	"net/http"
	"os"
	"strings"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

var (
	magnetPrefix   = "magnet:"
	infohashPrefix = "infohash:"
	hashLength     = 40
)

// addInfoHash takes an infohash and adds the torrent to the client. If
// it is a new torrent, the torrent info is added to the Bubbles list.
func addInfoHash(input string, dir *storage.ClientImpl, client *torrent.Client) error {
	if strings.HasPrefix(input, infohashPrefix) {
		input = strings.TrimPrefix(input, infohashPrefix)
	}

	hash := metainfo.NewHashFromHex(input)
	t, new := client.AddTorrentInfoHashWithStorage(hash, *dir)

	if !new {
		downloadTorrent(t)
	}
	return nil
}

// addMagnetLink takes a magnet link and adds the torrent to the
// client. If it is a new torrent, the torrent info is added to the
// Bubbles list.
func addMagnetLink(input string, dir *storage.ClientImpl, client *torrent.Client) error {
	spec, err := torrent.TorrentSpecFromMagnetUri(input)
	if err != nil {
	}
	spec.Storage = *dir

	t, new, err := client.AddTorrentSpec(spec)

	if !new {
		downloadTorrent(t)
	}
	return nil
}

// AddTorrents takes a group of torrents as strings, adds them to the
// client, and returns them batched.
func AddTorrents(t *[]string, dir string, client *torrent.Client) {
	for _, v := range *t {
		err := AddTorrent(v, dir, client)
		if err != nil {
		}
	}
}

// AddTorrent takes a torrent as a string and adds it to the client
// based on its format.
func AddTorrent(t, dir string, client *torrent.Client) error {
	store := getStorage(dir)
	if strings.HasPrefix(t, magnetPrefix) {
		return addMagnetLink(t, &store, client)
	} else if strings.HasPrefix(t, infohashPrefix) || len(t) == hashLength {
		return addInfoHash(t, &store, client)
	} else {
		return addMetaInfoFile(t, &store, client)
	}
}

// downloadTorrent asynchronously waits for torrent info to arrive,
// triggers the download, and returns a Bubble Tea start message.
func downloadTorrent(t *torrent.Torrent) {
	<-t.GotInfo()
	t.DownloadAll()
}

// addMetaInfoFile takes a metainfo (.torrent) file and adds the torrent
// to the client. The file can either be local or accessible over
// HTTP(S). If it is a new torrent, the torrent info is added to the
// Bubbles list.
func addMetaInfoFile(input string, dir *storage.ClientImpl, client *torrent.Client) error {
	var meta *metainfo.MetaInfo

	if strings.HasPrefix(input, "http") {
		response, err := http.Get(input)
		if err != nil {
			return err
		}

		meta, err = metainfo.Load(response.Body)
		defer response.Body.Close()
		if err != nil {
			return err
		}
	} else {
		path := os.ExpandEnv(input)
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
		downloadTorrent(t)
	}
	return nil
}

// getStorage returns a storage implementation that writes downloaded
// files to a user-defined directory, and writes metadata files to a
// temporary directory.
func getStorage(dir string) storage.ClientImpl {
	metadataDirectory := os.TempDir()
	if metadataStorage, err := storage.NewDefaultPieceCompletionForDir(metadataDirectory); err != nil {
		return storage.NewMMap(dir)
	} else {
		return storage.NewMMapWithCompletion(dir, metadataStorage)
	}
}
