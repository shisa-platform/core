package uuid

import (
	"testing"
)

func TestStringer(t *testing.T) {
	if "6ba7b810-9dad-11d1-80b4-00c04fd430c8" != NamespaceDNS.String() {
		t.Fatalf("Unexpected dns string: %s", NamespaceDNS.String())
	}

	if "6ba7b811-9dad-11d1-80b4-00c04fd430c8" != NamespaceURL.String() {
		t.Fatalf("Unexpected url string: %s", NamespaceURL.String())
	}

	if "e06bac7c-4f4b-5c54-a632-d48a16695027" != ShisaNS.String() {
		t.Fatalf("Unexpected ns string: %s", ShisaNS.String())
	}
}

func TestNewDNS(t *testing.T) {
	// Namespace DNS + "python.org" = "886313e1-3b8a-5372-9b90-0c9aee199e5d"
	expected := "886313e1-3b8a-5372-9b90-0c9aee199e5d"
	u := New(NamespaceDNS, "python.org")
	if expected != u.String() {
		t.Fatalf("received unexpected uuid: %s", u.String())
	}
}

func TestNewURL(t *testing.T) {
	// Namespace URL + "https://percolate.com" = "d24eb514-1c92-5ac6-8d10-61569d14c180"
	expected := "d24eb514-1c92-5ac6-8d10-61569d14c180"
	u := New(NamespaceURL, "https://percolate.com")
	if expected != u.String() {
		t.Fatalf("received unexpected uuid: %s", u.String())
	}
}

func TestParseDNS(t *testing.T) {
	expected := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	u, err := Parse(expected)
	if err != nil {
		t.Fatalf("failed to parse uuid: %s", err)
	}
	if expected != u.String() {
		t.Fatalf("received unexpected uuid: %s", u.String())
	}
}

func TestParseURL(t *testing.T) {
	expected := "6ba7b811-9dad-11d1-80b4-00c04fd430c8"
	u, err := Parse(expected)
	if err != nil {
		t.Fatalf("failed to parse uuid: %s", err)
	}
	if expected != u.String() {
		t.Fatalf("received unexpected uuid: %s", u.String())
	}
}

func TestParseUppercaseURL(t *testing.T) {
	u, err := Parse("6BA7B811-9DAD-11D1-80B4-00C04FD430C8")
	if err != nil {
		t.Fatalf("failed to parse uuid: %s", err)
	}
	if "6ba7b811-9dad-11d1-80b4-00c04fd430c8" != u.String() {
		t.Fatalf("received unexpected uuid: %s", u.String())
	}
}

func TestParseMixedCaseURL(t *testing.T) {
	u, err := Parse("6ba7b811-9DAD-11D1-80b4-00C04FD430C8")
	if err != nil {
		t.Fatalf("failed to parse uuid: %s", err)
	}
	if "6ba7b811-9dad-11d1-80b4-00c04fd430c8" != u.String() {
		t.Fatalf("received unexpected uuid: %s", u.String())
	}
}

func TestShortValue(t *testing.T) {
	_, err := Parse("123")
	if err == nil {
		t.Fatalf("didn't receive expected error")
	}
}

func TestNonHexValueUpper(t *testing.T) {
	_, err := Parse("6ba7b811-9dad-zyxw-80b4-00c04fd430c8")
	if err == nil {
		t.Fatalf("didn't receive expected error")
	}
}

func TestNonHexValueLower(t *testing.T) {
	_, err := Parse("6ba7b811-9dad-axyz-80b4-00c04fd430c8")
	if err == nil {
		t.Fatalf("didn't receive expected error")
	}
}

func TestExtraHyphen(t *testing.T) {
	_, err := Parse("6ba7b811-9d-ad-11d1-80b4-00c04fd430c")
	if err == nil {
		t.Fatalf("didn't receive expected error")
	}
}

func TestMisplacedHyphen(t *testing.T) {
	_, err := Parse("6ba7-b8119dad-11d1-80b4-00c04fd430c8")
	if err == nil {
		t.Fatalf("didn't receive expected error")
	}
}

func TestMissingHyphen(t *testing.T) {
	_, err := Parse("6ba7b811-9dad11d1-80b4-00c04fd430c8x")
	if err == nil {
		t.Fatalf("didn't receive expected error")
	}
}
