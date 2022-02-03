package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() {
	flag.Usage = func() {
		fmt.Printf("usage: %s [-d] [-m] <book number> [-q] [-t]\n\n", os.Args[0])

		fmt.Print("options:\n")
		flag.PrintDefaults()
	}

	twitter := flag.Bool("t", false, "post the quote to twitter (requires additional configuration)")
	quiet := flag.Bool("q", false, "don't display any output")
	manual := flag.Int("m", 0, "manually specify the book number")
	debug := flag.Bool("d", false, "print more information during the run")

	flag.Parse()

	if *twitter == false && *quiet == true {
		fmt.Print("flag -q implies -t\n")
		flag.Usage()
		os.Exit(2)
	}

	if *quiet == true && *debug == true {
		fmt.Print("flag -q and -d are mutually exclusive\n")
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

MAIN:
	for {
		var number string
		download_error_count := 0

		if *manual != 0 {
			number = strconv.Itoa(*manual)
		} else {
			rand.Seed(time.Now().UnixNano())
			max := len(catalog) - 1
			rand := rand.Intn(max + 1)
			number = catalog[rand]
		}

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

		if *debug {
			fmt.Print("[debug] page_link: ", page_link, "\n")
			fmt.Print("[debug] book_link: ", book_link, "\n")
		}

		resp, err := http.Get(book_link)
		if err != nil {
			if *manual != 0 {
				fmt.Print(err)
				os.Exit(2)
			} else if download_error_count == 20 {
				fmt.Print("[error] download limit (20) exceeded\n")
				os.Exit(2)
			} else {
				download_error_count = download_error_count + 1
				fmt.Print(err)
				continue
			}
		}
		defer resp.Body.Close()

		book_bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Print(err)
			os.Exit(2)
		}

		scanner := bufio.NewScanner(strings.NewReader(string(book_bytes)))
		scanner.Split(bufio.ScanLines)

		var title, author string
		var header, body, footer []string

		_head, _body, _foot := 1, 0, 0
		if *debug {
			fmt.Print("[debug] parser is in head\n")
		}

	PARSE:
		for scanner.Scan() {
			line := scanner.Text()

			// despite not being as concise, the performance of 2 Contains is still
			// better than 1 MatchString
			if strings.Contains(line, "START OF THE PROJECT") || strings.Contains(line, "START OF THIS PROJECT") {
				if *debug {
					fmt.Print("[debug] parser is in body\n")
				}
				_head = 0
				_body = 1
				continue PARSE
			}

			if strings.Contains(line, "END OF THE PROJECT") || strings.Contains(line, "END OF THIS PROJECT") {
				if *debug {
					fmt.Print("[debug] parser is in footer\n")
				}
				_body = 0
				_foot = 1
				continue PARSE
			}

			// remove whitespace at the start and end of lines
			line = strings.Trim(line, " ")

			// correct double spacing
			double_spacing_regex, _ := regexp.Compile(`\s{2}`)
			line = double_spacing_regex.ReplaceAllString(line, " ")

			if _head == 1 {
				header = append(header, line)
			} else if _body == 1 {
				body = append(body, line)
			} else if _foot == 1 {
				footer = append(footer, line)
			}
		}

	PROCESS:
		for _, line := range header {
			if strings.Contains(line, "The New McGuffey") {
				fmt.Print("[info] ebook is The New McGuffey Reader - ", number, "\n")
				if *manual != 0 {
					os.Exit(0)
				}
				continue MAIN
			}

			if strings.Contains(line, "Language:") {
				if strings.Contains(line, "English") == false {
					fmt.Print("[info] ebook isn't in English - ", number, "\n")
					if *manual != 0 {
						os.Exit(0)
					}
					continue MAIN
				}
			}

			if strings.Contains(line, "Title:") {
				language_regex, _ := regexp.Compile(`^Title:\s+(.+)`)
				match := language_regex.FindStringSubmatch(line)
				if len(match) == 2 {
					title = match[1]
					if *debug {
						fmt.Print("[debug] title: ", title, "\n")
					}
				} else {
					fmt.Print("[error] there was an issue parsing title - ", number, "\n")
					if *debug {
						fmt.Print("[debug] title match: ", match, "\n")
					}
					if *manual != 0 {
						os.Exit(2)
					}
					continue MAIN
				}
			}

			if strings.Contains(line, "Author:") {
				author_regex, _ := regexp.Compile(`^Author:\s+(.+)`)
				match := author_regex.FindStringSubmatch(line)
				if len(match) == 2 {
					author = match[1]
					if *debug {
						fmt.Print("[debug] author: ", author, "\n")
					}
				} else {
					fmt.Print("[error] there was an issue parsing author - ", number, "\n")
					if *debug {
						fmt.Print("[debug] author match: ", match, "\n")
					}
					if *manual != 0 {
						os.Exit(2)
					}
					continue MAIN
				}
			}
		}

		if len(title) == 0 {
			fmt.Print("[info] title was not found - ", number, "\n")
			if *manual != 0 {
				os.Exit(0)
			}
			continue MAIN

		}

		if len(author) == 0 {
			fmt.Print("[info] author was not found - ", number, "\n")
			if *manual != 0 {
				os.Exit(0)
			}
			continue MAIN
		}
	}
}
