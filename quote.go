package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Usage = func() {
		fmt.Printf("usage: %s [-t] [-q] [-m] <book number>\n\n", os.Args[0])

		fmt.Print("options:\n")
		flag.PrintDefaults()
	}

	twitter := flag.Bool("t", false, "post the quote to twitter (requires additional configuration)")
	quiet := flag.Bool("q", false, "don't display any output")
	manual := flag.Int("m", 0, "manually specify the book number")

	flag.Parse()

	if *twitter == false && *quiet == true {
		fmt.Print("flag -q implies -t\n")
		flag.Usage()
		os.Exit(2)
	}
}
