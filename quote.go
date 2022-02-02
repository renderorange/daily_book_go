package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
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

	// main process loop
	for {
		var number string

		if *manual != 0 {
			number = strconv.Itoa(*manual)
		} else {
			rand.Seed(time.Now().UnixNano())
			max := len(catalog) - 1
			rand := rand.Intn(max + 1)
			number = catalog[rand]
		}

		file := number + ".txt"
		page_link := "https://gutenberg.org/ebooks/" + number
		book_link := "https://gutenberg.pglaf.org"

		if len(number) == 1 {
			book_link = book_link + "/0/" + number + "/" + number + ".txt"
		} else {
			for i := 0; i <= len(number)-2; i++ {
				book_link = book_link + "/" + string(number[i])
			}

			book_link = book_link + "/" + number + "/" + number + ".txt"
		}
	}
}
