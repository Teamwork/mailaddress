package mailaddress

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/teamwork/test"
	"github.com/teamwork/test/diff"
)

func TestParseList(t *testing.T) {
	// This just tests the expected return values and the like. Testing the full
	// parsing logic is in parse_test.go.
	cases := []struct {
		str           string
		expected      List
		expectedValid bool
	}{
		{``, List{}, false},
		{`asd`, List{Address{Raw: "asd"}}, true},
		{"Martin <martin@example.com>", List{Address{Name: "Martin", Address: "martin@example.com"}}, false},
	}

	for _, tc := range cases {
		t.Run(tc.str, func(t *testing.T) {
			got, gotValid := ParseList(tc.str)
			if !cmplist(tc.expected, got) {
				t.Errorf(diff.Cmp(tc.expected, got))
			}

			if gotValid != tc.expectedValid {
				t.Errorf("gotValid: %v, expectedValid: %v",
					gotValid, tc.expectedValid)
			}
		})
	}
}

func TestParse(t *testing.T) {
	// This just tests the expected return values and the like. Testing the
	// parsing logic is in parse_test.go.
	cases := []struct {
		str         string
		expected    Address
		expectedErr string
	}{
		{"Martin <martin@example.com>", Address{Name: "Martin", Address: "martin@example.com"}, ""},
		{"Amy <acollins@edgeclub.edu,>", Address{Name: "Amy", Address: "acollins@edgeclub.edu"}, ""},
		{"Martin <martin@example.com>, another@foo.com", Address{}, ErrTooManyEmails.Error()},
	}

	for _, tc := range cases {
		t.Run(tc.str, func(t *testing.T) {
			got, gotErr := Parse(tc.str)

			if !test.ErrorContains(gotErr, tc.expectedErr) {
				t.Errorf("wrong error\nexpected: %#v\ngot     : %v\n",
					tc.expectedErr, gotErr)
			}

			if !cmpaddr(got, tc.expected) {
				t.Errorf(diff.Cmp(tc.expected, got))
			}
		})
	}
}

func TestNewHelpers(t *testing.T) {
	cases := []struct {
		name, address string
		expected      Address
	}{
		{"", "", Address{Name: "", Address: "", err: errors.New("")}},
		{"Martin", "", Address{Name: "Martin", Address: "", err: errors.New("")}},
		{"", "martin@example.com", Address{Name: "", Address: "martin@example.com"}},
		{"Martin", "martin@example.com", Address{Name: "Martin", Address: "martin@example.com"}},

		// Invalid addresses should result in an invalid Address{} (that is, one
		// without the Address field set).
		{"Martin", "invalid", Address{Name: "Martin", Address: "", err: errors.New("")}},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%v <%v>", tc.name, tc.address), func(t *testing.T) {
			got := New(tc.name, tc.address)

			if tc.expected.Error() != nil && got.Error() == nil {
				t.Fatal("expected error but got nil")
			}
			if tc.expected.Error() == nil && got.Error() != nil {
				t.Fatal("expected no error but got error")
			}

			if got.Name != tc.expected.Name || got.Address != tc.expected.Address {
				t.Errorf(diff.Cmp(tc.expected, got))
			}

			gotList := NewList(tc.name, tc.address)

			if len(gotList) != 1 || gotList[0].Name != tc.expected.Name || gotList[0].Address != tc.expected.Address {
				t.Errorf(diff.Cmp(tc.expected, gotList[0]))
			}
		})
	}
}

func TestFromMap(t *testing.T) {
	cases := []struct {
		in       map[string]string
		expected List
	}{
		{map[string]string{}, *new(List)},
		{nil, *new(List)},

		{
			map[string]string{"Martin": "martin@example.com"},
			List{Address{Name: "Martin", Address: "martin@example.com"}},
		},
		//{
		//	map[string]string{"Martin": "martin@example.com", "foo": "bar@example.com"},
		//	List{
		//		Address{Name: "Martin", Address: "martin@example.com"},
		//		Address{Name: "foo", Address: "bar@example.com"},
		//	},
		//},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%#v", tc.in), func(t *testing.T) {
			got := FromMap(tc.in)
			if !reflect.DeepEqual(tc.expected, got) {
				t.Errorf(diff.Cmp(tc.expected, got))
			}
		})
	}
}

func TestFromSlice(t *testing.T) {
	cases := []struct {
		in       []string
		expected List
	}{
		{[]string{}, *new(List)},
		{nil, *new(List)},

		{
			[]string{"martin@example.com"},
			List{Address{Name: "", Address: "martin@example.com"}},
		},
		{
			[]string{"martin@example.com", "foo@example.com"},
			List{Address{Address: "martin@example.com"}, Address{Address: "foo@example.com"}},
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%#v", tc.in), func(t *testing.T) {
			got := FromSlice(tc.in)
			if diff.Diff(tc.expected, got) != "" {
				t.Errorf(diff.Cmp(tc.expected, got))
			}
		})
	}
}
