package main

import (
	"bufio"
	"errors"
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

var opts struct {
	debug  bool
	manual int
}

func init() {
	flag.Usage = func() {
		fmt.Printf("usage: %s [-d] [-m] <book number>\n\n", os.Args[0])

		fmt.Println("options:")
		flag.PrintDefaults()
	}

	flag.BoolVar(&opts.debug, "d", false, "print more information during the run")
	flag.IntVar(&opts.manual, "m", 0, "manually specify the book number")
}

func parse(debug *bool, book string) ([]string, []string, []string) {
	scanner := bufio.NewScanner(strings.NewReader(book))
	scanner.Split(bufio.ScanLines)

	var header, body, footer []string
	markHead, markBody, markFoot := 1, 0, 0

	if opts.debug {
		log.Println("[debug] parser is in head")
	}

	for scanner.Scan() {
		line := scanner.Text()

		// despite not being as concise, the performance of 2 Contains is still
		// better than 1 MatchString
		if strings.Contains(line, "START OF THE PROJECT") || strings.Contains(line, "START OF THIS PROJECT") {
			if opts.debug {
				log.Println("[debug] parser is in body")
			}
			markHead = 0
			markBody = 1
			continue
		}

		if strings.Contains(line, "END OF THE PROJECT") || strings.Contains(line, "END OF THIS PROJECT") {
			if opts.debug {
				log.Println("[debug] parser is in footer")
			}
			markBody = 0
			markFoot = 1
			continue
		}

		// remove whitespace at the start and end of lines
		line = strings.Trim(line, " ")

		// correct double spacing
		double_spacing_regex, _ := regexp.Compile(`\s{2}`)
		line = double_spacing_regex.ReplaceAllString(line, " ")

		if markHead == 1 {
			header = append(header, line)
		} else if markBody == 1 {
			body = append(body, line)
		} else if markFoot == 1 {
			footer = append(footer, line)
		}
	}

	return header, body, footer
}

func process(debug *bool, header, body []string) (string, string, []string, error) {
	var title, author string
	var quotes []string

	for _, line := range header {
		if strings.Contains(line, "The New McGuffey") {
			return "", "", nil, errors.New("[info] ebook is The New McGuffey Reader")
		}

		if strings.Contains(line, "Language:") {
			if strings.Contains(line, "English") == false {
				return "", "", nil, errors.New("[info] ebook isn't in English")
			}
		}

		if strings.Contains(line, "Title:") {
			languageRegex, _ := regexp.Compile(`^Title:\s+(.+)`)
			match := languageRegex.FindStringSubmatch(line)
			if len(match) == 2 {
				title = match[1]
				if opts.debug {
					log.Println("[debug] title:", title)
				}
			}
		}

		if strings.Contains(line, "Author:") {
			authorRegex, _ := regexp.Compile(`^Author:\s+(.+)`)
			match := authorRegex.FindStringSubmatch(line)
			if len(match) == 2 {
				author = match[1]
				if opts.debug {
					log.Println("[debug] author:", author)
				}
			}
		}
	}

	if len(title) == 0 {
		return "", "", nil, errors.New("[info] title was not found")

	}

	if len(author) == 0 {
		return "", "", nil, errors.New("[info] author was not found")
	}

	var buildVariable string
	var paragraphs []string
	for _, line := range body {
		if len(line) != 0 {
			buildVariable = buildVariable + line + " "
		} else {
			paragraphs = append(paragraphs, buildVariable)
			buildVariable = ""
		}
	}

	if opts.debug {
		log.Println("[debug] paragraphs found:", len(paragraphs))
	}

	for _, paragraph := range paragraphs {
		quoteRegex, _ := regexp.Compile(`^["].+["]\s*$`)
		if quoteRegex.MatchString(paragraph) {
			if len(paragraph) > 90 && len(paragraph) < 113 {
				quotes = append(quotes, paragraph)
				if opts.debug {
					log.Println("[debug] quote was found:", paragraph)
				}
			}
		}
	}

	if len(quotes) == 0 {
		return "", "", nil, errors.New("[info] quote was not found")
	}

	return title, author, quotes, nil
}

func main() {
	flag.Parse()

	log.SetFlags(0)

	catalog_fh, err := os.Open("catalog.txt")
	if err != nil {
		log.Fatalln("[error]", err)
	}

	var catalog []string
	scanner := bufio.NewScanner(catalog_fh)
	for scanner.Scan() {
		catalog = append(catalog, scanner.Text())
	}
	catalog_fh.Close()

	downloadErrorCount := 0
	for {
		var number string

		if opts.manual != 0 {
			number = strconv.Itoa(opts.manual)
		} else {
			rand.Seed(time.Now().UnixNano())
			number = catalog[rand.Intn(len(catalog)-1)]
		}

		pageLink := "https://gutenberg.org/ebooks/" + number
		bookLink := "https://gutenberg.pglaf.org"

		if len(number) == 1 {
			bookLink = bookLink + "/0/" + number + "/" + number + ".txt"
		} else {
			for i := 0; i <= len(number)-2; i++ {
				bookLink = bookLink + "/" + string(number[i])
			}

			bookLink = bookLink + "/" + number + "/" + number + ".txt"
		}

		if opts.debug {
			log.Println("[debug] page link:", pageLink)
			log.Println("[debug] book link:", bookLink)
		}

		resp, err := http.Get(bookLink)
		if err != nil {
			log.Fatalln("[error]", err)
		}

		if resp.StatusCode != 200 {
			downloadErrorCount = downloadErrorCount + 1
			log.Println("[error] download response was", resp.StatusCode, "-", number)
			if opts.manual != 0 {
				os.Exit(1)
			} else if downloadErrorCount == 20 {
				log.Fatalln("[error] download limit (20) reached")
			} else {
				continue
			}
		}
		defer resp.Body.Close()

		bookBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln("[error]", err)
		}

		// no data is yet needed from the parsed footer, so don't
		// store the return.
		header, body, _ := parse(&opts.debug, string(bookBytes))

		title, author, quotes, err := process(&opts.debug, header, body)
		if err != nil {
			log.Println(err, "-", number)
			if opts.manual != 0 {
				os.Exit(0)
			}
			continue
		}

		var quoteIndex int
		if len(quotes) > 1 {
			rand.Seed(time.Now().UnixNano())
			quoteIndex = rand.Intn(len(quotes) - 1)
		} else {
			quoteIndex = len(quotes) - 1
		}
		fmt.Printf("\ntitle: %s\nauthor: %s\n\n%s %s\n\n", title, author, quotes[quoteIndex], pageLink)
		break
	}
}
