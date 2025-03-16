package net

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
)

func DrainHttpBody(b io.ReadCloser) (body []byte, r io.ReadCloser, err error) {
	if b == nil || b == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return nil, http.NoBody, nil
	}

	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err = b.Close(); err != nil {
		return nil, b, err
	}
	var bs = buf.Bytes()

	if len(bs) > 1 && bs[0] == 31 && bs[1] == 139 {
		greader, _ := gzip.NewReader(bytes.NewReader(buf.Bytes()))
		body, err := io.ReadAll(greader)
		if err != nil {
			return nil, ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
		}
		return body, io.NopCloser(bytes.NewReader(body)), nil
	}

	data := buf.Bytes()
	return data, io.NopCloser(bytes.NewReader(data)), nil
}
