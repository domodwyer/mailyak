package mailyak

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
)

// DetectContentType needs at most 512 bytes
const sniffLen = 512

type partCreator interface {
	CreatePart(header textproto.MIMEHeader) (io.Writer, error)
}

type writeWrapper interface {
	new(w io.Writer) io.Writer
}

type attachment struct {
	filename string
	content  io.Reader
	inline   bool
}

func (m *MailYak) Attach(name string, r io.Reader) {
	m.attach(name, r, false)
}

func (m *MailYak) AttachInline(name string, r io.Reader) {
	m.attach(name, r, true)
}

// Attach adds an attachment to the email with the given filename.
//
// The attachment data isn't read until Send() is called.
func (m *MailYak) attach(name string, r io.Reader, inline bool) {
	m.attachments = append(m.attachments, attachment{
		filename: name,
		content:  r,
		inline: inline,
	})
}

// ClearAttachments removes all current attachments.
func (m *MailYak) ClearAttachments() {
	m.attachments = []attachment{}
}

// writeAttachments loops over the attachments, guesses their content-type and
// writes the data as a line-broken base64 string (using the splitter mutator).
func (m *MailYak) writeAttachments(mixed partCreator, splitter writeWrapper) error {
	h := make([]byte, sniffLen)

	for _, item := range m.attachments {
		hLen, err := item.content.Read(h)
		if err != nil && err != io.EOF {
			return err
		}

		ctype := fmt.Sprintf("%s;\n\tfilename=%s", http.DetectContentType(h[:hLen]), item.filename)

		part, err := mixed.CreatePart(getMIMEHeader(item, ctype))
		if err != nil {
			return err
		}

		encoder := base64.NewEncoder(base64.StdEncoding, splitter.new(part))
		if _, err := encoder.Write(h[:hLen]); err != nil {
			return err
		}

		// More to write?
		if hLen == sniffLen {
			if _, err := io.Copy(encoder, item.content); err != nil {
				return err
			}
		}

		if err := encoder.Close(); err != nil {
			return err
		}
	}

	return nil
}

func getMIMEHeader(a attachment, ctype string) textproto.MIMEHeader {
	var disp string
	var header textproto.MIMEHeader

	if a.inline {
		disp = fmt.Sprintf("inline;\n\tfilename=%s", a.filename)
		header = textproto.MIMEHeader{
			"Content-Type":              {ctype},
			"Content-Disposition":       {disp},
			"Content-Transfer-Encoding": {"base64"},
		}
	} else {
		disp = fmt.Sprintf("attachment;\n\tfilename=%s", a.filename)
		cid := fmt.Sprintf("<%s>", a.filename)
		header = textproto.MIMEHeader{
			"Content-Type":              {ctype},
			"Content-Disposition":       {disp},
			"Content-Transfer-Encoding": {"base64"},
			"Content-ID":				 {cid},
		}
	}

	return header
}
