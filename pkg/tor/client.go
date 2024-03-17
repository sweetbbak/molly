package tor

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
)

type Client struct {
	// Client name
	Name string
	// directory to store sqlite torrent data
	DataDir string
	// actual place to store torrents
	Download string
	// internal torrent port
	TorrentPort int
	// listen address and port
	ListenAddr string
	// client
	TorrentClient *torrent.Client
	// slice of torrents
	// Torrents []*torrent.Torrent
	// use IPV4
	DisableIPV6 bool
}

// Initialize the torrent client
func NewClient() *Client {
	return &Client{}
}

// Initialize the torrent client with sensible defaults
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

	fmt.Println(c)
	cfg.DisableIPv6 = c.DisableIPV6

	// cfg.SetListenAddr("localhost:42099")
	cfg.DefaultStorage = storage.NewFileByInfoHash(c.DataDir)

	client, err := torrent.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("error creating a new torrent client: %v", err)
	}

	c.TorrentClient = client
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

func (c *Client) DropTorrent(t *torrent.Torrent) {
	t.Drop()
}

// Create storage path if it doesn't exist and return Path
func (c *Client) getStorage() (string, error) {
	s, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("Client error, couldnt get user cache directory: %v", err)
	}

	p := path.Join(s, c.Name)
	if p == "" || c.Name == "" {
		return "", fmt.Errorf("Client error, couldnt construct client path: Empty path or project name")
	}

	err = os.MkdirAll(p, 0o755)
	if err != nil {
		return "", fmt.Errorf("Client error, couldnt create project directory: %v", err)
	}

	_, err = os.Stat(p)
	if err == nil {
		return p, nil
	} else {
		return "", err
	}
}
