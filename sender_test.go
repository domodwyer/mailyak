package mailyak

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

// connAsserts wraps a net.Conn, performing writes and asserts responses within
// tests.
//
// Any errors cause the method to panic.
type connAsserts struct {
	conn net.Conn

	buf *bytes.Buffer
}

func (c *connAsserts) Write(put string) {
	n, err := c.conn.Write([]byte(put))
	if err != nil {
		panic(fmt.Sprintf("got error %v writing %q (wrote %d bytes)", err, put, n))
	}
}

func (c *connAsserts) Expect(want string) {
	c.buf.Reset()

	n, err := io.CopyN(c.buf, c.conn, int64(len(want)))
	if err != nil {
		panic(fmt.Sprintf("got error %v after reading %d bytes (got %q, want %q)", err, n, c.buf.String(), want))
	}
	if c.buf.String() != want {
		panic(fmt.Sprintf("read %q, want %q", c.buf.String(), want))
	}
}

func newConnAsserts(c net.Conn) *connAsserts {
	return &connAsserts{
		conn: c,
		buf:  &bytes.Buffer{},
	}
}
