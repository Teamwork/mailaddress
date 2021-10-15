package mailaddress

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
)

var validAddresses = map[string]Address{
	"addr-spec@example.net":                {Address: "addr-spec@example.net"},
	"<angle-addr@example.net>":             {Name: "", Address: "angle-addr@example.net"},
	"First Last <mailbox@example.net>":     {Name: "First Last", Address: "mailbox@example.net"},
	`"First Last" <mailbox@example.net>`:   {Name: "First Last", Address: "mailbox@example.net"},
	`"<Firšt; Låšt>" <x@example.net>`:      {Name: `<Firšt; Låšt>`, Address: `x@example.net`},
	"dot.in.local.part@example.com":        {Address: "dot.in.local.part@example.com"},
	"Name <dot.in.local.part@example.com>": {Name: "Name", Address: "dot.in.local.part@example.com"},
	"tagged+tag@example.com":               {Address: "tagged+tag@example.com"},
	"Name <tagged+tag@example.com>":        {Name: "Name", Address: "tagged+tag@example.com"},
	"dashed-dash@ex-ample.com":             {Address: "dashed-dash@ex-ample.com"},
	"Name <dashed-dash@ex-ample.com>":      {Name: "Name", Address: "dashed-dash@ex-ample.com"},
	"example@example.verylongtld":          {Address: "example@example.verylongtld"},
	`Po "Wiśnasd" <asd@asd-def-24h.zxc>`:   {Name: `Po Wiśnasd`, Address: "asd@asd-def-24h.zxc"},

	`Uni العَرَبِية Cøde <x@example.net>`: {Name: "Uni العَرَبِية Cøde", Address: "x@example.net"},

	// TODO: Quoting the local part isn't supported (yet).
	//`"quoted"@example.com`:                     {Address: `"quoted"@example.com`},
	//`Name <"quoted"@example.com>`:              {Name: "Name", Address: `"quoted"@example.com`},
	//`"very.unusual.@.unusual.com"@example.com`: {Address: `"very.unusual.@.unusual.com"@example.com`},
	//"/#!$%&'*+-/=?^_`{}|~@example.org":         {Address: "/#!$%&'*+-/=?^_`{}|~@example.org"},

	//"\" \"@example.org": {Address: "\" \"@example.org"},

	//"\"()<>[]:,;@\\\"!#$%&'-/=?^_`{}| ~.a\"@example.org": {
	//	Address: "\"()<>[]:,;@\\\"!#$%&'-/=?^_`{}| ~.a\"@example.org",
	//},

	//`"very.(),:;<>[]\".VERY.\"very@\ \"very\".unusual"@strange.example.com`: {
	//	Address: `"very.(),:;<>[]\".VERY.\"very@\ \"very\".unusual"@strange.example.com`,
	//},

	// \ is 'invisible', but can escape ".
	`"esc \some\ \"quotes\"" <q@example.net>`: {Name: `esc some "quotes"`, Address: `q@example.net`},

	// You can stop and start quoting
	`"Martin" foo "Tournoij" <martin@example.net>`: {Name: `Martin foo Tournoij`, Address: `martin@example.net`},
	`"Martin"foo"Tournoij" <martin@example.net>`:   {Name: `MartinfooTournoij`, Address: `martin@example.net`},
	`'Martin' foo 'Tournoij' <martin@example.net>`: {Name: `'Martin' foo 'Tournoij'`, Address: `martin@example.net`},
	`'Martin foo Tournoij' <martin@example.net>`:   {Name: `Martin foo Tournoij`, Address: `martin@example.net`},

	// do not support this yet
	`'Martin foo's Tournoij' <martin@example.net>`:       {Name: `'Martin foo's Tournoij'`, Address: `martin@example.net`},
	`'Martin foo's foo's Tournoij' <martin@example.net>`: {Name: `'Martin foo's foo's Tournoij'`, Address: `martin@example.net`},

	// One-letter local-part, domain
	"a@b.c":        {Address: "a@b.c"},
	"Name <a@b.c>": {Name: "Name", Address: "a@b.c"},

	// We can parse the old deprecated "comment" style; per RFC 5322:
	//
	//     Note: Some legacy implementations used the simple form where the
	//     addr-spec appears without the angle brackets, but included the name
	//     of the recipient in parentheses as a comment following the addr-spec.
	//     Since the meaning of the information in a comment is unspecified,
	//     implementations SHOULD use the full name-addr form of the mailbox,
	//     instead of the legacy form, to specify the display name associated
	//     with a mailbox.  Also, because some legacy implementations interpret
	//     the comment, comments generally SHOULD NOT be used in address fields
	//     to avoid confusing such implementations.
	//
	// In spite this being explicitly deprecated for at least 15 years, some
	// systems still use this format – mainly cron daemons and (ironically)
	// MTAs.
	"MAILER-DAEMON@example.org (Mail Delivery System)": {Name: "", Address: "MAILER-DAEMON@example.org"},
	"hello (world) <foo@foo.foo>":                      {Name: "hello (world)", Address: "foo@foo.foo"},
	"MAILER-DAEMON@example.org ()":                     {Name: "", Address: "MAILER-DAEMON@example.org"},
	"hello () <foo@foo.foo>":                           {Name: "hello ()", Address: "foo@foo.foo"},

	// Newlines are folded
	`Martin
Tournoij
<martin@example.net>`: {Name: `Martin Tournoij`, Address: `martin@example.net`},

	// Leading/trailing whitespace
	" Martin <martin@example.com> ":        {Name: "Martin", Address: "martin@example.com"},
	" Martin<martin@example.com> ":         {Name: "Martin", Address: "martin@example.com"},
	"   Martin    <martin@example.com>   ": {Name: "Martin", Address: "martin@example.com"},

	// RFC 2047. We don't need to extensibly test it since we use Go's package,
	// and assume that works well.
	`=?utf-8?q?=E6=97=A5=E6=9C=AC=D0=BA=D0=B8=E6=AD=A3=E9=AB=94=E0=B8=AD?=` +
		`=?utf-8?q?=E0=B8=B1=E0=B8=81=E0=B8=A9=ED=9B=88=EB=AF=BC?= <a@example.net>`: {
		Name: `日本ки正體อักษ훈민`, Address: `a@example.net`},

	// Non-utf8
	"=?iso-8859-2?Q?Bogl=E1rka_Tak=E1cs?= <a@example.com>":    {Name: "Boglárka Takács", Address: "a@example.com"},
	`=?koi8-r?B?IvfMwcTJzcnSIPPFzcXOz9ci?= <boris@rusky.com>`: {Name: `"Владимир Семенов"`, Address: `boris@rusky.com`},
}

