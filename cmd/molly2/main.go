package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"molly/pkg/tor"

	"github.com/anacrolix/torrent/metainfo"
)

func StartMolly() (*tor.Client, error) {
	client := tor.NewClient()
	client.Name = "molly"

	err := client.NewSession()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// StringPrompt asks for a string value using the label
func StringPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

func PrintHelp() {
	fmt.Println("commands: help, exit, list, add, remove, pause, start")
	fmt.Println("add [args] (examples: magnet link, url, /path/to/file.torrent)")
	fmt.Println("stop [infohash]\n")
}

func SimplePrompt(client *tor.Client) {
	fmt.Println("Molly: A CLI torrent client (type \"help\" for commands)")

	for {
		input := StringPrompt("> ")
		argsx := strings.Split(input, " ")
		op := argsx[0]

		var args []string
		if len(argsx) > 1 {
			args = argsx[1:]
		}

		fmt.Printf("INPUT: op [%s] args [%s]\n", op, args)

		switch op {
		case "help":
			PrintHelp()
		case "exit":
			fmt.Println("exiting...")
			return
		case "list":
			ts := client.ShowTorrents()
			if len(ts) == 0 || ts == nil {
				fmt.Println("no torrents to show")
			}
			for _, t := range ts {
				fmt.Printf("%s | started: %v | %s\n", t.Name(), client.DBHasStarted(t.InfoHash().String()), t.InfoHash().String())
			}
		case "add":
			for _, a := range args {
				t, err := client.AddTorrent(a)
				if err != nil {
					log.Println(err)
				} else {
					fmt.Printf("added torrent [%s]", t.Name())
				}
			}
		case "remove", "delete":
			if len(args) == 0 {
				log.Println("error: action requires argument [infohash]")
			}
			for _, a := range args {
				ih := metainfo.NewHashFromHex(a)
				err := client.StopTorrent(ih)
				if err != nil {
					log.Println(err)
				}
			}
		case "pause", "stop":
			if len(args) == 0 {
				log.Println("error: action requires argument [infohash]")
			}
			for _, a := range args {
				ih := metainfo.NewHashFromHex(a)
				err := client.StopTorrent(ih)
				if err != nil {
					log.Println(err)
				}
			}
		case "start":
			if len(args) == 0 {
				log.Println("error: action requires argument [infohash]")
			}
			for _, a := range args {
				ih := metainfo.NewHashFromHex(a)
				t, new := client.TorrentClient.AddTorrentInfoHash(ih)
				if !new {
					log.Println("restarted torrent [%s]", t.Name())
				} else {
					log.Println("added new torrent [%s]", t.Name())
				}
			}
		default:
			fmt.Println("unknown command")
		}
	}
}

func molly() error {
	client, err := StartMolly()
	if err != nil {
		return err
	}

	SimplePrompt(client)
	return nil
}

func breaker() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	for {
		<-c
		println("canceling...")
		os.Exit(0)
	}
}

func main() {
	// handle Ctrl+C
	go breaker()

	if err := molly(); err != nil {
		log.Fatal(err)
	}
}
