package main

import (
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
