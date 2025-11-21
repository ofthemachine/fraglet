package fraglet

// FragletProc is just code - the portable unit
type FragletProc string

// NewFragletProc creates a FragletProc from code string
func NewFragletProc(code string) FragletProc {
	return FragletProc(code)
}

// Code returns the code string
func (p FragletProc) Code() string {
	return string(p)
}

