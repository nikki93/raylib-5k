package str

//gx:extern std::isspace
func IsSpace(c byte) bool

//gx:extern &
func address(b byte) *byte

func Set(s *string, i int, b byte) {
	*address((*s)[i]) = b
}

//gx:extern (const char *)
func constCharPtr(b *byte) string

func CString(s *string, i int) string {
	return constCharPtr(address((*s)[i]))
}

//gx:extern cprint
func Print(fmt string, args ...interface{})

//gx:extern dprint
func Display(fmt string, args ...interface{})
