package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var challengeDotGovUrl = "https://www.challenge.gov/"

type challenge struct {
	Name       string    `json:"name"`
	Agency     string    `json:"agency"`
	Summary    string    `json:"summary"`
	Deadline   time.Time `json:"deadline"`
	DetailsUrl string    `json:"details_url"`
	PubDate    time.Time `json:"pub_date"`
}

var loc *time.Location

func init() {
	var err error
	loc, err = time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}
}

func parse(r io.Reader) ([]challenge, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("parsing HTML doc: %w", err)
	}
	var challenges []challenge
	var hadError error
	doc.Find("#main-content section.usa-section .card").Each(func(i int, s *goquery.Selection) {
		if hadError != nil {
			return
		}
		var c challenge
		nameNode := s.Find("h3")
		c.Name = nameNode.Text()
		c.Agency = nameNode.Prev().Text()
		c.Summary = nameNode.Next().Text()
		{
			var err error
			deadline := strings.TrimSpace(nameNode.Parent().Next().Text())
			deadline = strings.TrimPrefix(deadline, "Open Until: ")
			tz := "ET"
			if !strings.HasSuffix(deadline, tz) {
				hadError = fmt.Errorf("expected timezone %q only, found %q", tz, deadline[strings.LastIndex(deadline, " ")+1:])
				return
			}
			deadline = strings.TrimSuffix(deadline, " ET")
			c.Deadline, err = time.ParseInLocation("01/02/2006 15:04 PM", deadline, loc)
			if err != nil {
				hadError = err
				return
			}
		}
		c.DetailsUrl = nameNode.Parent().Next().Next().Find("a").AttrOr("href", "")
		if !(strings.HasPrefix(c.DetailsUrl, "http://") || strings.HasPrefix(c.DetailsUrl, "https://")) && c.DetailsUrl[0] == '/' {
			c.DetailsUrl = challengeDotGovUrl + c.DetailsUrl[1:]
		}
		challenges = append(challenges, c)
	})
	if hadError != nil {
		return nil, fmt.Errorf("during HTML query: %w", hadError)
	}
	return challenges, nil
}

func scrape() ([]challenge, error) {
	resp, err := http.Get(challengeDotGovUrl)
	if err != nil {
		return nil, fmt.Errorf("getting challenge.gov homepage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error getting challenge.gov homepage: %d %s", resp.StatusCode, resp.Status)
	}

	challenges, err := parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML document: %w", err)
	}

	return challenges, nil
}

func exists(pathname string) bool {
	_, err := os.Stat(pathname)
	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		panic(err)
	}
	return true
}

func main() {
	flag.Parse()
	if flag.NArg() < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <command>\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "commands\n")
		fmt.Fprintf(os.Stderr, "  - json /path/to/challenges.json\n")
		fmt.Fprintf(os.Stderr, "  - rss /path/to/challenges.json\n")
		os.Exit(1)
	}
	switch flag.Arg(0) {
	case "json":
		outputFile := flag.Arg(1)
		var existing []challenge
		if exists(outputFile) {
			f, err := os.Open(outputFile)
			if err != nil {
				log.Fatalf("opening %q: %v", outputFile, err)
			}
			defer f.Close()
			dec := json.NewDecoder(f)
			if err := dec.Decode(&existing); err != nil {
				log.Fatalf("error reading JSON: %v", err)
			}
		}
		lookup := make(map[string]challenge)
		for _, c := range existing {
			lookup[c.DetailsUrl] = c
		}
		scrapeTs := time.Now()
		challenges, err := scrape()
		if err != nil {
			log.Fatalf("error scraping: %v", err)
		}
		for i, c := range challenges {
			if old, ok := lookup[c.DetailsUrl]; ok && !old.PubDate.IsZero() {
				challenges[i].PubDate = old.PubDate
			} else {
				challenges[i].PubDate = scrapeTs
			}
		}
		tmpfile, err := ioutil.TempFile("", "challenges-*.json")
		if err != nil {
			log.Fatalf("opening temp file: %v", err)
		}
		enc := json.NewEncoder(tmpfile)
		enc.SetIndent("", "\t")
		if err := enc.Encode(challenges); err != nil {
			log.Fatalf("error writing JSON: %v", err)
		}
		if err := tmpfile.Close(); err != nil {
			log.Fatalf("closing temp file: %v", err)
		}
		if err := os.Rename(tmpfile.Name(), outputFile); err != nil {
			log.Fatalf("replacing existing JSON file with new generated output: %v", err)
		}
	case "rss":
		var cs []challenge
		f, err := os.Open(flag.Arg(1))
		if err != nil {
			log.Fatalf("opening JSON file: %v", err)
		}
		defer f.Close()
		dec := json.NewDecoder(f)
		if err := dec.Decode(&cs); err != nil {
			log.Fatalf("error reading JSON: %v", err)
		}
		if err := challenges(cs).emitRssFeed(os.Stdout); err != nil {
			log.Fatalf("error writing RSS: %v", err)
		}
	default:
		log.Fatalf("unknown argument %q", flag.Arg(0))
	}
}
