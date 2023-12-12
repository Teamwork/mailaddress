package mailaddress

import (
	"fmt"
	"mime"
	"testing"

	"github.com/teamwork/test/diff"
)

func TestNameEncoded(t *testing.T) {
	cases := []struct {
		in, expected string
	}{
		{"", ""},
		{"martin", "martin"},
		{"m€rtin", "=?utf-8?q?m=E2=82=ACrtin?="},
		{"martin, tournoij", `"martin, tournoij"`},
		{"m@rtin", `"m@rtin"`},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%v", tc.in), func(t *testing.T) {
			out := Address{Name: tc.in}.NameEncoded()
			if out != tc.expected {
				t.Errorf("\nout:      %v\nexpected: %v\n", out, tc.expected)
			}
		})
	}
}

func TestAddressEncoded(t *testing.T) {
	cases := []struct {
		in, expected string
	}{
		{"", ""},
		{"martin", "martin"},
		{"m€rtin", "=?utf-8?q?m=E2=82=ACrtin?="},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%v", tc.in), func(t *testing.T) {
			out := Address{Address: tc.in}.AddressEncoded()
			if out != tc.expected {
				t.Errorf("\nout:      %v\nexpected: %v\n", out, tc.expected)
			}
		})
	}
}

func TestStringEncoded(t *testing.T) {
	cases := []struct {
		in, expected string
		count        int
	}{
		{`martin@example.net`, `martin@example.net`, 1},
		{`Martin Tournoij <martin@example.net>`, `Martin Tournoij <martin@example.net>`, 1},
		{`"Martin Tournoij" <martin@example.net>`, `Martin Tournoij <martin@example.net>`, 1},
		{`Martin Tour<noij> <martin.t@example.com>`, `Martin Tour noij <martin.t@example.com>`, 1},
		{
			`Martin, Nichole <Nichole.Martin@harrisgroup.com <mailto:Nichole.Martin@harrisgroup.com>>r`,
			"Nichole <Nichole.Martin@harrisgroup.com>",
			1,
		},
		{
			`a العَرَبِي b <a@example.net>`,
			`=?utf-8?q?a_=D8=A7=D9=84=D8=B9=D9=8E=D8=B1=D9=8E=D8=A8=D9=90=D9=8A_b?= <a@example.net>`,
			1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			out, outErr := ParseList(tc.in)
			out = out.ValidAddresses()
			if outErr && len(out) != tc.count {
				t.Error(out.Errors(), " count:", len(out))
			}

			if out.StringEncoded() != tc.expected {
				t.Errorf("\nout:      %#v\nexpected: %#v\n",
					out.StringEncoded(), tc.expected)
			}

			dec := new(mime.WordDecoder)
			_, err := dec.DecodeHeader(out.StringEncoded())
			if err != nil {
				t.Errorf("can't parse %s: %s", out.StringEncoded(), err.Error())
			}
		})
	}
}

func TestToList(t *testing.T) {
	cases := []struct {
		in       Address
		expected List
	}{
		{Address{}, List{Address{}}},
		{Address{Name: "Martin", Address: ""}, List{Address{Name: "Martin", Address: ""}}},
		{Address{Name: "Martin", Address: "martin@"}, List{Address{Name: "Martin", Address: "martin@"}}},
		{
			Address{Name: "Martin", Address: "martin@example.com"},
			List{Address{Name: "Martin", Address: "martin@example.com"}},
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%v", tc.in), func(t *testing.T) {
			out := tc.in.ToList()

			if diff.Diff(tc.expected, out) != "" {
				t.Errorf(diff.Cmp(tc.expected, out))
			}
		})
	}
}

func TestLocal(t *testing.T) {
	cases := []struct {
		in       Address
		expected string
	}{
		{Address{}, ""},
		{Address{Address: "martin"}, "martin"},
		{Address{Address: "martin@example.com"}, "martin"},
		{Address{Address: "martin+tag@example.com"}, "martin+tag"},
		{Address{Address: "martin@example.co.com.many.domains"}, "martin"},
		{Address{Address: "@example.com"}, ""},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%v", tc.in), func(t *testing.T) {
			out := tc.in.Local()

			if out != tc.expected {
				t.Errorf("\nout:      %#v\nexpected: %#v\n", out, tc.expected)
			}

		})
	}
}

func TestDomain(t *testing.T) {
	cases := []struct {
		in       Address
		expected string
	}{
		{Address{}, ""},
		{Address{Address: "martin"}, "martin"},
		{Address{Address: "martin@example.com"}, "example.com"},
		{Address{Address: "martin@example.co.com.many.domains"}, "example.co.com.many.domains"},
		{Address{Address: "@example.com"}, "example.com"},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%v", tc.in), func(t *testing.T) {
			out := tc.in.Domain()

			if out != tc.expected {
				t.Errorf("\nout:      %#v\nexpected: %#v\n", out, tc.expected)
			}

		})
	}
}

func TestWithoutTag(t *testing.T) {
	cases := []struct {
		in       Address
		expected string
	}{
		{Address{}, ""},
		{Address{Address: "martin@example.com"}, "martin@example.com"},
		{Address{Address: "martin+tag@example.com"}, "martin@example.com"},
		{Address{Address: "martin+tag+tag@example.com"}, "martin@example.com"},

		// Don't support - separated tags
		{Address{Address: "martin-tag@example.com"}, "martin-tag@example.com"},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%v", tc.in), func(t *testing.T) {
			out := tc.in.WithoutTag()
			if out != tc.expected {
				t.Errorf("\nout:      %#v\nexpected: %#v\n", out, tc.expected)
			}

		})
	}
}
