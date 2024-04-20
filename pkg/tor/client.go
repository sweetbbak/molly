package tor

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	_ "modernc.org/sqlite"
)

type Client struct {
	// Client name
	Name string
	// directory to store sqlite torrent data
	DataDir string
	// actual place to store torrents
	DownloadDir string
	// internal torrent port
	TorrentPort int
	// listen address and port
	ListenAddr string
	// client
	TorrentClient *torrent.Client
	// use IPV4
	DisableIPV6 bool
	// Torrent Database
	db *sql.DB
}

// Initialize the torrent client
func NewClient() *Client {
	return &Client{}
}

// Initialize the torrent client with defaults
func NewDefaultClient() *Client {
	return &Client{
		DisableIPV6: false,
	}
}

func (c *Client) NewSession() error {
	cfg := torrent.NewDefaultClientConfig()

	var err error
	c.DataDir, err = c.getStorage()
	if err != nil {
		return err
	}

	cfg.DisableIPv6 = c.DisableIPV6
	cfg.Seed = true

	// TODO: change this to storing metadata in data dir (~/.local/share) and torrents wherever the user wants

	// cfg.SetListenAddr("localhost:42099")
	cfg.DefaultStorage = storage.NewFileByInfoHash(c.DataDir)

	client, err := torrent.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("error creating a new torrent client: %v", err)
	}

	c.TorrentClient = client
	return nil
}

// func (c *Client) ConfigureStorage() {
// }

// set the metadata location to a default directory. Torrents should be configured separately
func (c *Client) SetMetadataLocation(metadataDir string) storage.PieceCompletion {
	mstor, err := storage.NewDefaultPieceCompletionForDir(metadataDir)
	if err != nil {
		log.Println(err)
	}
	return mstor
}

func (c *Client) SetDefaultDownloadLocation(downloadDir string, mstor storage.PieceCompletion) storage.ClientImpl {
	tstor := storage.NewMMapWithCompletion(downloadDir, mstor)
	return tstor
}

// getStorage returns a storage implementation that writes downloaded
// files to a user-defined directory, and writes metadata files to a
// temporary directory.
func (c *Client) getMetadataDir(metadataDir, downloadDir string) (storage.ClientImpl, error) {
	mstor, err := storage.NewDefaultPieceCompletionForDir(metadataDir)
	if err != nil {
		log.Println(err)
		return storage.NewMMap(downloadDir), nil
	}

	tstor := storage.NewMMapWithCompletion(downloadDir, mstor)
	if err != nil {
		return nil, err
	}

	return tstor, err
}

// stop a torrent with a given infohash
func (c *Client) StopTorrent(infohash metainfo.Hash) error {
	t, ok := c.TorrentClient.Torrent(infohash)
	if !ok {
		return fmt.Errorf("error finding torrent with infohash [%s]", t.InfoHash().AsString())
	}
	if t.Info() != nil {
		t.CancelPieces(0, t.NumPieces())
	} else {
		return fmt.Errorf("torrent cannot be stopped, missing torrent metainfo for infohash [%s]", t.InfoHash().AsString())
	}
	return nil
}

// Add trackers to a specific torrent
func (c *Client) AddTrackers(infohash metainfo.Hash, announcelist [][]string) error {
	t, ok := c.TorrentClient.Torrent(infohash)
	if !ok {
		return fmt.Errorf("torrent cannot be stopped, missing torrent metainfo for infohash [%s]", t.InfoHash().AsString())
	}

	t.AddTrackers(announcelist)
	return nil
}

// stop all torrents that have seeded to the given ratio
func (c *Client) StopOnRatio(seedRatio float64) {
	ts := c.TorrentClient.Torrents()
	for _, t := range ts {
		if t == nil {
			continue
		}
		if t.Info() == nil {
			continue
		}
		stats := t.Stats()
		seedratio := float64(stats.BytesWrittenData.Int64()) / float64(stats.BytesReadData.Int64())
		if seedratio >= seedRatio {
			c.StopTorrent(t.InfoHash())
		}
	}
}

