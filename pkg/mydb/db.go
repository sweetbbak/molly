package mydb

import (
	"context"
	"database/sql"
	"fmt"
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

var db *sql.DB

func Init(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(context.Background(), `create table if not exists torrent (infohash text primary key,started boolean,addedat timestamptz,startedat timestamptz);`)
	if err != nil {
		return err
	}
	return nil
}

func Close() {
	db.Close()
}

func Exists(ih metainfo.Hash) (ret bool) {
	var err error
	row := db.QueryRowContext(context.Background(), `select true from torrent where infohash=$1;`, ih.HexString())
	err = row.Scan(&ret)
	if err != nil {
		return false
	}
	return
}

func HasStarted(ih string) (ret bool) {
	row := db.QueryRowContext(context.Background(), `select started from torrent where infohash=$1;`, ih)
	err := row.Scan(&ret)
	if err != nil {
		return false
	}
	return
}

func Add(ih metainfo.Hash) (err error) {
	_, err = db.ExecContext(context.Background(), `insert into torrent (infohash,started,addedat,startedat) values ($1,$2,$3,$4) on conflict (infohash) do update set startedat=$4;`, ih.HexString(), false, time.Now(), time.Now())
	return
}

func Delete(ih metainfo.Hash) (err error) {
	_, err = db.ExecContext(context.Background(), `delete from torrent where infohash=$1;`, ih.HexString())
	return
}

func Start(ih metainfo.Hash) (err error) {
	_, err = db.ExecContext(context.Background(), `update torrent set started=$1,startedat=$2 where infohash=$3;`, true, time.Now(), ih.HexString())
	return
}

func SetStarted(ih metainfo.Hash, inp bool) (err error) {
	_, err = db.ExecContext(context.Background(), `update torrent set started=$1 where infohash=$2;`, inp, ih.HexString())
	return
}

func GetTorrent(ih metainfo.Hash) (*Torrent, error) {
	var trnt Torrent
	var infoh string
	row := db.QueryRowContext(context.Background(), `select * from torrent where infohash=$1;`, ih.HexString())
	err := row.Scan(&infoh, &trnt.Started, &trnt.AddedAt, &trnt.StartedAt)
	if err != nil {
		return nil, err
	}
	trnt.Infohash = ih
	return &trnt, nil
}

func GetTorrents() (Trnts []*Torrent, err error) {
	Trnts = make([]*Torrent, 0)
	rows, err := db.QueryContext(context.Background(), `select * from torrent;`)
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
