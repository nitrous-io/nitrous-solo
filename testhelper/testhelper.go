package testhelper

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	. "github.com/onsi/gomega"
)

type H map[string]interface{}
type A []interface{}
type Any interface{}

func Noop() {}

func EncodeJSON(h interface{}) []byte {
	j, err := json.Marshal(h)
	Expect(err).To(BeNil())
	return j
}

func ExpectSameFunc(a interface{}, b interface{}) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)
	Expect(av.Kind().String()).To(Equal("func"), "a is not a func")
	Expect(bv.Kind().String()).To(Equal("func"), "b is not a func")
	Expect(av.Pointer()).To(Equal(bv.Pointer()), "a and b are not the same func")
}

func ParseJSONBody(body io.Reader) map[string]interface{} {
	var j map[string]interface{}
	err := json.NewDecoder(body).Decode(&j)
	Expect(err).To(BeNil())
	return j
}

func ReadBody(res *http.Response) []byte {
	b := &bytes.Buffer{}
	_, err := b.ReadFrom(res.Body)
	Expect(err).To(BeNil())
	return b.Bytes()
}

// Do it all http request method
func MakeRequest(method, url string, data url.Values, headers http.Header, reqFn func(*http.Request)) (resp *http.Response, err error) {
	var (
		req    *http.Request
		isForm = method == "POST" || method == "PUT" || method == "PATCH"
	)

	if data == nil {
		req, err = http.NewRequest(method, url, nil)
	} else if isForm {
		r := strings.NewReader(data.Encode())
		req, err = http.NewRequest(method, url, r)
	} else {
		req, err = http.NewRequest(method, url+"?"+data.Encode(), nil)
	}

	if err != nil {
		return nil, err
	}

	if isForm {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	if headers != nil {
		for k, v := range headers {
			for _, h := range v {
				req.Header.Add(k, h)
			}
		}
	}

	if reqFn != nil {
		reqFn(req)
	}

	return http.DefaultClient.Do(req)
}
