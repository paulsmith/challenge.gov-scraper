package main

import (
	"encoding/xml"
	"fmt"
	"io"
)

type rssEntry struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
}

type rss20Feed struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel struct {
		Title       string     `xml:"title"`
		Link        string     `xml:"link"`
		Description string     `xml:"description"`
		Entries     []rssEntry `xml:"item"`
	} `xml:"channel"`
}

type challenges []challenge

func (cs challenges) emitRssFeed(w io.Writer) error {
	var feed rss20Feed
	feed.Version = "2.0"
	feed.Channel.Title = "Unofficial Challenge.gov challenges feed"
	feed.Channel.Link = "https://paulsmith.github.io/challenge.gov-scraper/challenges.rss"
	feed.Channel.Description = "Unofficial Challenge.gov challenges feed"
	for i := range cs {
		var entry rssEntry
		entry.Title = cs[i].Name + " · " + cs[i].Agency
		entry.Link = cs[i].DetailsUrl
		entry.Description = fmt.Sprintf("Summary: %s\n\nDeadline: %v", cs[i].Summary, cs[i].Deadline)
		feed.Channel.Entries = append(feed.Channel.Entries, entry)
	}
	fmt.Fprintf(w, "%s", xml.Header)
	enc := xml.NewEncoder(w)
	enc.Indent("", "\t")
	if err := enc.Encode(feed); err != nil {
		return fmt.Errorf("encoding XML: %w", err)
	}
	return nil
}
