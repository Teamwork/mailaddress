package mailaddress

import (
	"errors"
	"mime"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/teamwork/toutf8"
)

var (
	reSanitizeWhitespace = regexp.MustCompile(`\s+`)
	reRemoveComment      = regexp.MustCompile(`\s+\(.*?\)$`)
	reFindEmail          = regexp.MustCompile(`[^\s<>]+@[^\s<>]+\.[^\s<>]+`)

	// Note: this is repeated in helpers/form.coffee
	reValidEmail = regexp.MustCompile(`` +
		// Anchor
		`^` +

		// Local part; allow almost everything
		`[^\s<>@;]+` +

		// @
		`@` +

		// Domain part
		//
		// See RFC 1034, section 3.1, RFC 1035, secion 2.3.1
		//
		// - Only allow letters, numbers
		// - Max size of a single label is 63 characters (RFC specifies bytes, but that's
		//   not so easy to check AFAIK).
		// - Need at least two labels
		`[\p{L}\d-]{1,63}` + // Label
		`(\.[\p{L}\d-]{1,63})+` + // More labels

		// Anchor
		`$`)

	// ErrInvalidEncoding is used when we can't decode an address because the
	// encoding is invalid (>95% of the time this means it's spam).
	ErrInvalidEncoding = errors.New("invalid or incomplete multibyte or wide character")

	// ErrNoEmail is used when we can't find an email address at all.
	ErrNoEmail = errors.New("unable to find an email address")

	// ErrTooManyEmails is used when too many email addresses were found.
	ErrTooManyEmails = errors.New("only one address expected")

	// ErrInvalidCharacter is used when unexpected data is encountered.
	ErrInvalidCharacter = errors.New("invalid character")
)

func parse(str string) (list List, haveError bool) {
	// Sanitize whitespace
	str = reSanitizeWhitespace.ReplaceAllString(str, " ")

	list = List{}
	addr := Address{}
	inAddress := false
	inQuote := false
	for i, code := range str {
		chr := string(code)

		switch {
		case code == utf8.RuneError:
			addr.Raw += chr
			addr.err = ErrInvalidEncoding
			haveError = true

		// Don't allow unprintable characters.
		case code < 0x09 || (code >= 0x0b && code < 0x20):
			addr.Raw += chr
			addr.err = ErrInvalidCharacter
			haveError = true

		case chr == `\`:
			// Ignore
			addr.Raw += `\`

		// Quote
		// TODO: support quoting the local part too.
		case chr == `"`:
			addr.Raw += chr

			// Escaped
			if inQuote && i > 0 && str[i-1] == '\\' {
				if inAddress {
					addr.Address += chr
				} else {
					addr.Name += chr
				}
				continue
			}

			inQuote = !inQuote

		// Start <angl-addr>
		case !inQuote && chr == "<":
			addr.Raw += "<"
			inAddress = true

		// End <angl-addr>
		case !inQuote && chr == ">":
			addr.Raw += ">"
			// we've observed name including `<>`
			if i < len(str)-1 && !addr.Valid() {
				addr.Name += " " + addr.Address
				addr.Address = ""
				addr.err = nil // valid is not idempoent
			}

			inAddress = false

		// Next <address>
		case !inQuote && (chr == "," || chr == ";" || inAddress && unicode.IsSpace(code)): // ';' introduced by outlook
			haveError = end(&addr) || haveError
			if addr.Name != "" || addr.Address != "" || addr.err != nil {
				list = append(list, addr)
			}
			addr = Address{}

		// We've seen <angl-addr> but more data :-/
		case !inQuote && !inAddress && addr.Address != "" && !unicode.IsSpace(code):
			// Set error and read over it.
			if addr.err == nil {
				addr.err = ErrInvalidCharacter
				haveError = true
			}

		// Append to address.
		case inAddress:
			addr.Raw += chr
			addr.Address += chr

		// Append to name.
		default:
			addr.Raw += chr
			addr.Name += chr
		}
	}

	haveError = end(&addr) || haveError
	if addr.Name != "" || addr.Address != "" || addr.err != nil {
		list = append(list, addr)
	}

	return list, haveError
}

func end(a *Address) (goterror bool) {
	a.Name = strings.TrimSpace(a.Name)
	a.Raw = strings.TrimSpace(a.Raw)

	// remove single quotes if they are only around the name
	if len(a.Name) > 2 && !strings.Contains(a.Name[1:len(a.Name)-1], "'") {
		a.Name = strings.Trim(a.Name, "'")
	}

	// Remove any RFC 2047 encoding. Any encoded word is a single <atom>
	// (i.e. characters such as comma, <, ", etc. don't get interpreted in
	// their special meaning), so this is why we do this here.
	decoder := mime.WordDecoder{CharsetReader: toutf8.Reader}
	decoded, err := decoder.DecodeHeader(a.Name)
	if err != nil {
		a.err = err
		a.Name = ""
		return true
	}
	a.Name = decoded

	// It was just an <addr-spec> and not a <angle-addr> or <name-addr>.
	if a.Address == "" && a.Name != "" {
		// Remove the "comment" part: "daemon@foo.org (Mailer Daemon)".
		a.Name = reRemoveComment.ReplaceAllString(a.Name, "")

		// Technically "martin" is also a valid address (a local one) but this
		// is not something people are going to send emails from.
		mail := reFindEmail.FindString(a.Name)
		if mail != "" {
			a.Address = mail
			if len(mail) != len(a.Name) {
				a.err = ErrInvalidCharacter
				goterror = true
			}
		} else {
			a.err = ErrNoEmail
			goterror = true
		}

		a.Name = ""
	}

	// Includes some sanity checks; it sets Error.
	if a.Address != "" {
		e := a.Valid()
		goterror = goterror && e
	}

	return goterror
}
