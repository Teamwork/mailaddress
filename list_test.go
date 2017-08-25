package mailaddress

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/teamwork/test"
	"github.com/teamwork/test/diff"
)

func TestAppend(t *testing.T) {
	cases := []struct {
		name, address string
		in, expected  List
	}{
		{
			"Kees", "kees@example.com",
			List{Address{Name: "Martin", Address: "martin@example.com"}},
			List{
				Address{Name: "Martin", Address: "martin@example.com"},
				Address{Name: "Kees", Address: "kees@example.com"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%v-%v", tc.name, tc.address), func(t *testing.T) {
			tc.in.Append(tc.name, tc.address)

			if diff.Diff(tc.expected, tc.in) != "" {
				t.Errorf(diff.Cmp(tc.expected, tc.in))
			}
		})
	}
}

func TestJSON(t *testing.T) {
	cases := []struct {
		in          string
		expected    []string
		expectedErr string
	}{
		{"", []string{}, "unexpected end of JSON input"},
		{"[", []string{}, "unexpected end of JSON input"},
		{`["invalid"]`, []string{}, ""},

		{
			`["robert@teamwork.com", "martin@beanwork.com"]`,
			[]string{"robert@teamwork.com", "martin@beanwork.com"},
			"",
		},
		{
			`["robert@teamwork.com", "martin.com"]`,
			[]string{"robert@teamwork.com"},
			"",
		},
		{
			`"robert@teamwork.com, beanwork@teamstyle.org"`,
			[]string{"robert@teamwork.com", "beanwork@teamstyle.org"},
			"",
		},
		{
			`"robert@teamwork.com, beanwork"`,
			[]string{"robert@teamwork.com"},
			"",
		},
		{
			`[{"name": "Robert O'Leary", "address": "rob@teamwork.com"}]`,
			[]string{"rob@teamwork.com"},
			"",
		},
		{
			`[
				{"name": "Robert O'Leary", "address": "rob@teamwork.com"},
				{"name": "Brandon Hansen", "address": "brandon@dreamwork.ie"}
			]`,
			[]string{"rob@teamwork.com", "brandon@dreamwork.ie"},
			"",
		},
		{
			`[
				{"name": "Robert O'Leary", "address": "rob@teamwork.com"},
				{"name": "Brandon Hansen", "address": "brandon@dreamwork.ie"},
				{"name": "bad email", "address": "bad"}
			]`,
			[]string{"rob@teamwork.com", "brandon@dreamwork.ie"},
			"",
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			var list List
			gotErr := json.Unmarshal([]byte(tc.in), &list)

			if !test.ErrorContains(gotErr, tc.expectedErr) {
				t.Errorf("got error %#v – expected %#v", gotErr, tc.expectedErr)
			}

			out := list.Slice()
			if diff.Diff(tc.expected, out) != "" {
				t.Errorf(diff.Cmp(tc.expected, out))
			}
		})
	}
}

func TestToSlice(t *testing.T) {
	cases := []struct {
		in       List
		expected []string
	}{
		{List{}, []string{}},
		{List{Address{}}, []string{}},

		{
			List{Address{Name: "Martin", Address: ""}},
			[]string{},
		},
		{
			List{Address{Name: "Martin", Address: "martin@example.com"}},
			[]string{"martin@example.com"},
		},
		{
			List{Address{Name: "Martin", Address: "martin@example.com"}, Address{Name: "", Address: "a@b.c"}},
			[]string{"martin@example.com", "a@b.c"},
		},
		{
			List{Address{Name: "Martin", Address: "martin@example.com"}, Address{}, Address{Name: "", Address: "a@b.c"}},
			[]string{"martin@example.com", "a@b.c"},
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%v", tc.in), func(t *testing.T) {
			out := tc.in.Slice()

			if diff.Diff(tc.expected, out) != "" {
				t.Errorf(diff.Cmp(tc.expected, out))
			}
		})
	}
}

