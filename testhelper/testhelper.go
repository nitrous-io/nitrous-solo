package testhelper

import (
	"encoding/json"

	. "github.com/onsi/gomega"
)

type H map[string]interface{}

func EncodeJSON(h map[string]interface{}) []byte {
	j, err := json.Marshal(h)
	Expect(err).To(BeNil())
	return j
}
