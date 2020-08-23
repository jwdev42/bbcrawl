package libhttp

import (
	"errors"
	"golang.org/x/net/html/charset"
	"io"
	"net/http"
)

func BodyUTF8(resp *http.Response) (io.Reader, error) {
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		return nil, errors.New("No Content-Type found in HTTP response")
	}
	return charset.NewReader(resp.Body, contentType)
}
