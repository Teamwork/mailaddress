package mailaddress

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
)

// List of zero or more addresses.
type List []Address

// String formats all addresses. It is *not* RFC 2047 encoded!
func (l List) String() string {
	var out []string
	for _, a := range l {
		out = append(out, a.String())
	}
	return strings.Join(out, ", ")
}

// UnmarshalJSON allows accepting several different formats for a list of mail
// addresses, which are (in order):
//
// 1. Standard List struct JSON string output.
// 2. Slice of strings containing emails.
// 3. Comma-separated string of emails, as accepted by ParseList().
func (l *List) UnmarshalJSON(data []byte) error {
	type Alias List
	var alias Alias
	err := json.Unmarshal(data, &alias)
	if err == nil {
		*l = List(alias)
		return nil
	}

	var slice []string
	err = json.Unmarshal(data, &slice)
	if err != nil {
		var str string

		err = json.Unmarshal(data, &str)
		if err != nil {
			return err
		}

		*l, _ = ParseList(str)
		return nil
	}

	*l = FromSlice(slice)
	return nil
}

// StringEncoded makes a string that *is* RFC 2047 encoded.
func (l List) StringEncoded() string {
	var out []string
	for _, a := range l {
		out = append(out, a.StringEncoded())
	}
	return strings.Join(out, ", ")
}

// Append adds a new Address to the list.
func (l *List) Append(name, address string) {
	*l = append(*l, New(name, address))
}

// Slice gets all valid addresses in a []string slice. The names are lost and
// invalid addresses are skipped.
func (l List) Slice() []string {
	mails := []string{}
	for _, m := range l {
		if m.Valid() {
			mails = append(mails, m.Address)
		}
	}
	return mails
}

// Errors gets a list of all errors. The returned error is a multierror
// (github.com/hashicorp/go-multierror).
func (l List) Errors() (errs error) {
	for _, a := range l {
		if !a.Valid() {
			errs = multierror.Append(errs, a.err)
		}
	}
	return errs
}

// ValidAddresses returns a copy of the list which only includes valid email
// addresses.
func (l List) ValidAddresses() (valid List) {
	for _, addr := range l {
		if addr.Valid() {
			valid = append(valid, addr)
		}
	}
	return valid
}

// ContainsAddress reports if the list contains the specified email address.
func (l List) ContainsAddress(address string) bool {
	for _, addr := range l {
		if strings.EqualFold(addr.Address, address) {
			return true
		}
	}
	return false
}

// ContainsDomain reports if the list contains one or more addresses with the
// given domain.
func (l List) ContainsDomain(domain string) bool {
	for _, addr := range l {
		if strings.EqualFold(addr.Domain(), domain) {
			return true
		}
	}
	return false
}

// Sort keys
const (
	ByAddress = iota
	ByName
)

// Sort the list in-place using one of the By* keys.
func (l List) Sort(key int8) {
	var sortFunc func(int, int) bool
	switch key {
	case ByAddress:
		sortFunc = func(i, j int) bool { return l[i].Address < l[j].Address }
	case ByName:
		sortFunc = func(i, j int) bool { return l[i].Name < l[j].Name }
	default:
		panic(fmt.Sprintf("invalid sort key: %v", key))
	}

	sort.Slice(l, sortFunc)
}
