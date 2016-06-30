package fuzzyfind

import "errors"

var (
	ErrFindDirectoryRequired = errors.New("a directory is required")
)

type Finder interface {
	Find(dir, query string) (results []File, err error)
}

type File struct {
	// Path is the relative path to the file.
	Path    string
	Offsets []Offset
}

// Offset holds two integers denoting the offsets of a matched substring.
type Offset [2]int
