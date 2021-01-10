package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	snapshotPath := "snapshots/challenge.gov_20210105-192544.html"
	f, err := os.Open(snapshotPath)
	if err != nil {
		t.Fatalf("opening snapshot file %q: %v", snapshotPath, err)
	}
	defer f.Close()
	challenges, err := parse(f)
	if err != nil {
		t.Fatalf("parsing snapshot file %q: %v", snapshotPath, err)
	}
	want := 31
	if len(challenges) != want {
		t.Errorf("want %d, got %d", want, len(challenges))
	}
}

func TestEmitRss(t *testing.T) {
	var cs []challenge
	b := `[
        {
                "name": "EcoTox TARGET Challenge",
                "agency": "Environmental Protection Agency",
                "summary": "Develop high quality, low-cost tools that assess global gene expression in common aquatic toxicity test organisms",
                "deadline": "2021-06-15T23:59:00-04:00",
                "details_url": "https://www.challenge.gov/challenge/ecotox-challenge/"
        },
        {
                "name": "Break the Ice Phase 1",
                "agency": "NASA",
                "summary": "NASA’s Break the Ice Lunar Challenge seeks to incentivize innovative approaches for excavating icy regolith and delivering water in extreme...",
                "deadline": "2021-06-18T23:59:00-04:00",
                "details_url": "https://www.challenge.gov/challenge/break-the-ice-phase1/"
        },
        {
                "name": "Streamflow Forecast Rodeo",
                "agency": "Department of the Interior - Bureau of Reclamation",
                "summary": "Improving short-term streamflow forecast skill.",
                "deadline": "2021-09-30T21:00:00-04:00",
                "details_url": "https://www.topcoder.com/community/streamflow"
        }
    ]`
	if err := json.Unmarshal([]byte(b), &cs); err != nil {
		t.Fatalf("unmarshalling JSON: %v", err)
	}
	if err := challenges(cs).emitRssFeed(os.Stdout); err != nil {
		t.Fatalf("emitting RSS: %v", err)
	}
}
