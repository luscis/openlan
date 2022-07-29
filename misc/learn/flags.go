package main

import (
	"flag"
	"os"
)

func main() {
	var alias string

	var cl0 = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.StringVar(&alias, "alias", "", "the alias for this point")
	cl0.Var(&alias, name, usage)
	cl0.Parse(os.Args[1:])

	print(alias)
}
