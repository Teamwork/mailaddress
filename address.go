package mailaddress

import (
	"fmt"
	"mime"
	"strings"
)

// Address is a single mail address.
type Address struct {
	Name    string `db:"name" json:"name"`
	Address string `db:"email" json:"address"`
	Raw     string `db:"-" json:"-"`
	Error   error  `db:"-" json:"-"`
}

// String formats an address. It is *not* RFC 2047 encoded!
func (a Address) String() string {
	if a.Name == "" {
		return a.Address
	}
	return fmt.Sprintf(`"%s" <%s>`, strings.Replace(a.Name, `"`, `\"`, -1), a.Address)
}

// NameEncoded returns the name ready to be put in an email header. Special
// characters will be appropriately escaped and RFC 2047 encoding will be
// applied.
func (a Address) NameEncoded() string {
	if a.Name == "" {
		return ""
	}

	name := a.Name
	if strings.ContainsAny(name, `",;@<>()`) {
		name = fmt.Sprintf(`"%s"`, strings.Replace(name, `"`, `\\"`, -1))
	}

	return mime.QEncoding.Encode("utf-8", name)
}

// AddressEncoded returns the address ready to be put in an email header.
// Special characters will be appropriately escaped and RFC 2047 encoding will
// be applied.
func (a Address) AddressEncoded() string {
	if a.Address == "" {
		return ""
	}

	return mime.QEncoding.Encode("utf-8", a.Address)
}

// StringEncoded makes a string that *is* RFC 2047 encoded
//
// TODO: This won't work with IDN. This is okay since most email clients don't
// work with IDN. Last I checked this included Gmail, FastMail, Thunderbird,
// etc. The only client that works 100% correct AFAIK is mutt.
func (a Address) StringEncoded() string {
	if a.Name == "" {
		return a.Address
	}

	return fmt.Sprintf("%v <%v>", a.NameEncoded(), a.AddressEncoded())
}

// ToList puts this Address in an List.
func (a Address) ToList() (l List) {
	l = append(l, a)
	return l
}

// Local gets the local part of an address (i.e. everything before the first @).
//
// TODO: the local part can contain a quoted/escaped @, but practically no email
// system deals with that, so it's not a huge deal at the moment.
func (a Address) Local() string {
	s := strings.Split(a.Address, "@")
	return s[0]
}

// Domain gets the domain part of an address (i.e. everything after the first
// @).
//
// TODO: Same as Local().
func (a Address) Domain() string {
	s := strings.Split(a.Address, "@")
	if len(s) < 2 {
		return s[0]
	}
	return strings.Join(s[1:], "")
}

// WithoutTag gets the address with the tag part removed (if any). The tag part
// is everything in the local part after the first +.
func (a Address) WithoutTag() string {
	if !a.Valid() {
		return ""
	}
	plus := strings.Index(a.Address, "+")
	at := strings.Index(a.Address, "@")

	if plus != -1 && at != -1 {
		return a.Address[:plus] + a.Address[at:]
	}
	return a.Address
}

// Valid reports if this email looks valid. This includes some small extra
// checks for sanity. For example "martin@arp242 is a "valid" email address in
// the RFC sense, but not in the "something we can send emails to"-sense.
//
// TODO: Perhaps consider renaming to CanSend() or Sendable() or Deliverable()?
//
// It is also useful if the address wasn't created with ParseList() but directly
// (e.g. addr := Address{...}).
func (a *Address) Valid() bool {
	if a.Address == "" || !reValidEmail.MatchString(a.Address) {
		a.Error = ErrNoEmail
		return false
	}

	return a.Error == nil
}
