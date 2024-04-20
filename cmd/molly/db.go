package main

import (
	"molly/pkg/mydb"
)

func Start() error {
	return mydb.Init("test.sqlite")
}
