package fs

import (
	"encoding/json"

	. "github.com/nitrous-io/nitrous-solo/handlers"
	"github.com/nitrous-io/nitrous-solo/pkg/finder"
	"github.com/nitrous-io/nitrous-solo/pkg/ws"
)

type SearchRequest struct {
	Term        string `json:"term"`
	Path        string `json:"path"`
	IgnoreCase  bool   `json:"ignore_case"`
	UseRegexp   bool   `json:"use_regexp"`
	FilePattern string `json:"file_pattern"`
}

type SearchResponse struct {
	Results []*finder.SearchResult `json:"results"`
}

func FindInFiles(c *ws.Context) {
	defer c.Done()
	data := c.Data
	reply := c.Reply

	var req SearchRequest
	if err := json.Unmarshal(data, &req); err != nil {
		ReplyJSON(H{"error": "invalid_json"}, reply)
		return
	}

	rootPath := req.Path
	term := req.Term
	ignoreCase := req.IgnoreCase
	useRegexp := req.UseRegexp
	filePattern := req.FilePattern

	if term == "" {
		ReplyJSON(H{"error": "invalid_term"}, reply)
		return
	}

	result, err := finder.FindContent(rootPath, term, filePattern, ignoreCase, useRegexp)
	if err != nil {
		ReplyJSON(H{"error": err.Error()}, reply)
		return
	}

	resp := &SearchResponse{
		Results: result,
	}

	rawResponse, err := json.Marshal(resp)
	if err != nil {
		ReplyUnexpectedError(err, reply)
		return
	}

	reply(rawResponse)
}
