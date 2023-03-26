package mailyak

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"mime/quotedprintable"
	"net/textproto"
	"strings"
)

func (m *MailYak) buildMime(w io.Writer) error {
	mb, err := randomBoundary()
	if err != nil {
		return err
	}

	ab, err := randomBoundary()
	if err != nil {
		return err
	}

	return m.buildMimeWithBoundaries(w, mb, ab)
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
	return hex.EncodeToString(buf), nil
}

// buildMimeWithBoundaries creates the MIME message using mb and ab as MIME
// boundaries, and returns the generated MIME data as a buffer.
func (m *MailYak) buildMimeWithBoundaries(w io.Writer, mb, ab string) error {
	if err := m.writeHeaders(w); err != nil {
		return err
	}

	// Start our multipart/mixed part
	mixed := multipart.NewWriter(w)
	if err := mixed.SetBoundary(mb); err != nil {
		return err
	}

	// To avoid deferring a mixed.Close(), run the write in a closure and
	// close the mixed after.
	tryWrite := func() error {
		fmt.Fprintf(w, "Content-Type: multipart/mixed;\r\n\tboundary=\"%s\"; charset=UTF-8\r\n\r\n", mixed.Boundary())

		ctype := fmt.Sprintf("multipart/alternative;\r\n\tboundary=\"%s\"", ab)

		altPart, err := mixed.CreatePart(textproto.MIMEHeader{"Content-Type": {ctype}})
		if err != nil {
			return err
		}

		if err := m.writeBody(altPart, ab); err != nil {
			return err
		}

		return m.writeAttachments(mixed, lineSplitterBuilder{})
	}

	if err := tryWrite(); err != nil {
		return err
	}

	return mixed.Close()
}

// writeHeaders writes the Mime-Version, Date, Reply-To, From, To and Subject headers,
// plus any custom headers set via AddHeader().
func (m *MailYak) writeHeaders(w io.Writer) error {

	if _, err := w.Write([]byte(m.fromHeader())); err != nil {
		return err
	}

	if _, err := w.Write([]byte("Mime-Version: 1.0\r\n")); err != nil {
		return err
	}

	fmt.Fprintf(w, "Date: %s\r\n", m.date)

	if m.replyTo != "" {
		fmt.Fprintf(w, "Reply-To: %s\r\n", m.replyTo)
	}

	fmt.Fprintf(w, "Subject: %s\r\n", m.subject)

	if len(m.toAddrs) > 0 {
		commaSeparatedToAddrs := strings.Join(m.toAddrs, ",")
		fmt.Fprintf(w, "To: %s\r\n", commaSeparatedToAddrs)
	}

	if len(m.ccAddrs) > 0 {
		commaSeparatedCCAddrs := strings.Join(m.ccAddrs, ",")
		fmt.Fprintf(w, "CC: %s\r\n", commaSeparatedCCAddrs)
	}

	if m.writeBccHeader && len(m.bccAddrs) > 0 {
		commaSeparatedBCCAddrs := strings.Join(m.bccAddrs, ",")
		fmt.Fprintf(w, "BCC: %s\r\n", commaSeparatedBCCAddrs)
	}

	for k, values := range m.headers {
		for _, v := range values {
			fmt.Fprintf(w, "%s: %s\r\n", k, v)
		}
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
	if m.plain.Len() == 0 && m.html.Len() == 0 {
		// No body to write - just skip it
		return nil
	}

	alt := multipart.NewWriter(w)

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
		_, _ = qpw.Write(data)
		_ = qpw.Close()

		_, err = part.Write(buf.Bytes())
	}

	writePart("text/plain", m.plain.Bytes())
	writePart("text/html", m.html.Bytes())

	// If closing the alt fails, and there's not already an error set, return the
	// close error.
	if closeErr := alt.Close(); err == nil && closeErr != nil {
		return closeErr
	}

	return err
}
