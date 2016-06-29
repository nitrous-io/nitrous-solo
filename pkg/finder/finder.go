package finder

import (
	"bytes"
	"encoding/base64"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/Sirupsen/logrus"
)

// Line represents a matched line in search response
type Line struct {
	Row                int    `json:"row"`
	Content            string `json:"content"`
	HighlightedContent string `json:"highlighted_content"`
	Count              int    `json:"count"` // Number of matches on this line
}

type SearchResult struct {
	File    string  `json:"file"`
	Matches []*Line `json:"matches"`
}

const (
	AgBinary = "ag"
)

var (
	HighlightColorCode  = "30;43"
	EmptyColorStartCode = []byte("\x1b[m")
	ColorStartCode      = []byte("\x1b[" + HighlightColorCode + "m")
	ColorEndCode        = []byte("\x1b[0m\x1b[K")
)

var (
	ErrRootPathNotFound     = errors.New("invalid_path")
	ErrRootPathNotDirectory = errors.New("path_not_directory")

	DefaultOptions = []string{"--color", "--nogroup", "--follow", "--color-match=" + HighlightColorCode, "--color-line-number=", "--color-path="}

	ContentSearchLimit = 500
)

func FindContent(rootPath, term, filePattern string, ignoreCase, useRegexp bool) ([]*SearchResult, error) {
	fileinfo, err := os.Stat(rootPath)
	if os.IsNotExist(err) {
		return nil, ErrRootPathNotFound
	}

	if !fileinfo.IsDir() {
		return nil, ErrRootPathNotDirectory
	}

	var args []string
	if filePattern != "" {
		args = append(args, "-G="+filePattern)
	}

	if ignoreCase {
		args = append(args, "-i")
	}

	if !useRegexp {
		args = append(args, "-w")
	}

	args = append(args, DefaultOptions...)
	args = append(args, term, rootPath)
	cmd := exec.Command(AgBinary, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if len(out) != 0 {
			logrus.Errorf("failure on find_content: fails run ag due to %v", string(out))
		}

		return nil, nil
	}

	fileMap := make(map[string]*SearchResult)
	searchResults := make([]*SearchResult, 0)
	matches := bytes.Split(out, []byte("\n"))
	searchResultCount := 0
	for _, match := range matches {
		if len(match) == 0 {
			continue
		}

		if bytes.HasPrefix(match, []byte("ERR:")) {
			continue
		}

		tokens := bytes.SplitN(match, []byte(":"), 3) // fileName:LineNumber:Content
		fileName, err := filepath.Rel(rootPath, string(trimColorCode(tokens[0])))
		if err != nil {
			logrus.Errorf("failure on find_content: fails to get relative path for %s due to %v", tokens[0], err)
			continue
		}

		lineNumber, err := strconv.Atoi(string(trimColorCode(tokens[1])))
		if err != nil {
			logrus.Errorf("failure on find_content: fails to convert %s to int due to %v", tokens[1], err)
			continue
		}

		content := tokens[2]
		searchResult := fileMap[fileName]
		if searchResult == nil {
			searchResult = &SearchResult{
				File:    fileName,
				Matches: make([]*Line, 0),
			}
			fileMap[fileName] = searchResult
		}

		searchResult.Matches = append(searchResult.Matches, &Line{
			Row:                lineNumber,
			Content:            base64.StdEncoding.EncodeToString(trimColorCode(content)),
			HighlightedContent: base64.StdEncoding.EncodeToString(content),
			Count:              bytes.Count(content, ColorStartCode),
		})

		searchResults = append(searchResults, searchResult)
		searchResultCount++

		if searchResultCount >= ContentSearchLimit {
			break
		}
	}

	return searchResults, nil
}

func trimColorCode(str []byte) []byte {
	str = bytes.Replace(str, ColorStartCode, []byte{}, -1)
	str = bytes.Replace(str, ColorEndCode, []byte{}, -1)
	return bytes.Replace(str, EmptyColorStartCode, []byte{}, -1)
}
