package main

import (
	"bufio"
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

	catalog_fh, err := os.Open("catalog.txt")
	if err != nil {
		fmt.Print(err, "\n")
		os.Exit(2)
	}

	scanner := bufio.NewScanner(catalog_fh)
	scanner.Split(bufio.ScanLines)

	var catalog []string
	for scanner.Scan() {
		catalog = append(catalog, scanner.Text())
	}

	catalog_fh.Close()

	// TODO: load env file for twitter settings

	if *quiet == false && *manual == 0 {
		fmt.Print("finding a quote, just a moment\n")
		fmt.Print("for more information, please see quote.log\n\n")
	}
}
