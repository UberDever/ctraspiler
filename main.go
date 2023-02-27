package main

import (
	"flag"
)

func main() {
	filepath := flag.String("path", "undefined", "filepath")
	flag.Parse()
	_ = filepath
}