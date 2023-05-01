package main

import "flag"

func main() {
	filepath := flag.String("path", "undefined", "filepath to source")
	flag.Parse()
	_ = filepath
}
