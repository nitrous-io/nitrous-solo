package fuzzyfind

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

// FzfFinder is an implementation of Finder that relies on shelling out to fzf
// (https://github.com/junegunn/fzf) to find files.
type FzfFinder struct {
	BinaryPath string
}

func (f *FzfFinder) Find(dir, query string) ([]File, error) {
	dir = filepath.Clean(dir)
	fi, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, ErrFindDirectoryRequired
	}

	// Connect Stderr of both commands to a buffer.
	var stderr bytes.Buffer

	findCmd := exec.Command("find", dir,
		"-type", "f", // Restrict to regular files.
		"-printf", `%P\n`, // Print paths relative to `dir`.
	)
	findCmd.Stderr = &stderr

	pQuery := prepareForFzf(query)
	fzfCmd := exec.Command(f.BinaryPath, "--filter", pQuery)

	fzfCmd.Stdin, err = findCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	var stdout bytes.Buffer
	fzfCmd.Stdout = &stdout
	fzfCmd.Stderr = &stderr

	if err := fzfCmd.Start(); err != nil {
		return nil, err
	}
	if err := findCmd.Run(); err != nil {
		return nil, fmt.Errorf("error running `find`, err: %v, stderr: %s",
			err, stderr.Bytes())
	}
	if err := fzfCmd.Wait(); err != nil {
		return nil, fmt.Errorf("error running `fzf`, err: %v, stderr: %s",
			err, stderr.Bytes())
	}

	files := []File{}
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		files = append(files, File{
			Path: scanner.Text(),
			// Offsets: fzf does not seem to support displaying matching offset in
			// --filter mode, so we don't set this field for now.
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

// prepareForFzf prepares a query for use with the fzf binary.
func prepareForFzf(query string) string {
	// fzf treats words delimited by spaces as separate search terms when run in
	// extended mode (--extended). We remove those spaces since we only ever want
	// a single search term.
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, query)
}
