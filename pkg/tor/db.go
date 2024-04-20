package tor

import (
	"context"
	"database/sql"
	"time"

	"github.com/anacrolix/torrent/metainfo"
	_ "modernc.org/sqlite"
	// "github.com/anacrolix/torrent"
	// "github.com/anacrolix/torrent/storage"
)

type TorrentDB struct {
	db *sql.DB
}

type Torrent struct {
	Infohash  metainfo.Hash
	Started   bool
	AddedAt   time.Time
	StartedAt time.Time
}

// Initialize the database and create initial tables
func (c *Client) DatabaseInit(dbPath string) error {
	var err error
	c.db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	_, err = c.db.ExecContext(context.Background(), `create table if not exists torrent (infohash text primary key,started boolean,addedat timestamptz,startedat timestamptz);`)
	if err != nil {
		return err
	}
	return nil
}

// Close the database
func (c *Client) DBClose() {
	c.db.Close()
}

// Returns whether infohash exists in DB
func (c *Client) DBExists(ih metainfo.Hash) (ret bool) {
	var err error
	row := c.db.QueryRowContext(context.Background(), `select true from torrent where infohash=$1;`, ih.HexString())
	err = row.Scan(&ret)
	if err != nil {
		return false
	}
	return
}

// Returns whether torrent is started or not
func (c *Client) DBHasStarted(ih string) (ret bool) {
	row := c.db.QueryRowContext(context.Background(), `select started from torrent where infohash=$1;`, ih)
	err := row.Scan(&ret)
	if err != nil {
		return false
	}
	return
}

// Create a new torrent entry in the database
func (c *Client) DBAdd(ih metainfo.Hash) (err error) {
	_, err = c.db.ExecContext(context.Background(), `insert into torrent (infohash,started,addedat,startedat) values ($1,$2,$3,$4) on conflict (infohash) do update set startedat=$4;`, ih.HexString(), false, time.Now(), time.Now())
	return
}

// Remove torrent from DB
func (c *Client) DBDelete(ih metainfo.Hash) (err error) {
	_, err = c.db.ExecContext(context.Background(), `delete from torrent where infohash=$1;`, ih.HexString())
	return
}

// Set torrent state to started, and add the current time to startedat
func (c *Client) DBStart(ih metainfo.Hash) (err error) {
	_, err = c.db.ExecContext(context.Background(), `update torrent set started=$1,startedat=$2 where infohash=$3;`, true, time.Now(), ih.HexString())
	return
}

// Set torrent to started state
func (c *Client) DBSetStarted(ih metainfo.Hash, inp bool) (err error) {
	_, err = c.db.ExecContext(context.Background(), `update torrent set started=$1 where infohash=$2;`, inp, ih.HexString())
	return
}

// Save the torrent state for when a torrent is stopeed (no longer seeding, is completed or stopped by user)
func (c *Client) DBSetStopped(ih metainfo.Hash, inp bool) (err error) {
	_, err = c.db.ExecContext(context.Background(), `update torrent set started=$1 where infohash=$2;`, inp, ih.HexString())
	return
}

// Get info about torrent with Infohash
func (c *Client) GetTorrent(ih metainfo.Hash) (*Torrent, error) {
	var trnt Torrent
	var infoh string
	row := c.db.QueryRowContext(context.Background(), `select * from torrent where infohash=$1;`, ih.HexString())
	err := row.Scan(&infoh, &trnt.Started, &trnt.AddedAt, &trnt.StartedAt)
	if err != nil {
		return nil, err
	}
	trnt.Infohash = ih
	return &trnt, nil
}

// Retreive all torrents
func (c *Client) GetTorrents() (Trnts []*Torrent, err error) {
	Trnts = make([]*Torrent, 0)
	rows, err := c.db.QueryContext(context.Background(), `select * from torrent;`)
	if err != nil {
		return
	}

	for rows.Next() {
		var trnt Torrent
		var ih string
		err = rows.Scan(&ih, &trnt.Started, &trnt.AddedAt, &trnt.StartedAt)
		if err != nil {
			return Trnts, err
		}

		trnt.Infohash, err = MetafromHex(ih)
		if err != nil {
			return Trnts, err
		}
		Trnts = append(Trnts, &trnt)
	}

	return Trnts, rows.Err()
}