var invalidAddresses = []string{
	// Invalid encoding
	"=?GB2312?B?us6V08qk?= <secmocu@jshjkj.com>",
	"no.at.example.com",
	"multiple@at@signs@example.com",

	// none of the special characters in this local-part are allowed outside
	// quotation marks
	`a"b(c)d,e:f;gi[j\k]l@example.com`,

	// spaces, quotes, and backslashes may only exist when within quoted strings
	// and preceded by a backslash
	`this is"not\allowed@example.com`,

	// even if escaped (preceded by a backslash), spaces, quotes, and
	// backslashes must still be contained by quotes
	`this\ still\"not\allowed@example.com`,

	// sent from localhost
	"example@localhost",
	"admin@mailserver1",

	// double dot after @ – caveat: Gmail lets this through, Email
	// address#Local-part the dots altogether
	`john.doe@example..com`,

	// Multiple addresses.
	"multiple@example.com addresses@example.com",

	// Invalid UTF-8
	string([]byte{0xed, 0xa0, 0x80}) + " <micro@example.net>",
	"\"" + string([]byte{0xed, 0xa0, 0x80}) + "\" <half-surrogate@example.com>",
	"\"\\" + string([]byte{0x80}) + "\" <escaped-invalid-unicode@example.net>",

	// Don't allow unprintable characters.
	"\"\x00\" <null@example.net>",
	"\"\\\x00\" <escaped-null@example.net>",
	"asd\x06asd <null@example.net>",

	// Random data after >
	"foo <foo@example.com> huh",

	// Various junk data collected.
	`roby.bell@comcast.netVortex666!!`,
	`martimbault@.qc.aira.com`,
	`foo jlzuniga@comcast.net>`,
	`noreply@http://www.acadiapinesmotel.com`,
	`heartinternet.co.uk NO-REPLY@heartinternet.co.uk`,
	`Arthurgrebenuk@gmail`,
	`lgomberg@dmh.co.la.ca.usordiglg@earth`,
	`smellycats612@aol..com`,
	`concretesawing@comcast.net13113602@dwsg`,
	`MM522@aol.com315=269-5244`,
	`"much.more unusual"@example.com`,

	// quoted strings must be dot separated or the only element making up the
	// local-part.
	//`just"not"right@example.com`,

	// Technically valid, but we don't want to accept this.
	//TODO: accepted as valid "user@192.168.1.1",
	"user@[IPv6:2001:DB8::1]",
}

func TestParseAddress(t *testing.T) {
	for test, expected := range validAddresses {
		t.Run(test, func(t *testing.T) {
			out, err := Parse(test)
			if err != nil {
				t.Fatal(err)
			}

			if !cmpaddr(out, expected) {
				t.Errorf("\n  out:      %v\n  expected: %v\n",
					fmtaddr(out), fmtaddr(expected))
			}
		})
	}
}

func TestParseStringOutput(t *testing.T) {
	for test, expected := range validAddresses {
		t.Run(test, func(t *testing.T) {
			parsed, err := Parse(test)
			// TestParseAddress will show an error already.
			if err != nil || !cmpaddr(parsed, expected) {
				t.Skip()
			}

			out, err := Parse(parsed.StringEncoded())
			if err != nil {
				t.Fatal(err)
			}

			if !cmpaddr(out, expected) {
				t.Errorf("\n  out:      %v\n  expected: %v\n",
					fmtaddr(out), fmtaddr(expected))
			}
		})
	}
}

