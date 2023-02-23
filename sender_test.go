package mailyak

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/smtp"
	"reflect"
	"testing"
	"time"
)

var (
	// Test RSA key & self-signed certificate populated by init()
	testRSAKey    *rsa.PrivateKey
	testCertBytes []byte
	testCert      *x509.Certificate
)

// Initialise the TLS certificate and key material for TLS tests.
func init() {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(fmt.Sprintf("failed to generate RSA test key: %v", err))
	}
	testRSAKey = key

	// Define the certificate template
	self := &x509.Certificate{
		Version:      3,
		SerialNumber: big.NewInt(42),
		Issuer: pkix.Name{
			CommonName: "CA Bananas Inc",
		},
		Subject: pkix.Name{
			CommonName: "The Banana Factory",
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(time.Hour),
		IPAddresses: []net.IP{
			net.IPv4(127, 0, 0, 1),
		},
	}

	// Sign the template certificate
	cert, err := x509.CreateCertificate(rand.Reader, self, self, &testRSAKey.PublicKey, testRSAKey)
	if err != nil {
		panic(fmt.Sprintf("failed to generate self-signed test cert: %v", err))
	}
	testCertBytes = cert

	// Parse the signed certificate
	serverCert, err := x509.ParseCertificate(testCertBytes)
	if err != nil {
		panic(fmt.Sprintf("failed to bind to localhost: %v", err))
	}
	testCert = serverCert
}

// connAsserts wraps a net.Conn, performing writes and asserts responses within
// tests.
//
// Any errors cause the method to panic.
type connAsserts struct {
	net.Conn

	t   *testing.T
	buf *bytes.Buffer
}

func (c *connAsserts) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	c.t.Logf("MailYak -> Server:\n%s\n", hex.Dump(b))
	return n, err
}

func (c *connAsserts) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	c.t.Logf("Server -> MailYak:\n%s\n", hex.Dump(b))
	return n, err
}

func (c *connAsserts) Expect(want string) {
	c.buf.Reset()

	n, err := io.CopyN(c.buf, c, int64(len(want)))
	if err != nil {
		c.t.Fatalf("got error %v after reading %d bytes (got %q, want %q)", err, n, c.buf.String(), want)
	}
	if c.buf.String() != want {
		c.t.Fatalf("read %q, want %q", c.buf.String(), want)
	}
}

func (c *connAsserts) Respond(put string) {
	n, err := c.Write([]byte(put))
	if err != nil {
		c.t.Fatalf("got error %v writing %q (wrote %d bytes)", err, put, n)
	}
}

func newConnAsserts(c net.Conn, t *testing.T) *connAsserts {
	return &connAsserts{
		Conn: c,
		t:    t,
		buf:  &bytes.Buffer{},
	}
}

// mockMail provides the methods for a sendableMail, allowing for deterministic
// MIME content in tests.
type mockMail struct {
	localName string
	toAddrs   []string
	fromAddr  string
	auth      smtp.Auth
	mime      string
}

// getLocalName should return the sender domain to be used in the EHLO/HELO
// command.
func (m *mockMail) getLocalName() string {
	return m.localName
}

// toAddrs should return a slice of email addresses to be added to the RCPT
// TO command.
func (m *mockMail) getToAddrs() []string {
	return stripNames(m.toAddrs)
}

// fromAddr should return the address to be used in the MAIL FROM command.
func (m *mockMail) getFromAddr() string {
	return m.fromAddr
}

// auth should return the smtp.Auth if configured, nil if not.
func (m *mockMail) getAuth() smtp.Auth {
	return m.auth
}

// buildMime should write the generated MIME to w.
//
// The emailSender implementation is responsible for providing appropriate
// buffering of writes.
func (m *mockMail) buildMime(w io.Writer) error {
	_, err := w.Write([]byte(m.mime))
	return err
}

