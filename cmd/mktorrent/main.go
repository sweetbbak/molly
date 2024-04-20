package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/inhies/go-bytesize"
	"github.com/jessevdk/go-flags"
)

var args struct {
	AnnounceList      []string `short:"a" description:"extra announce-list tier entry"`
	EmptyAnnounceList bool     `short:"n" description:"exclude default announce-list entries"`
	Comment           string   `short:"t" description:"comment"`
	CreatedBy         string   `short:"c" description:"created by"`
	OutFile           string   `short:"o" description:"output file name"`
	InfoName          *string  `short:"i" description:"override info short (defaults to ROOT)"`
	Url               []string `short:"u" description:"add webseed url"`
	Private           *bool    `short:"p" description:"private"`
	Root              string   `short:"d" description:"root of directory"`
	PieceLength       string   `short:"P" description:"piece length, (default is auto-calculated) ie: [128kb, 256kb, 512kb, 1mb, 2mb, 4mb, 8mb...]"`
}

var builtinAnnounceList = [][]string{
	{"http://p4p.arenabg.com:1337/announce"},
	{"udp://tracker.opentrackr.org:1337/announce"},
	{"udp://tracker.openbittorrent.com:6969/announce"},
}

func Mktorrent() error {
	bs, err := bytesize.Parse(args.PieceLength)
	if err != nil {
		return err
	}
	pl := int64(bs)

	mi := metainfo.MetaInfo{
		AnnounceList: builtinAnnounceList,
	}
	if args.EmptyAnnounceList {
		mi.AnnounceList = make([][]string, 0)
	}
	for _, a := range args.AnnounceList {
		mi.AnnounceList = append(mi.AnnounceList, []string{a})
	}
	mi.SetDefaults()
	if len(args.Comment) > 0 {
		mi.Comment = args.Comment
	}
	if len(args.CreatedBy) > 0 {
		mi.CreatedBy = args.CreatedBy
	}
	mi.UrlList = args.Url

	info := metainfo.Info{
		// PieceLength: pl,
		Private: args.Private,
	}
	if args.PieceLength != "" {
		info.PieceLength = pl
	}
	err = info.BuildFromFilePath(args.Root)
	if err != nil {
		return err
	}
	if args.InfoName != nil {
		info.Name = *args.InfoName
	}
	mi.InfoBytes, err = bencode.Marshal(info)
	if err != nil {
		return err
	}

	if args.OutFile != "" {
		fi, err := os.Create(args.OutFile)
		if err != nil {
			log.Fatal(err)
		}
		err = mi.Write(fi)
	} else {
		err = mi.Write(os.Stdout)
	}
	err = pprintMetainfo(&mi)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func pprintMetainfo(metainfo *metainfo.MetaInfo) error {
	info, err := metainfo.UnmarshalInfo()
	if err != nil {
		return fmt.Errorf("error unmarshalling info: %s", err)
	}
	d := map[string]interface{}{
		"Name":         info.Name,
		"Name.Utf8":    info.NameUtf8,
		"NumPieces":    info.NumPieces(),
		"PieceLength":  info.PieceLength,
		"InfoHash":     metainfo.HashInfoBytes().HexString(),
		"NumFiles":     len(info.UpvertedFiles()),
		"TotalLength":  info.TotalLength(),
		"Announce":     metainfo.Announce,
		"AnnounceList": metainfo.AnnounceList,
		"UrlList":      metainfo.UrlList,
	}
	if len(metainfo.Nodes) > 0 {
		d["Nodes"] = metainfo.Nodes
	}

	d["Files"] = info.UpvertedFiles()

	d["PieceHashes"] = func() (ret []string) {
		for i := 0; i < info.NumPieces(); i++ {
			ret = append(ret, hex.EncodeToString(info.Pieces[i*20:(i+1)*20]))
		}
		return
	}()

	b, _ := json.MarshalIndent(d, "", "  ")
	_, err = os.Stdout.Write(b)
	return err
}

func main() {
	_, err := flags.Parse(&args)
	if flags.WroteHelp(err) {
		os.Exit(0)
	}
	if err != nil {
		log.Fatal(err)
	}

	if err := Mktorrent(); err != nil {
		log.Fatal(err)
	}
}
