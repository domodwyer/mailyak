package mailyak

// BodyPart is a byte slice that implements io.Writer
type BodyPart []byte

func (w *BodyPart) Write(p []byte) (int, error) {
	*w = append(*w, p...)
	return len(p), nil
}

// String returns the byte slice as a string in fmt.Printx(), etc
func (w BodyPart) String() string {
	return string(w)
}

// Set accepts a string as the contents of a BodyPart
func (w *BodyPart) Set(s string) {
	*w = []byte(s)
}