// TestSMTPProtocolExchange sends the same mock email over two different
// transports using two different sender implementations, ensuring parity
// between the two (specifically that both impleementations result in the same
// SMTP conversation).
//
// Because the mock server in the tests does not advertise STARTTLS support in,
// there is no upgrade.
func TestSMTPProtocolExchange(t *testing.T) {
	t.Parallel()

	const testTimeout = 15 * time.Second

	tests := []struct {
		name string
		mail *mockMail

		// Called once the Send() method is invoked, impersonating and asserting
		// the client/server conversation.
		connFn func(c *connAsserts)

		// Error returned when sending over TLS
		wantTLSErr error

		// Error returned when sending over plaintext
		wantPlaintextErr error
	}{
		{
			name: "ok",
			mail: &mockMail{
				toAddrs: []string{
					"to@example.org",
					"another@example.com",
					"Dom <dom@itsallbroken.com>",
				},
				fromAddr: "from@example.org",
				mime:     "bananas",
			},
			connFn: func(c *connAsserts) {
				c.Respond("220 localhost ESMTP bananas\r\n")

				c.Expect("EHLO localhost\r\n")
				c.Respond("250-localhost Hola\r\n")
				c.Respond("250 AUTH LOGIN PLAIN\r\n")

				c.Expect("MAIL FROM:<from@example.org>\r\n")
				c.Respond("250 OK\r\n")

				c.Expect("RCPT TO:<to@example.org>\r\n")
				c.Respond("250 OK\r\n")

				c.Expect("RCPT TO:<another@example.com>\r\n")
				c.Respond("250 OK\r\n")

				c.Expect("RCPT TO:<dom@itsallbroken.com>\r\n")
				c.Respond("250 OK\r\n")

				c.Expect("DATA\r\n")
				c.Respond("354 OK\r\n")
				c.Expect("bananas\r\n.\r\n")
				c.Respond("250 Will do friend\r\n")

				c.Expect("QUIT\r\n")
				c.Respond("221 Adios\r\n")
			},
			wantTLSErr:       nil,
			wantPlaintextErr: nil,
		},
		{
			name: "with auth",
			mail: &mockMail{
				toAddrs: []string{
					"to@example.org",
					"another@example.com",
					"dom@itsallbroken.com",
				},
				fromAddr: "from@example.org",
				mime:     "bananas",
				auth:     smtp.PlainAuth("ident", "user", "pass", "127.0.0.1"),
			},
			connFn: func(c *connAsserts) {
				c.Respond("220 localhost ESMTP bananas\r\n")

				c.Expect("EHLO localhost\r\n")
				c.Respond("250-localhost Hola\r\n")
				c.Respond("250 AUTH LOGIN PLAIN\r\n")

				c.Expect("AUTH PLAIN aWRlbnQAdXNlcgBwYXNz\r\n")
				c.Respond("235 Looks good\r\n")

				c.Expect("MAIL FROM:<from@example.org>\r\n")
				c.Respond("250 OK\r\n")

				c.Expect("RCPT TO:<to@example.org>\r\n")
				c.Respond("250 OK\r\n")

				c.Expect("RCPT TO:<another@example.com>\r\n")
				c.Respond("250 OK\r\n")

				c.Expect("RCPT TO:<dom@itsallbroken.com>\r\n")
				c.Respond("250 OK\r\n")

				c.Expect("DATA\r\n")
				c.Respond("354 OK\r\n")
				c.Expect("bananas\r\n.\r\n")
				c.Respond("250 Will do friend\r\n")

				c.Expect("QUIT\r\n")
				c.Respond("221 Adios\r\n")
			},
			wantTLSErr:       nil,
			wantPlaintextErr: nil,
		},
		{
			name: "with localname",
			mail: &mockMail{
				toAddrs: []string{
					"to@example.org",
					"another@example.com",
					"Dom <dom@itsallbroken.com>",
				},
				fromAddr:  "from@example.org",
				mime:      "bananas",
				localName: "example.com",
			},
			connFn: func(c *connAsserts) {
				c.Respond("220 localhost ESMTP bananas\r\n")

				c.Expect("EHLO example.com\r\n")
				c.Respond("250-example.com Hola\r\n")
				c.Respond("250 AUTH LOGIN PLAIN\r\n")

				c.Expect("MAIL FROM:<from@example.org>\r\n")
				c.Respond("250 OK\r\n")

				c.Expect("RCPT TO:<to@example.org>\r\n")
				c.Respond("250 OK\r\n")

				c.Expect("RCPT TO:<another@example.com>\r\n")
				c.Respond("250 OK\r\n")

				c.Expect("RCPT TO:<dom@itsallbroken.com>\r\n")
				c.Respond("250 OK\r\n")

				c.Expect("DATA\r\n")
				c.Respond("354 OK\r\n")
				c.Expect("bananas\r\n.\r\n")
				c.Respond("250 Will do friend\r\n")

				c.Expect("QUIT\r\n")
				c.Respond("221 Adios\r\n")
			},
			wantTLSErr:       nil,
			wantPlaintextErr: nil,
		},
	}

	// handleConn provides the accept loop for both the TLS server, and the
	// plain-text server, passing the accepted connection to the test actor
	// func.
	//
	// Once the actor func has finished, done is closed.
	handleConn := func(t *testing.T, l net.Listener, done chan<- struct{}, actor func(c *connAsserts)) {
		defer close(done)

		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		actor(newConnAsserts(conn, t))
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Send the mock email over each implementation of the sender
			// interface, including initialisation with the respective MailYak
			// constructor.

			t.Run("Explicit_TLS", func(t *testing.T) {
				t.Parallel()

				ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
				defer cancel()

				// Initialise a server TLS config using the self-signed test
				// certificate and key material.
				serverConfig := &tls.Config{
					Certificates: []tls.Certificate{
						{
							Certificate: [][]byte{testCertBytes},
							PrivateKey:  testRSAKey,
						},
					},
				}

				// Bind a TLS-enabled TCP socket to some random port
				socket, err := tls.Listen("tcp", "127.0.0.1:0", serverConfig)
				if err != nil {
					t.Fatalf("failed to bind to localhost: %v", err)
				}
				defer socket.Close()

				handlerDone := make(chan struct{})
				go handleConn(t, socket, handlerDone, tt.connFn)

				// Build a root store for the self-signed certificate.
				roots := x509.NewCertPool()
				roots.AddCert(testCert)

				// Initialise a TLS mailyak using the root store.
				m, err := NewWithTLS(socket.Addr().String(), nil, &tls.Config{
					RootCAs:    roots,
					ServerName: "127.0.0.1",
				})
				if err != nil {
					t.Fatal(err)
				}

				// Call into the sender directly, giving it the mock
				// sendableEmail
				sendErr := make(chan error)
				go func() {
					sendErr <- m.sender.Send(tt.mail)
				}()

				// Wait for the SMTP conversation to complete
				select {
				case <-ctx.Done():
					t.Fatal("timeout waiting for SMTP conversation to complete")
				case <-handlerDone:
					// The handler is complete, wait for the send error and
					// check it matches the expected value.
					select {
					case <-ctx.Done():
						t.Fatal("timeout waiting for Send() to return")

					case err := <-sendErr:
						if !reflect.DeepEqual(err, tt.wantTLSErr) {
							t.Errorf("got %v, want %v", err, tt.wantTLSErr)
						}
					}
				}
			})

			t.Run("Plaintext", func(t *testing.T) {
				t.Parallel()

				ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
				defer cancel()

				// Start listening to a local plain-text socket
				socket, err := net.Listen("tcp", "127.0.0.1:0")
				if err != nil {
					t.Fatalf("failed to bind to localhost: %v", err)
				}
				defer socket.Close()

				handlerDone := make(chan struct{})
				go handleConn(t, socket, handlerDone, tt.connFn)

				m := New(socket.Addr().String(), nil)

				// Call into the sender directly, giving it the mock
				// sendableEmail
				sendErr := make(chan error)
				go func() {
					sendErr <- m.sender.Send(tt.mail)
				}()

				// Wait for the SMTP conversation to complete
				select {
				case <-ctx.Done():
					t.Fatal("timeout waiting for SMTP conversation to complete")
				case <-handlerDone:
					// The handler is complete, wait for the send error and
					// check it matches the expected value.
					select {
					case <-ctx.Done():
						t.Fatal("timeout waiting for Send() to return")

					case err := <-sendErr:
						if !reflect.DeepEqual(err, tt.wantPlaintextErr) {
							t.Errorf("got %v, want %v", err, tt.wantPlaintextErr)
						}
					}
				}
			})
		})
	}
}
