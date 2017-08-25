/*
Package mailaddress deals with email addresses.

It's an alternative to net/mail; significant differences include:
- Better errors.
- When parsing a list it will continue to the next address on an error; this is
  especially useful when providing feedback to users.
- Some useful utility functions.

The Address and List types are compatible with the net/mail types.
*/
package mailaddress // import "github.com/teamwork/mailaddress"
