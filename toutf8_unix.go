// +build !windows

package mailaddress

import (
	"bytes"
	"io"
	"io/ioutil"

	iconv "gopkg.in/iconv.v1"
)

// Convert the bytes from the input reader to UTF-8
func toUTF8(charset string, input io.Reader) (_ io.Reader, returnErr error) {
	conv, err := iconv.Open("utf-8", charset)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := conv.Close(); err != nil {
			returnErr = err
		}
	}()

	r := iconv.NewReader(conv, input, 0)
	b, err := ioutil.ReadAll(r)
	if err != nil {
		// errno 84 from syscall. Unfortunately we can't check the errno :-/
		if err.Error() == "invalid or incomplete multibyte or wide character" {
			return nil, ErrInvalidEncoding
		}
		return nil, err
	}

	return bytes.NewReader(b), returnErr
}
