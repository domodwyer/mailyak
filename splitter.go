package mailyak

import "io"

const maxLineLen = 60

// lineSplitter breaks the given input into lines of maxLineLen characters
// before writing a "\r\n" newline
type lineSplitter struct {
	w       io.Writer
	lineLen int
	maxLen  int
}

type lineSplitterBuilder struct{}

func (b lineSplitterBuilder) new(w io.Writer) io.Writer {
	return &lineSplitter{w: w, maxLen: maxLineLen}
}

func (w *lineSplitter) Write(p []byte) (int, error) {
	offset := 0
	breaks := (len(p) / w.maxLen)
	for i := 0; i < breaks; i++ {
		w.w.Write(p[offset : offset+w.maxLen])
		w.w.Write([]byte("\r\n"))
		offset += w.maxLen
	}

	w.w.Write(p[offset:])
	return (len(p) + (breaks * 2)), nil
}
