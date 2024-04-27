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
	return nil
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

			// TODO: check if this range is inclusive or not
			for x := n1; x <= n2; x++ {
				idxs = append(idxs, x)
			}
		} else {
			n, err := strconv.Atoi(i)
			if err != nil {
				return idxs, err
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
