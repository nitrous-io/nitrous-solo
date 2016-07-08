package testhelper

import (
	"encoding/json"
	"reflect"
	"runtime"
	"time"

	. "github.com/onsi/gomega"
)

type H map[string]interface{}
type A []interface{}
type Any interface{}

func EncodeJSON(h interface{}) []byte {
	j, err := json.Marshal(h)
	Expect(err).To(BeNil())
	return j
}

func Tick() {
	runtime.Gosched()
	time.Sleep(10 * time.Millisecond)
}

func ExpectSameFunc(a interface{}, b interface{}) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)
	Expect(av.Kind().String()).To(Equal("func"), "a is not a func")
	Expect(bv.Kind().String()).To(Equal("func"), "b is not a func")
	Expect(av.Pointer()).To(Equal(bv.Pointer()), "a and b are not the same func")
}
