package mailyak

import "bytes"

// BodyPart is a buffer.
type BodyPart struct{ bytes.Buffer }

// Set accepts a string as the contents of a BodyPart and replaces any existing data.
func (w *BodyPart) Set(s string) {
	w.Reset()
	w.WriteString(s)
}
