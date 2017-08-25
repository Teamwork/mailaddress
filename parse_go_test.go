package mailaddress

import (
	"fmt"
	"testing"
)

// Copy of the relevant net/mail tests.

func TestAddressParsing(t *testing.T) {
	tests := []struct {
		addrsStr string
		expected List
	}{
		// Bare address
		{
			`jdoe@machine.example`,
			List{{
				Address: "jdoe@machine.example",
			}},
		},
		// RFC 5322, Appendix A.1.1
		{
			`John Doe <jdoe@machine.example>`,
			List{{
				Name:    "John Doe",
				Address: "jdoe@machine.example",
			}},
		},
		// RFC 5322, Appendix A.1.2
		{
			`"Joe Q. Public" <john.q.public@example.com>`,
			List{{
				Name:    "Joe Q. Public",
				Address: "john.q.public@example.com",
			}},
		},
		{
			`Mary Smith <mary@x.test>, jdoe@example.org, Who? <one@y.test>`,
			List{
				{
					Name:    "Mary Smith",
					Address: "mary@x.test",
				},
				{
					Address: "jdoe@example.org",
				},
				{
					Name:    "Who?",
					Address: "one@y.test",
				},
			},
		},
		{
			`<boss@nil.test>, "Giant; \"Big\" Box" <sysservices@example.net>`,
			List{
				{
					Address: "boss@nil.test",
				},
				{
					Name:    `Giant; "Big" Box`,
					Address: "sysservices@example.net",
				},
			},
		},
		// RFC 5322, Appendix A.1.3
		// TODO(dsymonds): Group addresses.

		// RFC 2047 "Q"-encoded ISO-8859-1 address.
		{
			`=?iso-8859-1?q?J=F6rg_Doe?= <joerg@example.com>`,
			List{
				{
					Name:    `Jörg Doe`,
					Address: "joerg@example.com",
				},
			},
		},
		// RFC 2047 "Q"-encoded US-ASCII address. Dumb but legal.
		{
			`=?us-ascii?q?J=6Frg_Doe?= <joerg@example.com>`,
			List{
				{
					Name:    `Jorg Doe`,
					Address: "joerg@example.com",
				},
			},
		},
		// RFC 2047 "Q"-encoded UTF-8 address.
		{
			`=?utf-8?q?J=C3=B6rg_Doe?= <joerg@example.com>`,
			List{
				{
					Name:    `Jörg Doe`,
					Address: "joerg@example.com",
				},
			},
		},
		// RFC 2047, Section 8.
		{
			`=?ISO-8859-1?Q?Andr=E9?= Pirard <PIRARD@vm1.ulg.ac.be>`,
			List{
				{
					Name:    `André Pirard`,
					Address: "PIRARD@vm1.ulg.ac.be",
				},
			},
		},
		// Custom example of RFC 2047 "B"-encoded ISO-8859-1 address.
		{
			`=?ISO-8859-1?B?SvZyZw==?= <joerg@example.com>`,
			List{
				{
					Name:    `Jörg`,
					Address: "joerg@example.com",
				},
			},
		},
		// Custom example of RFC 2047 "B"-encoded UTF-8 address.
		{
			`=?UTF-8?B?SsO2cmc=?= <joerg@example.com>`,
			List{
				{
					Name:    `Jörg`,
					Address: "joerg@example.com",
				},
			},
		},
		// Custom example with "." in name. For issue 4938
		{
			`Asem H. <noreply@example.com>`,
			List{
				{
					Name:    `Asem H.`,
					Address: "noreply@example.com",
				},
			},
		},
		// RFC 6532 3.2.3, qtext /= UTF8-non-ascii
		{
			`"Gø Pher" <gopher@example.com>`,
			List{
				{
					Name:    `Gø Pher`,
					Address: "gopher@example.com",
				},
			},
		},
		// RFC 6532 3.2, atext /= UTF8-non-ascii
		{
			`µ <micro@example.com>`,
			List{
				{
					Name:    `µ`,
					Address: "micro@example.com",
				},
			},
		},
		// RFC 6532 3.2.2, local address parts allow UTF-8
		{
			`Micro <µ@example.com>`,
			List{
				{
					Name:    `Micro`,
					Address: "µ@example.com",
				},
			},
		},
		// RFC 6532 3.2.4, domains parts allow UTF-8
		{
			`Micro <micro@µ.example.com>`,
			List{
				{
					Name:    `Micro`,
					Address: "micro@µ.example.com",
				},
			},
		},
		// Issue 14866
		{
			`"" <emptystring@example.com>`,
			List{
				{
					Name:    "",
					Address: "emptystring@example.com",
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			// Single address
			if len(test.expected) == 1 {
				addr, err := Parse(test.addrsStr)
				if err != nil {
					t.Fatalf("error parsing: %v", err)
				}

				if !cmpaddr(addr, test.expected[0]) {
					t.Errorf("\ngot  %+v\nwant %+v", addr, test.expected)
				}
			}

			addrs, haveErr := ParseList(test.addrsStr)
			if haveErr {
				t.Fatalf("error parsing: %v", addrs.Errors())
			}
			if !cmplist(addrs, test.expected) {
				t.Errorf("\ngot  %+v\nwant %+v", addrs, test.expected)
			}
		})
	}
}

func TestAddressString(t *testing.T) {
	tests := []struct {
		addr Address
		exp  string
	}{
		{Address{Address: "bob@example.com"}, "bob@example.com"},
		{Address{Name: "Bob", Address: "bob@example.com"}, `Bob <bob@example.com>`},

		// TODO: unsupported
		// quoted local parts: RFC 5322, 3.4.1. and 3.2.4.
		//{Address{Address: `my@idiot@address@example.com`}, `<"my@idiot@address"@example.com>`},
		// quoted local parts
		//{Address{Address: ` @example.com`}, `<" "@example.com>`},

		// note the ö (o with an umlaut)
		{Address{Name: "Böb", Address: "bob@example.com"}, `=?utf-8?q?B=C3=B6b?= <bob@example.com>`},

		{Address{Name: "Bob Jane", Address: "bob@example.com"}, `Bob Jane <bob@example.com>`},
		{Address{Name: "Böb Jacöb", Address: "bob@example.com"}, `=?utf-8?q?B=C3=B6b_Jac=C3=B6b?= <bob@example.com>`},

		// https://golang.org/issue/12098
		{Address{Name: "Rob", Address: ""}, `Rob <>`},
		{Address{Name: "Rob", Address: "@"}, `Rob <@>`},

		// TODO
		//{Address{Name: "Böb, Jacöb", Address: "bob@example.com"}, `=?utf-8?b?QsO2YiwgSmFjw7Zi?= <bob@example.com>`},
		//{Address{Name: "=??Q?x?=", Address: "hello@world.com"}, `"=??Q?x?=" <hello@world.com>`},
		{Address{Name: "=?hello", Address: "hello@world.com"}, `=?hello <hello@world.com>`},
		{Address{Name: "world?=", Address: "hello@world.com"}, `world?= <hello@world.com>`},

		// should q-encode even for invalid utf-8.
		{
			Address{Name: string([]byte{0xed, 0xa0, 0x80}), Address: "invalid-utf8@example.net"},
			"=?utf-8?q?=ED=A0=80?= <invalid-utf8@example.net>",
		},
	}

	for _, test := range tests {
		t.Run(test.exp, func(t *testing.T) {
			s := test.addr.StringEncoded()
			if s != test.exp {
				t.Fatalf("\ngot  %v\nwant %v", s, test.exp)
			}

			// Check round-trip.
			if test.addr.Address != "" && test.addr.Address != "@" {
				a, err := Parse(test.exp)
				if err != nil {
					t.Fatalf("%v", err)
				}

				if a.Name != test.addr.Name || a.Address != test.addr.Address {
					t.Errorf("\ngot  %#v\nwant %#v", a, test.addr)
				}
			}
		})
	}
}

/*
TODO

// Check if all valid addresses can be parsed, formatted and parsed again
func TestAddressParsingAndFormatting(t *testing.T) {

	// Should pass
	tests := []string{
		`<Bob@example.com>`,
		`<bob.bob@example.com>`,
		`<".bob"@example.com>`,
		`<" "@example.com>`,
		`<some.mail-with-dash@example.com>`,
		`<"dot.and space"@example.com>`,
		`<"very.unusual.@.unusual.com"@example.com>`,
		`<admin@mailserver1>`,
		`<postmaster@localhost>`,
		"<#!$%&'*+-/=?^_`{}|~@example.org>",
		`<"very.(),:;<>[]\".VERY.\"very@\\ \"very\".unusual"@strange.example.com>`, // escaped quotes
		`<"()<>[]:,;@\\\"!#$%&'*+-/=?^_{}| ~.a"@example.org>`,                      // escaped backslashes
		`<"Abc\\@def"@example.com>`,
		`<"Joe\\Blow"@example.com>`,
		`<test1/test2=test3@example.com>`,
		`<def!xyz%abc@example.com>`,
		`<_somename@example.com>`,
		`<joe@uk>`,
		`<~@example.com>`,
		`<"..."@test.com>`,
		`<"john..doe"@example.com>`,
		`<"john.doe."@example.com>`,
		`<".john.doe"@example.com>`,
		`<"."@example.com>`,
		`<".."@example.com>`,
		`<"0:"@0>`,
	}

	for _, test := range tests {
		addr, err := ParseAddress(test)
		if err != nil {
			t.Errorf("Couldn't parse address %s: %s", test, err.Error())
			continue
		}
		str := addr.String()
		addr, err = ParseAddress(str)
		if err != nil {
			t.Errorf("ParseAddr(%q) error: %v", test, err)
			continue
		}

		if addr.String() != test {
			t.Errorf("String() round-trip = %q; want %q", addr, test)
			continue
		}

	}

	// Should fail
	badTests := []string{
		`<Abc.example.com>`,
		`<A@b@c@example.com>`,
		`<a"b(c)d,e:f;g<h>i[j\k]l@example.com>`,
		`<just"not"right@example.com>`,
		`<this is"not\allowed@example.com>`,
		`<this\ still\"not\\allowed@example.com>`,
		`<john..doe@example.com>`,
		`<john.doe@example..com>`,
		`<john.doe@example..com>`,
		`<john.doe.@example.com>`,
		`<john.doe.@.example.com>`,
		`<.john.doe@example.com>`,
		`<@example.com>`,
		`<.@example.com>`,
		`<test@.>`,
		`< @example.com>`,
		`<""test""blah""@example.com>`,
		`<""@0>`,
	}

	for _, test := range badTests {
		_, err := ParseAddress(test)
		if err == nil {
			t.Errorf("Should have failed to parse address: %s", test)
			continue
		}

	}

}

func TestAddressFormattingAndParsing(t *testing.T) {
	tests := List{
		{Name: "@lïce", Address: "alice@example.com"},
		{Name: "Böb O'Connor", Address: "bob@example.com"},
		{Name: "???", Address: "bob@example.com"},
		{Name: "Böb ???", Address: "bob@example.com"},
		{Name: "Böb (Jacöb)", Address: "bob@example.com"},
		{Name: "à#$%&'(),.:;<>@[]^`{|}~'", Address: "bob@example.com"},
		// https://golang.org/issue/11292
		{Name: "\"\\\x1f,\"", Address: "0@0"},
		// https://golang.org/issue/12782
		{Name: "naé, mée", Address: "test.mail@gmail.com"},
	}

	for i, test := range tests {
		parsed, err := ParseAddress(test.String())
		if err != nil {
			t.Errorf("test #%d: ParseAddr(%q) error: %v", i, test.String(), err)
			continue
		}
		if parsed.Name != test.Name {
			t.Errorf("test #%d: Parsed name = %q; want %q", i, parsed.Name, test.Name)
		}
		if parsed.Address != test.Address {
			t.Errorf("test #%d: Parsed address = %q; want %q", i, parsed.Address, test.Address)
		}
	}
}
*/
