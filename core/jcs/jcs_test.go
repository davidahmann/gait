package jcs

import "testing"

func TestCanonicalizeJSON(t *testing.T) {
	in := []byte(`{ "b":2, "a":1 }`)
	want := `{"a":1,"b":2}`
	out, err := CanonicalizeJSON(in)
	if err != nil {
		t.Fatalf("canonicalize error: %v", err)
	}
	if string(out) != want {
		t.Fatalf("unexpected canonical form: %s", string(out))
	}
}

func TestDigestJCSStable(t *testing.T) {
	a := []byte(`{"a":1,"b":2}`)
	b := []byte(`{ "b":2, "a":1 }`)

	da, err := DigestJCS(a)
	if err != nil {
		t.Fatalf("digest error: %v", err)
	}
	db, err := DigestJCS(b)
	if err != nil {
		t.Fatalf("digest error: %v", err)
	}
	if da != db {
		t.Fatalf("expected same digest for equivalent JSON")
	}
}

func TestCanonicalizeJSONInvalid(t *testing.T) {
	_, err := CanonicalizeJSON([]byte(`{`))
	if err == nil {
		t.Fatalf("expected error for invalid JSON")
	}
}

func TestDigestJCSInvalid(t *testing.T) {
	_, err := DigestJCS([]byte(`{`))
	if err == nil {
		t.Fatalf("expected error for invalid JSON digest")
	}
}
