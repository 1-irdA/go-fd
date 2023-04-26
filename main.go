package main

import (
	"flag"

	"github.com/garrou/fd/lib"
)

func main() {
	hidden := flag.Bool("i", false, "search hidden entries (default: false)")
	recurse := flag.Bool("r", false, "find recursively (default: false)")
	count := flag.Bool("c", false, "count the number of results (default: false)")
	pattern := flag.Bool("p", false, "search pattern (default: false)")
	routine := flag.Uint("n", 50, "number of routines (default: 50, min: 10, max: 100)")
	flag.Parse()

	walker := lib.NewWalker(flag.Args(), *routine, *recurse, *hidden, *count, *pattern)
	walker.Search()
}
