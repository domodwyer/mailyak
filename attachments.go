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

// AttachmentEmailer extends the Emailer interface to provide attachment support
type AttachmentEmailer interface {
	Emailer
	Attach(name string, data io.Reader)
	ClearAttachments()
}

type partCreator interface {
	CreatePart(header textproto.MIMEHeader) (io.Writer, error)
}

type writeWrapper interface {
	new(w io.Writer) io.Writer
}

type attachment struct {
	filename string
	content  io.Reader
}

// Attach adds an attachment to the email with the given filename.
//
// Note: The attachment data isn't read until Send() is called
func (m *MailYak) Attach(name string, r io.Reader) {
	m.attachments = append(m.attachments, attachment{
		filename: name,
		content:  r,
	})
}

// ClearAttachments removes all current attachments
func (m *MailYak) ClearAttachments() {
	m.attachments = []attachment{}
}

// writeAttachments loops over the attachments, guesses their content-type and
// writes the data as a line-broken base64 string (using the splitter mutator)
func (m *MailYak) writeAttachments(mixed partCreator, splitter writeWrapper) error {
	h := make([]byte, sniffLen)

	for _, item := range m.attachments {
		hLen, err := item.content.Read(h)
		if err != nil && err != io.EOF {
			return err
		}

		ctype := fmt.Sprintf("%s;\n\tfilename=%s", http.DetectContentType(h[:hLen]), item.filename)
		disp := fmt.Sprintf("attachment;\n\tfilename=%s", item.filename)

		part, err := mixed.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {ctype},
			"Content-Disposition":       {disp},
			"Content-Transfer-Encoding": {"base64"},
		})
		if err != nil {
			return err
		}

		encoder := base64.NewEncoder(base64.StdEncoding, splitter.new(part))
		encoder.Write(h[:hLen])

		// More to write?
		if hLen == sniffLen {
			io.Copy(encoder, item.content)
		}

		encoder.Close()
	}

	return nil
}
