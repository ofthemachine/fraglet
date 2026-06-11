package main

import (
	"errors"
	"testing"
)

func TestParseGuideEssenceArgs_orderAgnostic(t *testing.T) {
	want := guideEssenceOpts{VeinName: "ada", Mode: "main", Image: ""}

	o, err := parseGuideEssenceArgs([]string{"ada", "--mode", "main"})
	if err != nil {
		t.Fatal(err)
	}
	if o != want {
		t.Fatalf("got %+v want %+v", o, want)
	}

	o, err = parseGuideEssenceArgs([]string{"--mode", "main", "ada"})
	if err != nil {
		t.Fatal(err)
	}
	if o != want {
		t.Fatalf("got %+v want %+v", o, want)
	}
}

func TestParseGuideEssenceArgs_imageOnly(t *testing.T) {
	o, err := parseGuideEssenceArgs([]string{"-i", "img:latest", "--mode", "m"})
	if err != nil {
		t.Fatal(err)
	}
	if o.Image != "img:latest" || o.Mode != "m" || o.VeinName != "" {
		t.Fatalf("got %+v", o)
	}
}

func TestParseGuideEssenceArgs_help(t *testing.T) {
	_, err := parseGuideEssenceArgs([]string{"--help"})
	if !errors.Is(err, errGuideEssenceUsage) {
		t.Fatalf("expected usage sentinel, got %v", err)
	}
}

func TestParseGuideEssenceArgs_xorPositionals(t *testing.T) {
	_, err := parseGuideEssenceArgs([]string{"ada", "extra"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseGuideEssenceArgs_unknownFlag(t *testing.T) {
	_, err := parseGuideEssenceArgs([]string{"--foo"})
	if err == nil {
		t.Fatal("expected error")
	}
}
