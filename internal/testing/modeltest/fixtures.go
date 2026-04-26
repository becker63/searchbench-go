package modeltest

import (
	"embed"
	"fmt"
)

//go:embed testdata/*
var fixtureFS embed.FS

// Fixture returns a scrubbed test fixture body from this package's testdata
// directory.
func Fixture(name string) ([]byte, error) {
	path := "testdata/" + name
	body, err := fixtureFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read model fixture %q: %w", name, err)
	}
	return body, nil
}

// MustFixture returns a fixture body or panics.
func MustFixture(name string) []byte {
	body, err := Fixture(name)
	if err != nil {
		panic(err)
	}
	return body
}
