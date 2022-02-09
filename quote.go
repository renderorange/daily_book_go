package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
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
		fmt.Printf("usage: %s [-d] [-m] <book number>\n\n", os.Args[0])

		fmt.Println("options:")
		flag.PrintDefaults()
	}

	debug := flag.Bool("d", false, "print more information during the run")
	manual := flag.Int("m", 0, "manually specify the book number")

	flag.Parse()

	log.SetFlags(0)

	catalog_fh, err := os.Open("catalog.txt")
	if err != nil {
		log.Fatalln("[error]", err)
	}

	scanner := bufio.NewScanner(catalog_fh)
	scanner.Split(bufio.ScanLines)

	var catalog []string
	for scanner.Scan() {
		catalog = append(catalog, scanner.Text())
	}

	catalog_fh.Close()

	download_error_count := 0
MAIN:
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
			log.Println("[debug] page_link:", page_link)
			log.Println("[debug] book_link:", book_link)
		}

		resp, err := http.Get(book_link)
		if err != nil {
			log.Fatalln("[error]", err)
		}

		if resp.StatusCode != 200 {
			download_error_count = download_error_count + 1
			log.Println("[error] download response was", resp.StatusCode, "-", number)
			if *manual != 0 {
				os.Exit(1)
			} else if download_error_count == 20 {
				log.Fatalln("[error] download limit (20) reached")
			} else {
				continue
			}
		}
		defer resp.Body.Close()

		book_bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln("[error]", err)
		}

		scanner := bufio.NewScanner(strings.NewReader(string(book_bytes)))
		scanner.Split(bufio.ScanLines)

		var title, author string
		var header, body, footer []string

		_head, _body, _foot := 1, 0, 0
		if *debug {
			log.Println("[debug] parser is in head")
		}

	PARSE:
		for scanner.Scan() {
			line := scanner.Text()

			// despite not being as concise, the performance of 2 Contains is still
			// better than 1 MatchString
			if strings.Contains(line, "START OF THE PROJECT") || strings.Contains(line, "START OF THIS PROJECT") {
				if *debug {
					log.Println("[debug] parser is in body")
				}
				_head = 0
				_body = 1
				continue PARSE
			}

			if strings.Contains(line, "END OF THE PROJECT") || strings.Contains(line, "END OF THIS PROJECT") {
				if *debug {
					log.Println("[debug] parser is in footer")
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

		for _, line := range header {
			if strings.Contains(line, "The New McGuffey") {
				log.Println("[info] ebook is The New McGuffey Reader -", number)
				if *manual != 0 {
					os.Exit(0)
				}
				continue MAIN
			}

			if strings.Contains(line, "Language:") {
				if strings.Contains(line, "English") == false {
					log.Println("[info] ebook isn't in English -", number)
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
						log.Println("[debug] title:", title)
					}
				}
			}

			if strings.Contains(line, "Author:") {
				author_regex, _ := regexp.Compile(`^Author:\s+(.+)`)
				match := author_regex.FindStringSubmatch(line)
				if len(match) == 2 {
					author = match[1]
					if *debug {
						log.Println("[debug] author:", author)
					}
				}
			}
		}

		if len(title) == 0 {
			log.Println("[info] title was not found -", number)
			if *manual != 0 {
				os.Exit(0)
			}
			continue MAIN

		}

		if len(author) == 0 {
			log.Println("[info] author was not found -", number)
			if *manual != 0 {
				os.Exit(0)
			}
			continue MAIN
		}

		var build_variable string
		var paragraphs []string
		for _, line := range body {
			if len(line) != 0 {
				build_variable = build_variable + line + " "
			} else {
				paragraphs = append(paragraphs, build_variable)
				build_variable = ""
			}
		}

		if *debug {
			log.Println("[debug] paragraphs found:", len(paragraphs))
		}

		var quotes []string
		for _, paragraph := range paragraphs {
			quote_regex, _ := regexp.Compile(`^["].+["]\s*$`)
			if quote_regex.MatchString(paragraph) {
				if len(paragraph) > 90 && len(paragraph) < 113 {
					quotes = append(quotes, paragraph)
					if *debug {
						log.Println("[debug] quote was found:", paragraph)
					}
				}
			}
		}

		if len(quotes) == 0 {
			log.Println("[info] quote was not found -", number)
			if *manual != 0 {
				os.Exit(0)
			}
			continue MAIN
		}

		// TODO: select random from quotes if > 1 quote is found
		fmt.Printf("\ntitle: %s\nauthor: %s\n\n%s %s\n\n", title, author, quotes[len(quotes)-1], page_link)
		break
	}
}
