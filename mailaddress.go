package mailaddress

// ParseList will parse one or more addresses.
func ParseList(str string) (l List, haveError bool) {
	return parse(str)
}

// Parse will parse exactly one address. More than one addresses is an error,
// otherwise it behaves as ParseList().
func Parse(str string) (Address, error) {
	list, _ := ParseList(str)

	if len(list) == 0 {
		return Address{}, ErrNoEmail
	}

	if len(list) > 1 {
		return Address{}, ErrTooManyEmails
	}

	return list[0], list[0].err
}

// New is a shortcut to make a new Address
func New(name, address string) Address {
	a, err := Parse(address)
	if err != nil {
		return Address{Name: name, Address: "", err: err}
	}
	return Address{Name: name, Address: a.Address}
}

// NewList is a shortcut to make a new List
func NewList(name, address string) List {
	return List{New(name, address)}
}

// FromMap creates a List from a "map[name string]email string".
func FromMap(m map[string]string) (l List) {
	for k, v := range m {
		l.Append(k, v)
	}
	return l
}

// FromSlice creates a List from a []string. Only email addresses are set.
func FromSlice(s []string) (l List) {
	for _, v := range s {
		l.Append("", v)
	}
	return l
}
