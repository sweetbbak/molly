package main

import (
	"fmt"
	"log"
	"os"
	// "time"

	"molly/pkg/tor"
)

func molly() error {
	client := tor.NewClient()
	client.Name = "molly"

	err := client.NewSession()
	if err != nil {
		return err
	}

	t, err := client.AddTorrent(os.Args[1])
	if err != nil {
		return err
	}

	t.DownloadAll()
	fmt.Println(t.NumPieces())

	// for !t.Complete.Bool() {
	// 	vv := tor.Veri(t)
	// 	fmt.Print("\x1b[2K\r")
	// 	fmt.Printf("pieces [%s]\n", vv)
	// 	time.Sleep(time.Millisecond * 99)
	// }
	return nil
}

func main() {
	if err := molly(); err != nil {
		log.Fatal(err)
	}
}
