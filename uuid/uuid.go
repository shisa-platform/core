package uuid

import (
	"crypto/sha1"
	"fmt"
)

type UUID [16]byte

var (
	NamespaceDNS = mustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	NamespaceURL = mustParse("6ba7b811-9dad-11d1-80b4-00c04fd430c8")

	// URL Namespace and "https://github.com/shisa-platform/core"
	ShisaNS = mustParse("e06bac7c-4f4b-5c54-a632-d48a16695027")
)

func New(ns *UUID, name string) (u *UUID) {
	hash := sha1.New()
	hash.Write(ns[:])
	hash.Write([]byte(name))

	u = new(UUID)
	copy(u[:], hash.Sum(nil)[:16])
	u[6] = (u[6] & 0x0f) | (5 << 4) // version is 5
	u[8] = (u[8] & 0xbf) | 0x80     // variant is RFC 4122

	return
}

func Parse(s string) (u *UUID, err error) {
	if 36 != len(s) {
		err = fmt.Errorf("Invalid UUID format: length is %v", len(s))
		return
	}

	u = new(UUID)
	for i, j := 0, 0; i < len(s); {
		switch i {
		case 8, 13, 18, 23:
			if '-' != s[i] {
				err = fmt.Errorf("Invalid UUID format: missing hyphen at %v", i)
				return
			} else {
				i++
				continue
			}
		}
		if '-' == s[i] {
			err = fmt.Errorf("Invalid UUID format: unexpected hyphen at %v", i)
			return
		}
		upper, ok := fromHexChar(s[i])
		if !ok {
			err = fmt.Errorf("Invalid UUID format: bad character at %v", i)
			return
		}
		lower, ok := fromHexChar(s[i+1])
		if !ok {
			err = fmt.Errorf("Invalid UUID format: bad character at %v", i+1)
			return
		}
		u[j] = (upper << 4) | lower
		j++
		i += 2
	}
	return
}

func (u *UUID) String() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
}

func mustParse(s string) *UUID {
	u, err := Parse(s)
	if err != nil {
		panic(err)
	}

	return u
}

func fromHexChar(c byte) (byte, bool) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	}

	return 0, false
}