func TestSort(t *testing.T) {
	cases := []struct {
		in       List
		sortKey  int8
		expected List
	}{
		{List{}, ByName, List{}},
		{List{Address{}}, ByName, List{Address{}}},

		{
			List{Address{Name: "zz"}, Address{Name: "aa"}},
			ByName,
			List{Address{Name: "aa"}, Address{Name: "zz"}},
		},
		{
			List{Address{Address: "zz@zz.zz"}, Address{Address: "aa@aa.aa"}},
			ByAddress,
			List{Address{Address: "aa@aa.aa"}, Address{Address: "zz@zz.zz"}},
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%v", tc.in), func(t *testing.T) {
			tc.in.Sort(tc.sortKey)

			if diff.Diff(tc.expected, tc.in) != "" {
				t.Errorf(diff.Cmp(tc.expected, tc.in))
			}
		})
	}
}

func TestErrors(t *testing.T) {
	cases := []struct {
		in                 func() (List, bool)
		expectedHaveErr    bool
		expectedError      string
		expectedMultiError *multierror.Error
	}{
		{
			func() (List, bool) { return ParseList("martin@example.com") },
			false,
			"",
			nil,
		},
		{
			func() (List, bool) { return ParseList("martin@example.com, invalid") },
			true,
			"1 error occurred:",
			&multierror.Error{
				Errors: []error{errors.New("unable to find an email address")},
			},
		},
		{
			func() (List, bool) { return ParseList("@example.com") },
			true,
			"1 error occurred:",
			&multierror.Error{
				Errors: []error{errors.New("unable to find an email address")},
			},
		},
		{
			func() (List, bool) { return ParseList("@example.com, invalid") },
			true,
			"2 errors occurred:",
			&multierror.Error{
				Errors: []error{
					errors.New("unable to find an email address"),
					errors.New("unable to find an email address"),
				},
			},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			gotList, gotHaveErr := tc.in()

			if gotHaveErr != tc.expectedHaveErr {
				t.Errorf("expected haveErr %v, got %v",
					tc.expectedHaveErr, gotHaveErr)
			}

			if !test.ErrorContains(gotList.Errors(), tc.expectedError) {
				t.Errorf("wrong errors\nexpected: %#v\ngot     : %v\n",
					tc.expectedError, gotList.Errors())
			}

			if gotList.Errors() != nil {
				merr, ok := gotList.Errors().(*multierror.Error)
				if !ok {
					t.Fatalf("cannot convert %#v to multierror", gotList.Errors())

				}
				if diff.Diff(merr, tc.expectedMultiError) != "" {
					t.Fatalf("multierror didn't match:\ngot     : %+v\nexpected: %+v\n",
						merr, tc.expectedMultiError)
				}
			}
		})
	}
}

func TestValidAddresses(t *testing.T) {
	cases := []struct {
		in       List
		expected List
	}{
		{List{}, List{}},
		{
			List{Address{Name: "foo", Address: "foo@example.com"}},
			List{Address{Name: "foo", Address: "foo@example.com"}},
		},
		{
			List{
				Address{Name: "foo", Address: "foo@example.com"},
				Address{Name: "foo", Address: "not valid"},
			},
			List{
				Address{Name: "foo", Address: "foo@example.com"},
			},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			out := tc.in.ValidAddresses()
			if out.String() != tc.expected.String() {
				t.Errorf("\nout:      %#v\nexpected: %#v\n", out, tc.expected)
			}
		})
	}
}

func TestContainsAddress(t *testing.T) {
	cases := []struct {
		in       List
		test     string
		expected bool
	}{
		{List{}, "", false},
		{List{Address{Address: "FOO@EXAMPLE.COM"}}, "foo@example.com", true},
		{List{Address{Address: "f€@Ü.русские"}}, "f€@ü.русские", true},
		{List{Address{Address: "f€@Ü.русские"}}, "f€@ü.рсские", false},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			out := tc.in.ContainsAddress(tc.test)
			if out != tc.expected {
				t.Errorf("\nout:      %#v\nexpected: %#v\n", out, tc.expected)
			}
		})
	}
}

func TestContainsDomain(t *testing.T) {
	cases := []struct {
		in       List
		test     string
		expected bool
	}{
		{List{}, "", false},
		{List{Address{Address: "FOO@EXAMPLE.COM"}}, "example.com", true},
		{List{Address{Address: "f€@Ü.русские"}}, "ü.русские", true},
		{List{Address{Address: "f€@Ü.русские"}}, "ü.рсские", false},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			out := tc.in.ContainsDomain(tc.test)
			if out != tc.expected {
				t.Errorf("\nout:      %#v\nexpected: %#v\n", out, tc.expected)
			}
		})
	}
}
