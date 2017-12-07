// +build gofuzz

package mailaddress

// Fuzz it!
func Fuzz(data []byte) int {
	addr, _ := Parse(string(data))
	addr.AddressEncoded()
	addr.Domain()
	addr.Local()
	addr.NameEncoded()
	addr.String()
	addr.StringEncoded()
	addr.ToList()
	addr.Valid()
	addr.WithoutTag()

	list, _ := ParseList(string(data))
	list.ContainsAddress("")
	list.ContainsAddress(string(data))
	list.ContainsDomain("")
	list.ContainsDomain(string(data))
	list.Errors()
	list.Slice()
	list.Sort(ByAddress)
	list.String()
	list.StringEncoded()
	list.ValidAddresses()

	list = List{}
	list.UnmarshalJSON(data)
	list.ContainsAddress("")
	list.ContainsAddress(string(data))
	list.ContainsDomain("")
	list.ContainsDomain(string(data))
	list.Errors()
	list.Slice()
	list.Sort(ByAddress)
	list.String()
	list.StringEncoded()
	list.ValidAddresses()

	return 1
}
