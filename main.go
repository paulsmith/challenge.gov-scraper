package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

func main() {
	challenges, err := scrape()
	if err != nil {
		log.Fatalf("error scraping: %v", err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	if err := enc.Encode(challenges); err != nil {
		log.Fatalf("error writing JSON: %v", err)
	}
}