func TestParseListAddress(t *testing.T) {
	// Mix them up randomly to test ParseList
	var keys []string
	for k := range validAddresses {
		keys = append(keys, k)
	}

	mixed := make(map[string][]Address)
	for test, expected := range validAddresses {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(validAddresses)-3)))

		newTest := test
		newExpected := []Address{expected}

		for i := int64(0); i < n.Int64(); i++ {
			rnd, _ := rand.Int(rand.Reader, big.NewInt(int64(len(validAddresses))))
			pick := keys[rnd.Int64()]

			newTest += ", " + pick
			newExpected = append(newExpected, validAddresses[pick])
		}
		mixed[newTest] = List(newExpected).uniq()
	}

	for test, expected := range mixed {
		t.Run(test, func(t *testing.T) {
			out, gotErr := ParseList(test)
			if gotErr {
				t.Errorf("gotErr is true; errors: %#v", out.Errors())
			}

			if !cmplist(out, expected) {
				t.Fatalf("Error different length:\n  out     : %#v\n  expected: %#v\n",
					out, expected)
			}

			for i := range out {
				if !cmpaddr(out[i], expected[i]) {
					t.Errorf("Error:\n  out     : %#v\n  expected: %#v\n",
						out, expected)
				}
			}

			// Make sure we can re-parse our output.
			t.Run("reparse", func(t *testing.T) {
				reparsedOut, gotErr := ParseList(out.StringEncoded())
				if gotErr {
					t.Errorf("gotErr is true; errors: %#v", out.Errors())
				}

				if !cmplist(reparsedOut, expected) {
					t.Fatalf("Error different length:\n  out     : %#v\n  expected: %#v\n",
						reparsedOut, expected)
				}

				for i := range reparsedOut {
					if !cmpaddr(reparsedOut[i], expected[i]) {
						t.Errorf("Error:\n  out     : %#v\n  expected: %#v\n",
							reparsedOut, expected)
					}
				}
			})
		})
	}
}

func TestString(t *testing.T) {
	cases := map[string]string{
		`martin@example.net`:                              `martin@example.net`,
		`<martin@example.net>`:                            `martin@example.net`,
		`Martin Tournoij <martin@example.net>`:            `"Martin Tournoij" <martin@example.net>`,
		`"Martin Tournoij" <martin@example.net>`:          `"Martin Tournoij" <martin@example.net>`,
		`Martin العَرَبِية Tournoij <martin@example.net>`: `"Martin العَرَبِية Tournoij" <martin@example.net>`,
		`"<Martin; Tøurnoij>" <martin@example.net>`:       `"<Martin; Tøurnoij>" <martin@example.net>`,
		`"Martin \a\ \"Tournoij\"" <martin@example.net>`:  `"Martin a \"Tournoij\"" <martin@example.net>`,
		`"Martin" foo "Tournoij" <martin@example.net>`:    `"Martin foo Tournoij" <martin@example.net>`,
		`"Martin"foo"Tournoij" <martin@example.net>`:      `"MartinfooTournoij" <martin@example.net>`,
		`Martin
Tournoij
<martin@example.net>`: `"Martin Tournoij" <martin@example.net>`,

		`=?utf-8?q?=E6=97=A5=E6=9C=AC=D0=BA=D0=B8=E6=AD=A3=E9=AB=94=E0=B8=AD?=` +
			`=?utf-8?q?=E0=B8=B1=E0=B8=81=E0=B8=A9=ED=9B=88=EB=AF=BC?= <a@example.net>`: `` +
			`"日本ки正體อักษ훈민" <a@example.net>`,

		``:                     ``,
		`,`:                    ``,
		`martin@example.com,`:  `martin@example.com`,
		`,martin@example.com`:  `martin@example.com`,
		`,martin@example.com,`: `martin@example.com`,
		`"martin@example.com FOO" <martin@example.com>`: `"martin@example.com FOO" <martin@example.com>`,
	}

	for test, expected := range cases {
		t.Run(test, func(t *testing.T) {
			out, gotErr := ParseList(test)
			if gotErr {
				t.Error(out.Errors())
			}

			if out.String() != expected {
				t.Errorf("\nout:      %v\nexpected: %v\nlist:     %#v\n",
					out.String(), expected, out)
			}
		})
	}
}

func TestInvalid(t *testing.T) {
	for i, test := range invalidAddresses {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			out, err := Parse(test)
			if out.Valid() {
				t.Fatal("out.Valid() said it was valid.")
			}
			if err == nil {
				t.Fatal("err == nil")
			}
			if out.Error() == nil {
				t.Fatal("out.Error == nil")
			}
		})
	}
}

func fmtaddr(a Address) string {
	return fmt.Sprintf("Name: %v, Address: %v", a.Name, a.Address)
}

func cmpaddr(a, b Address) bool {
	return a.Address == b.Address && a.Name == b.Name
}

func cmplist(a, b List) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !(a[i].Address == b[i].Address && a[i].Name == b[i].Name) {
			return false
		}
	}

	return true
}
