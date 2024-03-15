package main

import (
	"fmt"
	"log"

	"github.com/anacrolix/torrent"
)

func molly() error {
	return nil
}

func main() {
	fmt.Println("Hello, world!")

	if err := molly(); err != nil {
		log.Fatal(err)
	}
}
