package main

import "github.com/simon-xia/fzf/src"

var revision string

func main() {
	fzf.Run(fzf.ParseOptions(), revision)
}
