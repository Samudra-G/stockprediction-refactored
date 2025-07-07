package pkg

import (
	"bytes"
	"encoding/json"
	"io"
)

func ToJSONReader(data interface{}) io.Reader {
	b, _ := json.Marshal(data)
	return bytes.NewReader(b)
}

func FromJSON(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}