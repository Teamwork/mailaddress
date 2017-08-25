[![Build Status](https://travis-ci.org/Teamwork/mailaddress.svg?branch=master)](https://travis-ci.org/Teamwork/mailaddress)
[![codecov](https://codecov.io/gh/Teamwork/mailaddress/branch/master/graph/badge.svg?token=n0k8YjbQOL)](https://codecov.io/gh/Teamwork/mailaddress)
[![GoDoc](https://godoc.org/github.com/Teamwork/mailaddress?status.svg)](https://godoc.org/github.com/Teamwork/mailaddress)
[![Go Report Card](https://goreportcard.com/badge/github.com/Teamwork/mailaddress)](https://goreportcard.com/report/github.com/Teamwork/mailaddress)

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
