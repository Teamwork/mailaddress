The mailaddress package parses email addresses.

It's an alternative to `net/mail`; significant differences include:

- Better errors.
- When parsing a list it will continue to the next address on an error; this is
  especially useful when providing feedback to users.
- Some useful utility functions.

Basic example:

	addr, err := mailaddress.Parse(`Martin <single_address@example.com>`)

	addrs, haveErr := mailaddress.ParseLint(`many@example.com, addresses@example.com`)
	if haveErr {
		fmt.Println(addrs.Errors())
	}

See godoc for more docs.