// start a specific file download
func (c *Client) StartFile(infohash metainfo.Hash, fp string) error {
	fp = filepath.ToSlash(fp)
	if fp == "" {
		return fmt.Errorf("Path cannot be empty")
	}

	t, ok := c.TorrentClient.Torrent(infohash)
	if !ok {
		return fmt.Errorf("error finding torrent with infohash [%s]", t.InfoHash().AsString())
	}

	var f *torrent.File
	for _, file := range t.Files() {
		if file.Path() == fp {
			f = file
			break
		}
	}

	if f == nil {
		return fmt.Errorf("could not find file [%s]", fp)
	}

	// normal should start the torrent file
	f.SetPriority(torrent.PiecePriorityNormal)

	return nil
}

// stop a specific file download
func (c *Client) StopFile(infohash metainfo.Hash, fp string) error {
	fp = filepath.ToSlash(fp)
	if fp == "" {
		return fmt.Errorf("Path cannot be empty")
	}

	t, ok := c.TorrentClient.Torrent(infohash)
	if !ok {
		return fmt.Errorf("error finding torrent with infohash [%s]", t.InfoHash().AsString())
	}

	var f *torrent.File
	for _, file := range t.Files() {
		if file.Path() == fp {
			f = file
			break
		}
	}

	if f == nil {
		return fmt.Errorf("could not find file [%s]", fp)
	}

	// None should stop the torrent file
	f.SetPriority(torrent.PiecePriorityNone)

	return nil
}

// returns a slice of loaded torrents or nil
func (c *Client) ShowTorrents() []*torrent.Torrent {
	return c.TorrentClient.Torrents()
}

// generic add torrent function
func (c *Client) AddTorrent(tor string) (*torrent.Torrent, error) {
	if strings.HasPrefix(tor, "magnet") {
		return c.AddMagnet(tor)
	} else if strings.Contains(tor, "http") {
		return c.AddTorrentURL(tor)
	} else {
		return c.AddTorrentFile(tor)
	}
}

// add a torrent from a magnet link
func (c *Client) AddMagnet(magnet string) (*torrent.Torrent, error) {
	t, err := c.TorrentClient.AddMagnet(magnet)
	if err != nil {
		return nil, err
	}
	<-t.GotInfo()
	return t, nil
}

// add torrent from a .torrent file
func (c *Client) AddTorrentFile(file string) (*torrent.Torrent, error) {
	t, err := c.TorrentClient.AddTorrentFromFile(file)
	if err != nil {
		return nil, err
	}
	<-t.GotInfo()
	return t, nil
}

// add a torrent from a URL containing a torrent file
func (c *Client) AddTorrentURL(url string) (*torrent.Torrent, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fname := path.Base(url)
	tmp := os.TempDir()
	path.Join(tmp, fname)

	file, err := os.Create(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return nil, err
	}

	t, err := c.TorrentClient.AddTorrentFromFile(file.Name())
	if err != nil {
		return nil, err
	}
	<-t.GotInfo()
	return t, nil
}

// add multiple torrents from DIR
func (c *Client) AddTorrentsFromDir(dir string) ([]*torrent.Torrent, error) {
	_, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("directory [%s] stat error: %v", err)
	}
	dir = fmt.Sprintf("%s/*.torrent", dir)

	matches, err := filepath.Glob(dir)
	if err != nil {
		return nil, err
	}

	var ts []*torrent.Torrent

	for _, tf := range matches {
		t, err := c.TorrentClient.AddTorrentFromFile(tf)
		if err != nil {
			return nil, err
		}
		<-t.GotInfo()
		ts = append(ts, t)
	}
	return ts, nil
}

// stops the client and closes all connections to peers
func (c *Client) Close() (errs []error) {
	return c.TorrentClient.Close()
}

// look through the torrent files the client is handling and return a torrent with a
// matching info hash
func (c *Client) FindByInfoHhash(infoHash string) (*torrent.Torrent, error) {
	torrents := c.TorrentClient.Torrents()
	for _, t := range torrents {
		if t.InfoHash().AsString() == infoHash {
			return t, nil
		}
	}
	return nil, fmt.Errorf("No torrents match info hash: %v", infoHash)
}

// drop a torrent entirely
func (c *Client) DropTorrent(t *torrent.Torrent) {
	t.Drop()
}
