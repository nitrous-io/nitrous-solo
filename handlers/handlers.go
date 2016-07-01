package handlers

import "encoding/json"

type H map[string]interface{}

func ReplyJSON(j map[string]interface{}, reply func([]byte)) {
	if j == nil {
		return
	}
	m, err := json.Marshal(j)
	if err != nil {
		reply([]byte(`{"error":"internal_server_error"}`))
		return
	}
	reply(m)
}

func ReplyUnexpectedError(err error, reply func([]byte)) {
	j := map[string]string{"error": "unexpected_error"}
	if err != nil {
		j["error_description"] = err.Error()
	}
	m, err := json.Marshal(j)
	if err != nil {
		reply([]byte(`{"error":"internal_server_error"}`))
		return
	}
	reply(m)
}
