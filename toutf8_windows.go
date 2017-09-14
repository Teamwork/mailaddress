// +build windows

package mailaddress

import (
	"io"
)

// Iconv does not work well on Windows, so just provide a stub for now.
func toUTF8(charset string, input io.Reader) (io.Reader, error) {
	return input, nil
}
