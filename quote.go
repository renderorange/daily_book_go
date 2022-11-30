package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
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
	debug   bool
	manual  int
	twitter bool
}

func init() {
	flag.Usage = func() {
		fmt.Printf("usage: %s [-d] [-m] <book number> [-t]\n\n", os.Args[0])

		fmt.Println("options:")
		flag.PrintDefaults()
	}

	flag.BoolVar(&opts.debug, "d", false, "print more information during the run")
	flag.IntVar(&opts.manual, "m", 0, "manually specify the book number")
	flag.BoolVar(&opts.twitter, "t", false, "post the quote to twitter")
}

func getTimestamp() string {
	now := time.Now()
	sec := now.Unix()
	return strconv.Itoa(int(sec))
}

func parse(book string) ([]string, []string, []string) {
	scanner := bufio.NewScanner(strings.NewReader(book))
	scanner.Split(bufio.ScanLines)

	var header, body, footer []string
	markHead, markBody, markFoot := 1, 0, 0

	if opts.debug {
		log.Println("[" + getTimestamp() + "] [debug] parser is in head")
	}

	for scanner.Scan() {
		line := scanner.Text()

		// despite not being as concise, the performance of 2 Contains is still
		// better than 1 MatchString
		if strings.Contains(line, "START OF THE PROJECT") || strings.Contains(line, "START OF THIS PROJECT") {
			if opts.debug {
				log.Println("[" + getTimestamp() + "] [debug] parser is in body")
			}
			markHead = 0
			markBody = 1
			continue
		}

		if strings.Contains(line, "END OF THE PROJECT") || strings.Contains(line, "END OF THIS PROJECT") {
			if opts.debug {
				log.Println("[" + getTimestamp() + "] [debug] parser is in footer")
			}
			markBody = 0
			markFoot = 1
			continue
		}

		// remove whitespace at the start and end of lines
		line = strings.Trim(line, " ")

		// correct double spacing
		doubleSpacingRegex, _ := regexp.Compile(`\s{2}`)
		line = doubleSpacingRegex.ReplaceAllString(line, " ")

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

func process(header, body []string) (string, string, []string, error) {
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
					log.Println("["+getTimestamp()+"] [debug] title:", title)
				}
			}
		}

		if strings.Contains(line, "Author:") {
			authorRegex, _ := regexp.Compile(`^Author:\s+(.+)`)
			match := authorRegex.FindStringSubmatch(line)
			if len(match) == 2 {
				author = match[1]
				if opts.debug {
					log.Println("["+getTimestamp()+"] [debug] author:", author)
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
		log.Println("["+getTimestamp()+"] [debug] paragraphs found:", len(paragraphs))
	}

	for _, paragraph := range paragraphs {
		quoteRegex, _ := regexp.Compile(`^["].+["]\s*$`)
		if quoteRegex.MatchString(paragraph) {
			if len(paragraph) > 90 && len(paragraph) < 113 {
				quotes = append(quotes, paragraph)
				if opts.debug {
					log.Println("["+getTimestamp()+"] [debug] quote was found:", paragraph)
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

	var twitterCredentials struct {
		consumerKey       string
		consumerSecret    string
		accessToken       string
		accessTokenSecret string
	}

	if opts.twitter {
		twitterCredentials.consumerKey = os.Getenv("TWITTER_CONSUMER_KEY")
		twitterCredentials.consumerSecret = os.Getenv("TWITTER_CONSUMER_SECRET")
		twitterCredentials.accessToken = os.Getenv("TWITTER_ACCESS_TOKEN")
		twitterCredentials.accessTokenSecret = os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")

		if opts.debug {
			log.Println("["+getTimestamp()+"] [debug] twitter consumer key:", twitterCredentials.consumerKey)
			log.Println("["+getTimestamp()+"] [debug] twitter consumer secret:", twitterCredentials.consumerSecret)
			log.Println("["+getTimestamp()+"] [debug] twitter access token:", twitterCredentials.accessToken)
			log.Println("["+getTimestamp()+"] [debug] twitter access token secret:", twitterCredentials.accessTokenSecret)
		}

		if twitterCredentials.consumerKey == "" || twitterCredentials.consumerSecret == "" || twitterCredentials.accessToken == "" || twitterCredentials.accessTokenSecret == "" {
			log.Fatalln("[" + getTimestamp() + "] [error] twitter API credentials are not configured")
		}
	}

	catalogFh, err := os.Open("catalog.txt")
	if err != nil {
		log.Fatalln("["+getTimestamp()+"] [error]", err)
	}

	var catalog []string
	scanner := bufio.NewScanner(catalogFh)
	for scanner.Scan() {
		catalog = append(catalog, scanner.Text())
	}
	catalogFh.Close()

	downloadErrorCount := 0
	for {
		var number string
		var book string

		if opts.manual != 0 {
			number = strconv.Itoa(opts.manual)
			for _, v := range catalog {
				if v == number+".txt" || v == number+"-0.txt" {
					book = v
					break
				}
			}
			if book == "" {
				log.Fatalln("["+getTimestamp()+"] [error] book not found -", number)
			}
		} else {
			rand.Seed(time.Now().UnixNano())
			book = catalog[rand.Intn(len(catalog)-1)]
			numberRegex, _ := regexp.Compile(`^(\d+)`)
			matches := numberRegex.FindStringSubmatch(book)
			number = matches[0]
		}

		pageLink := "https://gutenberg.org/ebooks/" + number
		bookLink := "https://aleph.pglaf.org"

		if len(number) == 1 {
			bookLink = bookLink + "/0/" + number + "/" + book
		} else {
			for i := 0; i <= len(number)-2; i++ {
				bookLink = bookLink + "/" + string(number[i])
			}

			bookLink = bookLink + "/" + number + "/" + book
		}

		if opts.debug {
			log.Println("["+getTimestamp()+"] [debug] page link:", pageLink)
			log.Println("["+getTimestamp()+"] [debug] book link:", bookLink)
		}

		resp, err := http.Get(bookLink)
		if err != nil {
			log.Fatalln("["+getTimestamp()+"] [error]", err)
		}

		if resp.StatusCode != 200 {
			downloadErrorCount = downloadErrorCount + 1
			log.Println("["+getTimestamp()+"] [error] download response was", resp.StatusCode, "-", number)
			if opts.manual != 0 {
				os.Exit(1)
			} else if downloadErrorCount == 20 {
				log.Fatalln("[" + getTimestamp() + "] [error] download limit (20) reached")
			} else {
				continue
			}
		}
		defer resp.Body.Close()

		bookBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln("["+getTimestamp()+"] [error]", err)
		}

		// no data is yet needed from the parsed footer, so don't
		// store the return.
		header, body, _ := parse(string(bookBytes))

		title, author, quotes, err := process(header, body)
		if err != nil {
			log.Println("["+getTimestamp()+"]", err, "-", number)
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

		quote := quotes[quoteIndex]

		fmt.Printf("\ntitle: %s\nauthor: %s\n\n%s %s\n\n", title, author, quote, pageLink)

		if opts.twitter {
			if opts.debug {
				log.Println("[" + getTimestamp() + "] [debug] posting to twitter")
			}

			config := oauth1.NewConfig(twitterCredentials.consumerKey, twitterCredentials.consumerSecret)
			token := oauth1.NewToken(twitterCredentials.accessToken, twitterCredentials.accessTokenSecret)
			httpClient := config.Client(oauth1.NoContext, token)
			client := twitter.NewClient(httpClient)

			_, _, err := client.Statuses.Update(quote+" "+pageLink, nil)
			if err != nil {
				log.Fatalln("["+getTimestamp()+"] [error]", err)
			}
		}

		break
	}
}
