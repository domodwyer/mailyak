package mailyak

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
)

func (m *MailYak) buildMime() (*bytes.Buffer, error) {
	mb, err := randomBoundary()
	if err != nil {
		return nil, err
	}

	ab, err := randomBoundary()
	if err != nil {
		return nil, err
	}

	return m.buildMimeWithBoundaries(mb, ab)
}

func randomBoundary() (string, error) {
	buf := make([]byte, 30)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", buf), nil
}

// buildMimeWithBoundaries creates the MIME message and returns it as a buffer
func (m *MailYak) buildMimeWithBoundaries(mb, ab string) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	m.writeHeaders(&buf)

	// Start our multipart/mixed part
	mixed := multipart.NewWriter(&buf)
	mixed.SetBoundary(mb)
	defer mixed.Close()

	fmt.Fprintf(&buf, "Content-Type: multipart/mixed;\r\n\tboundary=\"%s\"; charset=UTF-8\r\n\r\n", mixed.Boundary())

	ctype := fmt.Sprintf("multipart/alternative;\r\n\tboundary=\"%s\"", ab)

	altPart, err := mixed.CreatePart(textproto.MIMEHeader{"Content-Type": {ctype}})
	if err != nil {
		return nil, err
	}

	if err := m.writeBody(altPart, ab); err != nil {
		return nil, err
	}

	if err := m.writeAttachments(mixed, lineSplitterBuilder{}); err != nil {
		return nil, err
	}

	return &buf, nil
}

// writeHeaders writes the Mime-Version, Reply-To, From, To and Subject headers
func (m *MailYak) writeHeaders(buf io.Writer) {

	buf.Write([]byte(m.fromHeader()))
	buf.Write([]byte("Mime-Version: 1.0\r\n"))

	if m.replyTo != "" {
		fmt.Fprintf(buf, "Reply-To: %s\r\n", m.replyTo)
	}

	fmt.Fprintf(buf, "Subject: %s\r\n", m.subject)

	for _, to := range m.toAddrs {
		fmt.Fprintf(buf, "To: %s\r\n", to)
	}

	for _, bcc := range m.bccAddrs {
		fmt.Fprintf(buf, "BCC: %s\r\n", bcc)
	}
}

// fromHeader returns a correctly formatted From header, optionally with a name
// component
func (m *MailYak) fromHeader() string {
	if m.fromName == "" {
		return fmt.Sprintf("From: %s\r\n", m.fromAddr)
	}

	return fmt.Sprintf("From: %s <%s>\r\n", m.fromName, m.fromAddr)
}

// writeBody writes the text/plain and text/html mime parts
func (m *MailYak) writeBody(w io.Writer, boundary string) error {
	alt := multipart.NewWriter(w)
	defer alt.Close()

	alt.SetBoundary(boundary)

	var err error
	writePart := func(ctype string, data []byte) {
		if len(data) == 0 || err != nil {
			return
		}

		c := fmt.Sprintf("%s; charset=UTF-8", ctype)

		part, err2 := alt.CreatePart(textproto.MIMEHeader{"Content-Type": {c}})
		if err2 != nil {
			err = err2
			return
		}

		part.Write(data)
	}

	writePart("text/plain", m.plain.Bytes())
	writePart("text/html", m.html.Bytes())

	return err
}
