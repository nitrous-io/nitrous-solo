package fs

import (
	"encoding/json"

	"github.com/Sirupsen/logrus"
	. "github.com/nitrous-io/nitrous-solo/handlers"
	"github.com/nitrous-io/nitrous-solo/pkg/finder"
)

type SearchRequest struct {
	ID          string `json:"id"`
	Term        string `json:"term"`
	Path        string `json:"path"`
	IgnoreCase  bool   `json:"ignore_case"`
	UseRegexp   bool   `json:"use_regexp"`
	FilePattern string `json:"file_pattern"`
}

type SearchResponse struct {
	Results []*finder.SearchResult `json:"results"`
}

func FindContent(name string, data []byte, reply func([]byte)) {
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
		logrus.Errorf("Failed to encode %+v due to %v", resp, err)
		reply([]byte(`{"error": "internal_server_error"}`))
		return
	}

	reply(rawResponse)
}
