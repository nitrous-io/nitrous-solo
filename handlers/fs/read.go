package fs

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	. "github.com/nitrous-io/nitrous-solo/handlers"
)

var FileSizeLimit int64 = 50 * 1024 * 1024

func Read(name string, data []byte, reply func([]byte)) {
	if len(data) == 0 {
		reply([]byte(""))
		return
	}

	var j struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(data, &j); err != nil {
		ReplyJSON(H{"error": "invalid_json"}, reply)
		return
	}

	fi, err := os.Stat(j.Path)
	if err != nil {
		if os.IsNotExist(err) {
			ReplyJSON(H{"error": "invalid_path"}, reply)
			return
		}

		if strings.Contains(err.Error(), "too many levels of symbolic links") {
			ReplyJSON(H{"error": "symlink_too_deep"}, reply)
			return
		}

		ReplyUnexpectedError(err, reply)
		return
	}

	if fi.IsDir() {
		ReplyJSON(H{"error": "path_not_file"}, reply)
		return
	}

	mode := fi.Mode()
	if mode&os.ModeAppend > 0 || mode&os.ModeExclusive > 0 || mode&os.ModeDevice > 0 || mode&os.ModeNamedPipe > 0 || mode&os.ModeSocket > 0 || mode&os.ModeCharDevice > 0 {
		ReplyJSON(H{"error": "file_unreadable"}, reply)
		return
	}

	if fi.Size() > FileSizeLimit {
		ReplyJSON(H{"error": "file_too_big"}, reply)
		return
	}

	var fileType string
	filename := strings.ToLower(filepath.Base(j.Path))
	fileext := filepath.Ext(filename)
	// remove dot from extension
	if len(fileext) > 0 {
		fileext = fileext[1:]
	}

	if _, ok := TextFileExtensions[fileext]; ok {
		fileType = "text"
	} else if _, ok := ImageFileExtensions[fileext]; ok {
		fileType = "image"
	} else if _, ok := RecognizedTextFileNames[filename]; ok {
		fileType = "text"
	} else {
		// unable to determine file type, try to guess from content
		f, err := os.Open(j.Path)
		if err != nil {
			ReplyUnexpectedError(err, reply)
			return
		}
		defer f.Close()

		var buf []byte = make([]byte, 512)
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			ReplyUnexpectedError(err, reply)
			return
		}

		if n == 0 {
			fileType = "text"
		} else {
			mime := http.DetectContentType(buf)
			if strings.HasPrefix(mime, "text/") || mime == "application/postscript" {
				fileType = "text"
			} else if strings.HasPrefix(mime, "image/") {
				fileType = "image"
			} else {
				// check whether first 512 bytes are all comprised of printable chars + a few other chars
				allAscii := true
				for _, b := range buf[:n] {
					if !((b >= 32 && b < 127) || b == 10 /* LF */ || b == 13 /* CR */ || b == 9 /* TAB */) {
						allAscii = false
						break
					}
				}

				if !allAscii {
					ReplyJSON(H{"error": "unrecognized_file_type"}, reply)
					return
				}

				fileType = "text"
			}
		}
	}

	filedata, err := ioutil.ReadFile(j.Path)
	if err != nil {
		ReplyUnexpectedError(err, reply)
		return
	}

	var content string
	if fileType == "image" {
		// encode in base64
		content = base64.StdEncoding.EncodeToString(filedata)
	} else {
		content = string(filedata)
	}

	ReplyJSON(H{"type": fileType, "content": content}, reply)
}
