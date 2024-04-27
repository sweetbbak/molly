package tor

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

// add torrent from spec into the client
// this allows us to preserve more of the original state
func (c *Client) AddTorrentFromSpec(spec *torrent.TorrentSpec, dontStart bool) error {
	t, new, err := c.TorrentClient.AddTorrentSpec(spec)
	if err != nil {
		return fmt.Errorf("error adding torrent from spec: %v", t.Name())
	}

	if !new {
		return fmt.Errorf("torrent with infohash [%s] is not new", spec.InfoHash.AsString())
	}

	if !t.Complete.Bool() && !dontStart {
		t.DownloadAll()
	}

	return nil
}

// SpecfromPath Returns Torrent Spec from File Path
func SpecfromPath(path string) (spec *torrent.TorrentSpec, reterr error) {
	// TorrentSpecFromMetaInfo may panic if the info is malformed
	defer func() {
		if r := recover(); r != nil {
			reterr = fmt.Errorf("SpecfromPath: error loading new torrent from file %s: %v+", path, r)
		}
	}()

	fi, err := os.Stat(path)

	if os.IsNotExist(err) {
		return nil, fmt.Errorf("file doesn't exist")
	}

	if fi.IsDir() {
		return nil, fmt.Errorf("directory present")
	}

	info, reterr := metainfo.LoadFromFile(path)
	if reterr != nil {
		return
	}
	spec = torrent.TorrentSpecFromMetaInfo(info)
	return
}

// SpecfromBytes Returns Torrent Spec from Bytes
func SpecfromBytes(trnt []byte) (spec *torrent.TorrentSpec, reterr error) {
	// TorrentSpecFromMetaInfo may panic if the info is malformed
	defer func() {
		if r := recover(); r != nil {
			reterr = fmt.Errorf("SpecfromBytes: error loading new torrent from bytes")
		}
	}()
	info, reterr := metainfo.Load(bytes.NewReader(trnt))
	if reterr != nil {
		return nil, reterr
	}
	spec = torrent.TorrentSpecFromMetaInfo(info)
	return
}

// SpecfromB64String Returns Torrent Spec from Base64 Encoded Torrent File
func SpecfromB64String(trnt string) (spec *torrent.TorrentSpec, reterr error) {
	t, err := base64.StdEncoding.DecodeString(trnt)
	if err != nil {
		return nil, err
	}
	return SpecfromBytes(t)
}

// MetafromHex returns metainfo.Hash from given infohash string
func MetafromHex(infohash string) (h metainfo.Hash, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error parsing string to InfoHash")
		}
	}()

	h = metainfo.NewHashFromHex(infohash)

	return h, nil
}

// RemTrackersSpec removes trackers from torrent.Spec
func RemTrackersSpec(spec *torrent.TorrentSpec) {
	if spec == nil {
		return
	}
	spec.Trackers = [][]string{}
}
