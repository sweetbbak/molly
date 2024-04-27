// This file is explicitly for selecting files to be downloaded after adding a torrent
package tor

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/anacrolix/torrent"
)

// takes a slice of indexes for files to be added
func (c *Client) SelectFileByIndex(t *torrent.Torrent, idxs []int) error {
	files := t.Files()
	for _, i := range idxs {
		if i > len(files) {
			return fmt.Errorf("file index out of range: index [%d] of file count [%d]\n", i, len(files))
		}

		x := files[i]
		x.Download()
	}

	return nil
}

func (c *Client) Info(tor string, client *torrent.Client) error {
	t, err := c.AddTorrent(tor)
	if err != nil {
		return err
	}

	printInfo(t)

	if !c.DBExists(t.InfoHash()) {
		t.Drop()
	}

	return nil
}

func printInfo(t *torrent.Torrent) {
	info := t.Info()
	files := t.Files()
	sz := info.TotalLength()
	psz := prettyByteSize(int(sz))
	fmt.Printf("%s\n", psz)

	for i, f := range files {
		fmt.Printf("%d | %s\n", i, f.DisplayPath())
	}
}

// remove duplicate entries from slice
func removeDuplicate[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

// takes aria2c like strings ex: --select-file=0,1,3,30-33
func ParseIndices(s string) ([]int, error) {
	var idxs []int
	cuts := strings.Split(s, ",")

	if len(cuts) <= 0 {
		return idxs, fmt.Errorf("unable to parse file index string, length is 0")
	}

	for _, i := range cuts {
		if i == "" {
			return idxs, fmt.Errorf("error: file index is an empty string")
		}

		if strings.Contains(i, "-") {
			splits := strings.Split(i, "-")

			if len(splits) != 2 {
				return idxs, fmt.Errorf("index file range can only contain two numbers")
			}

			if splits[0] > splits[1] {
				return idxs, fmt.Errorf("index file range must be in format [1-99] where N1 is less than N2, got: [%s]", i)
			}

			n1, err := strconv.Atoi(splits[0])
			if err != nil {
				return idxs, err
			}

			n2, err := strconv.Atoi(splits[1])
			if err != nil {
				return idxs, err
			}

			if n1 < 0 || n2 < 0 {
				return idxs, fmt.Errorf("file index cannot be negative, got range [%d] [%d]", n1, n2)
			}

			// TODO: check if this range is inclusive or not
			for x := n1; x <= n2; x++ {
				idxs = append(idxs, x)
			}
		} else {
			n, err := strconv.Atoi(i)
			if err != nil {
				return idxs, err
			}

			if n < 0 {
				return idxs, fmt.Errorf("file index cannot be negative, got [%d]", n)
			}

			idxs = append(idxs, n)
		}
	}

	if len(idxs) <= 0 {
		return idxs, fmt.Errorf("unable to parse file indices")
	}

	// sort by value
	sort.Ints(idxs)
	idxs = removeDuplicate(idxs)

	return idxs, nil
}
