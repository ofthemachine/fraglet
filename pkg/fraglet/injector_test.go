package fraglet

import (
	"testing"

	"github.com/ofthemachine/fraglet/pkg/inject"
)

func TestInjectString(t *testing.T) {
	config := &inject.Config{
		Match: "FRAGLET",
	}

	result, err := inject.InjectString("Hello FRAGLET World", "Injected", config)
	if err != nil {
		t.Fatalf("InjectString failed: %v", err)
	}

	expected := "Hello Injected World"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestInjectStringWithRange(t *testing.T) {
	config := &inject.Config{
		MatchStart: "BEGIN",
		MatchEnd:   "END",
	}

	template := "Before\nBEGIN\nOld\nEND\nAfter"
	result, err := inject.InjectString(template, "New", config)
	if err != nil {
		t.Fatalf("InjectString failed: %v", err)
	}

	expected := "Before\nNew\nAfter"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestInjectStringNoMatch(t *testing.T) {
	config := &inject.Config{
		MatchStart: "NOTFOUND",
		MatchEnd:   "ALSONOTFOUND",
	}

	_, err := inject.InjectString("Some text", "Code", config)
	if err == nil {
		t.Error("Expected error for missing match, got nil")
	}
}
