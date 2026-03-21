package params

import (
	"os"
	"testing"
)

func TestCoerce_BasicParam(t *testing.T) {
	os.Setenv("FRAGLET_PARAM_CITY", "london")
	defer os.Unsetenv("FRAGLET_PARAM_CITY")

	got, err := Coerce()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Name != "CITY" || got[0].Value != "london" {
		t.Fatalf("got %+v, want {CITY london}", got[0])
	}
	// Transport var should be unset
	if _, exists := os.LookupEnv("FRAGLET_PARAM_CITY"); exists {
		t.Fatal("FRAGLET_PARAM_CITY should be unset")
	}
}

func TestCoerce_ValuePassedThrough(t *testing.T) {
	// Entrypoint is dumb — values pass through as-is, no decoding
	os.Setenv("FRAGLET_PARAM_MSG", "already decoded value")
	defer os.Unsetenv("FRAGLET_PARAM_MSG")

	got, err := Coerce()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Value != "already decoded value" {
		t.Fatalf("got %+v, want value passed through as-is", got)
	}
}

func TestCoerce_NoShadow(t *testing.T) {
	os.Setenv("CITY", "existing")
	os.Setenv("FRAGLET_PARAM_CITY", "new")
	defer os.Unsetenv("CITY")
	defer os.Unsetenv("FRAGLET_PARAM_CITY")

	got, err := Coerce()
	if err != nil {
		t.Fatal(err)
	}
	// Shadowed param should not appear in results
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0 (shadowed)", len(got))
	}
	// Existing value untouched
	if v := os.Getenv("CITY"); v != "existing" {
		t.Fatalf("CITY = %q, want existing", v)
	}
	// Transport var still cleaned up
	if _, exists := os.LookupEnv("FRAGLET_PARAM_CITY"); exists {
		t.Fatal("FRAGLET_PARAM_CITY should be unset even when shadowed")
	}
}

func TestCoerce_TransportCleaned(t *testing.T) {
	os.Setenv("FRAGLET_PARAM_FOO", "bar")
	defer os.Unsetenv("FOO")
	defer os.Unsetenv("FRAGLET_PARAM_FOO")

	_, err := Coerce()
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := os.LookupEnv("FRAGLET_PARAM_FOO"); exists {
		t.Fatal("transport var should be unset")
	}
}

func TestCoerce_CasePreserved(t *testing.T) {
	os.Setenv("FRAGLET_PARAM_HURL_VARIABLE_host", "localhost")
	defer os.Unsetenv("FRAGLET_PARAM_HURL_VARIABLE_host")

	got, err := Coerce()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "HURL_VARIABLE_host" {
		t.Fatalf("got %+v, want name=HURL_VARIABLE_host", got)
	}
}

func TestCoerce_NoParams(t *testing.T) {
	got, err := Coerce()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}

func TestCoerce_EmptyBareName(t *testing.T) {
	// FRAGLET_PARAM_ with nothing after it — should be skipped
	os.Setenv("FRAGLET_PARAM_", "val")
	defer os.Unsetenv("FRAGLET_PARAM_")

	got, err := Coerce()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0 (empty bare name skipped)", len(got))
	}
}

func TestCoerce_MultipleParams_Sorted(t *testing.T) {
	os.Setenv("FRAGLET_PARAM_ZEBRA", "z")
	os.Setenv("FRAGLET_PARAM_ALPHA", "a")
	defer os.Unsetenv("FRAGLET_PARAM_ZEBRA")
	defer os.Unsetenv("FRAGLET_PARAM_ALPHA")

	got, err := Coerce()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Name != "ALPHA" || got[1].Name != "ZEBRA" {
		t.Fatalf("got %+v, want sorted ALPHA then ZEBRA", got)
	}
}
