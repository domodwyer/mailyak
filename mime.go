package mailyak

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"mime/multipart"
	"mime/quotedprintable"
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

// randomBoundary returns a random hexadecimal string used for separating MIME
// parts.
//
// The returned string must be sufficiently random to prevent malicious users
// from performing content injection attacks.
func randomBoundary() (string, error) {
	buf := make([]byte, 30)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", buf), nil
}

// buildMimeWithBoundaries creates the MIME message using mb and ab as MIME
// boundaries, and returns the generated MIME data as a buffer.
func (m *MailYak) buildMimeWithBoundaries(mb, ab string) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	if err := m.writeHeaders(&buf); err != nil {
		return nil, err
	}

	// Start our multipart/mixed part
	mixed := multipart.NewWriter(&buf)
	if err := mixed.SetBoundary(mb); err != nil {
		return nil, err
	}
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

// writeHeaders writes the Mime-Version, Date, Reply-To, From, To and Subject headers,
// plus any custom headers set via AddHeader().
func (m *MailYak) writeHeaders(buf io.Writer) error {

	if _, err := buf.Write([]byte(m.fromHeader())); err != nil {
		return err
	}

	if _, err := buf.Write([]byte("Mime-Version: 1.0\r\n")); err != nil {
		return err
	}

	fmt.Fprintf(buf, "Date: %s\r\n", m.date)

	if m.replyTo != "" {
		fmt.Fprintf(buf, "Reply-To: %s\r\n", m.replyTo)
	}

	fmt.Fprintf(buf, "Subject: %s\r\n", m.subject)

	for _, to := range m.toAddrs {
		fmt.Fprintf(buf, "To: %s\r\n", to)
	}

	for _, cc := range m.ccAddrs {
		fmt.Fprintf(buf, "CC: %s\r\n", cc)
	}

	if m.writeBccHeader {
		for _, bcc := range m.bccAddrs {
			fmt.Fprintf(buf, "BCC: %s\r\n", bcc)
		}
	}

	for k, v := range m.headers {
		fmt.Fprintf(buf, "%s: %s\r\n", k, v)
	}

	return nil
}

// fromHeader returns a correctly formatted From header, optionally with a name
// component.
func (m *MailYak) fromHeader() string {
	if m.fromName == "" {
		return fmt.Sprintf("From: %s\r\n", m.fromAddr)
	}

	return fmt.Sprintf("From: %s <%s>\r\n", m.fromName, m.fromAddr)
}

// writeBody writes the text/plain and text/html mime parts.
func (m *MailYak) writeBody(w io.Writer, boundary string) error {
	alt := multipart.NewWriter(w)
	defer alt.Close()

	if err := alt.SetBoundary(boundary); err != nil {
		return err
	}

	var err error
	writePart := func(ctype string, data []byte) {
		if len(data) == 0 || err != nil {
			return
		}

		c := fmt.Sprintf("%s; charset=UTF-8", ctype)

		var part io.Writer
		part, err = alt.CreatePart(textproto.MIMEHeader{"Content-Type": {c}, "Content-Transfer-Encoding": {"quoted-printable"}})
		if err != nil {
			return
		}

		var buf bytes.Buffer
		qpw := quotedprintable.NewWriter(&buf)
		_, err = qpw.Write(data)
		qpw.Close()

		_, err = part.Write(buf.Bytes())
	}

	writePart("text/plain", m.plain.Bytes())
	writePart("text/html", m.html.Bytes())

	return err
}
