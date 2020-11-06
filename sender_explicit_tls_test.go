package mailyak

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"reflect"
	"testing"
	"time"
)

var (
	testBody = []byte("bananas")

	// Test RSA key & self-signed certificate populated by init()
	testRSAKey    *rsa.PrivateKey
	testCertBytes []byte
	testCert      *x509.Certificate
)

// Initialise the TLS certificate and key material
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

// newTLSServer starts a TLS server on a random port assigned by the kernel with
// a self-signed certificate, returning the bound socket and the TLS
// configuration populated with the root certificates required to connect to it.
func newTLSServer(t *testing.T, actor func(c *connAsserts)) (net.Listener, *tls.Config) {
	serverConfig := &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{testCertBytes},
				PrivateKey:  testRSAKey,
			},
		},
	}

	// Bind a TLS-enabled TCP socket to some random port
	l, err := tls.Listen("tcp", "127.0.0.1:0", serverConfig)
	if err != nil {
		t.Fatalf("failed to bind to localhost: %v", err)
	}

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			actor(newConnAsserts(conn))
			conn.Close()
		}
	}()

	roots := x509.NewCertPool()
	roots.AddCert(testCert)

	return l, &tls.Config{
		RootCAs: roots,
	}
}

func TestSenderExplicitTLS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string

		// Called with the MailYak instace before sending to set up the email
		// contents
		mailFn func(m *MailYak)

		// Called once the Send() method is invoked, impersonating and asserting
		// the client/server conversation.
		connFn func(c *connAsserts)

		wantErr error
	}{
		{
			name: "2 addresses in To",
			mailFn: func(m *MailYak) {
				m.From("dom@itsallbroken.com")
				m.To("to@example.org", "another@example.com")
			},
			connFn: func(c *connAsserts) {
				c.Write("220 smtp.itsallbroken.com ESMTP bananas\r\n")

				c.Expect("EHLO localhost\r\n")
				c.Write("250 smtp.itsallbroken.com Hola\r\n")

				c.Expect("MAIL FROM:<dom@itsallbroken.com>\r\n")
				c.Write("250 OK\r\n")

				c.Expect("RCPT TO:<to@example.org>\r\n")
				c.Write("250 OK\r\n")

				c.Expect("RCPT TO:<another@example.com>\r\n")
				c.Write("250 OK\r\n")

				c.Expect("DATA\r\n")
				c.Write("354 OK\r\n")
				c.Expect("bananas\r\n.\r\n")
				c.Write("250 Will do friend\r\n")

				c.Expect("QUIT\r\n")
				c.Write("221 Adios\r\n")
			},
			wantErr: nil,
		},
		{
			name: "with CC",
			mailFn: func(m *MailYak) {
				m.From("from@example.org")
				m.To("to@example.org", "another@example.com")
				m.Cc("dom@itsallbroken.com")
			},
			connFn: func(c *connAsserts) {
				c.Write("220 smtp.itsallbroken.com ESMTP bananas\r\n")

				c.Expect("EHLO localhost\r\n")
				c.Write("250 smtp.itsallbroken.com Hola\r\n")

				c.Expect("MAIL FROM:<from@example.org>\r\n")
				c.Write("250 OK\r\n")

				c.Expect("RCPT TO:<to@example.org>\r\n")
				c.Write("250 OK\r\n")

				c.Expect("RCPT TO:<another@example.com>\r\n")
				c.Write("250 OK\r\n")

				c.Expect("RCPT TO:<dom@itsallbroken.com>\r\n")
				c.Write("250 OK\r\n")

				c.Expect("DATA\r\n")
				c.Write("354 OK\r\n")
				c.Expect("bananas\r\n.\r\n")
				c.Write("250 Will do friend\r\n")

				c.Expect("QUIT\r\n")
				c.Write("221 Adios\r\n")
			},
			wantErr: nil,
		},
		{
			name: "with BCC",
			mailFn: func(m *MailYak) {
				m.From("from@example.org")
				m.To("to@example.org", "another@example.com")
				m.Bcc("dom@itsallbroken.com")
			},
			connFn: func(c *connAsserts) {
				c.Write("220 smtp.itsallbroken.com ESMTP bananas\r\n")

				c.Expect("EHLO localhost\r\n")
				c.Write("250 smtp.itsallbroken.com Hola\r\n")

				c.Expect("MAIL FROM:<from@example.org>\r\n")
				c.Write("250 OK\r\n")

				c.Expect("RCPT TO:<to@example.org>\r\n")
				c.Write("250 OK\r\n")

				c.Expect("RCPT TO:<another@example.com>\r\n")
				c.Write("250 OK\r\n")

				c.Expect("RCPT TO:<dom@itsallbroken.com>\r\n")
				c.Write("250 OK\r\n")

				c.Expect("DATA\r\n")
				c.Write("354 OK\r\n")
				c.Expect("bananas\r\n.\r\n")
				c.Write("250 Will do friend\r\n")

				c.Expect("QUIT\r\n")
				c.Write("221 Adios\r\n")
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		var tt = tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Start a test server
			socket, clientTLSConfig := newTLSServer(t, tt.connFn)
			defer socket.Close()

			m, err := NewWithTLS(socket.Addr().String(), nil, clientTLSConfig)
			if err != nil {
				t.Fatal(err)
			}

			// Populate the email
			tt.mailFn(m)

			// Override the call the sender directly for deterministic results
			// (avoiding the dynamic date / mime boundaries)
			err = m.sender.Send(m, testBody)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}
