package mailyak

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"regexp"
	"strings"
	"testing"
	"time"
)

// TestMailYakFromHeader ensures the fromHeader method returns valid headers
func TestMailYakFromHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rfromAddr string
		rfromName string
		// Expected results.
		want string
	}{
		{
			"With name",
			"dom@itsallbroken.com",
			"Dom",
			"From: Dom <dom@itsallbroken.com>\r\n",
		},
		{
			"Without name",
			"dom@itsallbroken.com",
			"",
			"From: dom@itsallbroken.com\r\n",
		},
		{
			"Without either",
			"",
			"",
			"From: \r\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := MailYak{
				fromAddr: tt.rfromAddr,
				fromName: tt.rfromName,
			}

			if got := m.fromHeader(); got != tt.want {
				t.Errorf("%q. MailYak.fromHeader() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// TestMailYakWriteHeaders ensures the Mime-Version, Date, Reply-To, From, To and
// Subject headers are correctly wrote
func TestMailYakWriteHeaders(t *testing.T) {
	t.Parallel()

	now := time.Now().Format(time.RFC1123Z)
	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rtoAddrs        []string
		rccAddrs        []string
		rbccAddrs       []string
		rsubject        string
		rreplyTo        string
		rwriteBccHeader bool
		// Expected results.
		wantBuf string
	}{
		{
			"All fields",
			[]string{"test@itsallbroken.com"},
			[]string{},
			[]string{},
			"Test",
			"help@itsallbroken.com",
			true,
			"From: Dom <dom@itsallbroken.com>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nReply-To: help@itsallbroken.com\r\nSubject: Test\r\nTo: test@itsallbroken.com\r\n",
		},
		{
			"No reply-to",
			[]string{"test@itsallbroken.com"},
			[]string{},
			[]string{},
			"",
			"",
			true,
			"From: Dom <dom@itsallbroken.com>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: test@itsallbroken.com\r\n",
		},
		{
			"Multiple To addresses",
			[]string{"test@itsallbroken.com", "repairs@itsallbroken.com"},
			[]string{},
			[]string{},
			"",
			"",
			true,
			"From: Dom <dom@itsallbroken.com>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: test@itsallbroken.com,repairs@itsallbroken.com\r\n",
		},
		{
			"Single Cc address, Multiple To addresses",
			[]string{"test@itsallbroken.com", "repairs@itsallbroken.com"},
			[]string{"cc@itsallbroken.com"},
			[]string{},
			"",
			"",
			true,
			"From: Dom <dom@itsallbroken.com>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: test@itsallbroken.com,repairs@itsallbroken.com\r\nCC: cc@itsallbroken.com\r\n",
		},
		{
			"Multiple Cc addresses, Multiple To addresses",
			[]string{"test@itsallbroken.com", "repairs@itsallbroken.com"},
			[]string{"cc1@itsallbroken.com", "cc2@itsallbroken.com"},
			[]string{},
			"",
			"",
			true,
			"From: Dom <dom@itsallbroken.com>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: test@itsallbroken.com,repairs@itsallbroken.com\r\nCC: cc1@itsallbroken.com,cc2@itsallbroken.com\r\n",
		},
		{
			"Single Bcc address, Multiple To addresses",
			[]string{"test@itsallbroken.com", "repairs@itsallbroken.com"},
			[]string{},
			[]string{"bcc@itsallbroken.com"},
			"",
			"",
			true,
			"From: Dom <dom@itsallbroken.com>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: test@itsallbroken.com,repairs@itsallbroken.com\r\nBCC: bcc@itsallbroken.com\r\n",
		},
		{
			"Multiple Bcc addresses, Multiple To addresses",
			[]string{"test@itsallbroken.com", "repairs@itsallbroken.com"},
			[]string{},
			[]string{"bcc1@itsallbroken.com", "bcc2@itsallbroken.com"},
			"",
			"",
			true,
			"From: Dom <dom@itsallbroken.com>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: test@itsallbroken.com,repairs@itsallbroken.com\r\nBCC: bcc1@itsallbroken.com,bcc2@itsallbroken.com\r\n",
		},
		{
			"Multiple Bcc addresses, Multiple To addresses",
			[]string{"test@itsallbroken.com", "repairs@itsallbroken.com"},
			[]string{},
			[]string{"bcc1@itsallbroken.com", "bcc2@itsallbroken.com"},
			"",
			"",
			false,
			"From: Dom <dom@itsallbroken.com>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: test@itsallbroken.com,repairs@itsallbroken.com\r\n",
		},
		{
			"All together now",
			[]string{"test@itsallbroken.com", "repairs@itsallbroken.com"},
			[]string{"cc1@itsallbroken.com", "cc2@itsallbroken.com"},
			[]string{"bcc1@itsallbroken.com", "bcc2@itsallbroken.com"},
			"",
			"",
			true,
			"From: Dom <dom@itsallbroken.com>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: test@itsallbroken.com,repairs@itsallbroken.com\r\nCC: cc1@itsallbroken.com,cc2@itsallbroken.com\r\nBCC: bcc1@itsallbroken.com,bcc2@itsallbroken.com\r\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := MailYak{
				toAddrs:        tt.rtoAddrs,
				subject:        tt.rsubject,
				fromAddr:       "dom@itsallbroken.com",
				fromName:       "Dom",
				replyTo:        tt.rreplyTo,
				ccAddrs:        tt.rccAddrs,
				bccAddrs:       tt.rbccAddrs,
				writeBccHeader: tt.rwriteBccHeader,
				date:           now,
			}

			buf := &bytes.Buffer{}
			if err := m.writeHeaders(buf); err != nil {
				t.Fatal(err)
			}

			if gotBuf := buf.String(); gotBuf != tt.wantBuf {
				t.Errorf("%q. MailYak.writeHeaders() = %v, want %v", tt.name, gotBuf, tt.wantBuf)
			}
		})
	}
}

// TestMailYakWriteBody ensures the correct MIME parts are wrote for the body
func TestMailYakWriteBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rHTML  string
		rPlain string
		// Parameters.
		boundary string
		// Expected results.
		wantW   string
		wantErr bool
	}{
		{
			"Empty",
			"",
			"",
			"test",
			"",
			false,
		},
		{
			"HTML",
			"HTML",
			"",
			"t",
			"--t\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/html; charset=UTF-8\r\n\r\nHTML\r\n--t--\r\n",
			false,
		},
		{
			"Plain text",
			"",
			"Plain",
			"t",
			"--t\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nPlain\r\n--t--\r\n",
			false,
		},
		{
			"Both",
			"HTML",
			"Plain",
			"t",
			"--t\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nPlain\r\n--t\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/html; charset=UTF-8\r\n\r\nHTML\r\n--t--\r\n",
			false,
		},
		{
			"Both with long lines",
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
			"t",
			"--t\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nLorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tem=\r\npor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, q=\r\nuis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo cons=\r\nequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillu=\r\nm dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non pr=\r\noident, sunt in culpa qui officia deserunt mollit anim id est laborum.\r\n--t\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/html; charset=UTF-8\r\n\r\nLorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tem=\r\npor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, q=\r\nuis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo cons=\r\nequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillu=\r\nm dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non pr=\r\noident, sunt in culpa qui officia deserunt mollit anim id est laborum.\r\n--t--\r\n",
			false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := MailYak{}
			m.HTML().WriteString(tt.rHTML)
			m.Plain().WriteString(tt.rPlain)

			w := &bytes.Buffer{}
			if err := m.writeBody(w, tt.boundary); (err != nil) != tt.wantErr {
				t.Fatalf("%q. MailYak.writeBody() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("%q. MailYak.writeBody() = %v, want %v", tt.name, gotW, tt.wantW)
			}
		})
	}
}

// TestMailYakBuildMime tests all the other mime-related bits combine in a sane way
func TestMailYakBuildMime(t *testing.T) {
	t.Parallel()

	now := time.Now().Format(time.RFC1123Z)
	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rHTML     []byte
		rPlain    []byte
		rtoAddrs  []string
		rsubject  string
		rfromAddr string
		rfromName string
		rreplyTo  string
		// Expected results.
		want    string
		wantErr bool
	}{
		{
			"Empty",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"",
			"",
			"",
			"From: \r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: \r\nContent-Type: multipart/mixed;\r\n\tboundary=\"mixed\"; charset=UTF-8\r\n\r\n--mixed\r\nContent-Type: multipart/alternative;\r\n\tboundary=\"alt\"\r\n\r\n\r\n--mixed--\r\n",
			false,
		},
		{
			"HTML",
			[]byte("HTML"),
			[]byte{},
			[]string{""},
			"",
			"",
			"",
			"",
			"From: \r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: \r\nContent-Type: multipart/mixed;\r\n\tboundary=\"mixed\"; charset=UTF-8\r\n\r\n--mixed\r\nContent-Type: multipart/alternative;\r\n\tboundary=\"alt\"\r\n\r\n--alt\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/html; charset=UTF-8\r\n\r\nHTML\r\n--alt--\r\n\r\n--mixed--\r\n",
			false,
		},
		{
			"Plain",
			[]byte{},
			[]byte("Plain"),
			[]string{""},
			"",
			"",
			"",
			"",
			"From: \r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: \r\nContent-Type: multipart/mixed;\r\n\tboundary=\"mixed\"; charset=UTF-8\r\n\r\n--mixed\r\nContent-Type: multipart/alternative;\r\n\tboundary=\"alt\"\r\n\r\n--alt\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nPlain\r\n--alt--\r\n\r\n--mixed--\r\n",
			false,
		},
		{
			"Reply-To",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"",
			"",
			"reply",
			"From: \r\nMime-Version: 1.0\r\nDate: " + now + "\r\nReply-To: reply\r\nSubject: \r\nTo: \r\nContent-Type: multipart/mixed;\r\n\tboundary=\"mixed\"; charset=UTF-8\r\n\r\n--mixed\r\nContent-Type: multipart/alternative;\r\n\tboundary=\"alt\"\r\n\r\n\r\n--mixed--\r\n",
			false,
		},
		{
			"From name",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"",
			"name",
			"",
			"From: name <>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: \r\nContent-Type: multipart/mixed;\r\n\tboundary=\"mixed\"; charset=UTF-8\r\n\r\n--mixed\r\nContent-Type: multipart/alternative;\r\n\tboundary=\"alt\"\r\n\r\n\r\n--mixed--\r\n",
			false,
		},
		{
			"From name + address",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"addr",
			"name",
			"",
			"From: name <addr>\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: \r\nContent-Type: multipart/mixed;\r\n\tboundary=\"mixed\"; charset=UTF-8\r\n\r\n--mixed\r\nContent-Type: multipart/alternative;\r\n\tboundary=\"alt\"\r\n\r\n\r\n--mixed--\r\n",
			false,
		},
		{
			"From",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"from",
			"",
			"",
			"From: from\r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: \r\nContent-Type: multipart/mixed;\r\n\tboundary=\"mixed\"; charset=UTF-8\r\n\r\n--mixed\r\nContent-Type: multipart/alternative;\r\n\tboundary=\"alt\"\r\n\r\n\r\n--mixed--\r\n",
			false,
		},
		{
			"Subject",
			[]byte{},
			[]byte{},
			[]string{""},
			"subject",
			"",
			"",
			"",
			"From: \r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: subject\r\nTo: \r\nContent-Type: multipart/mixed;\r\n\tboundary=\"mixed\"; charset=UTF-8\r\n\r\n--mixed\r\nContent-Type: multipart/alternative;\r\n\tboundary=\"alt\"\r\n\r\n\r\n--mixed--\r\n",
			false,
		},
		{
			"To addresses",
			[]byte{},
			[]byte{},
			[]string{"one", "two"},
			"",
			"",
			"",
			"",
			"From: \r\nMime-Version: 1.0\r\nDate: " + now + "\r\nSubject: \r\nTo: one,two\r\nContent-Type: multipart/mixed;\r\n\tboundary=\"mixed\"; charset=UTF-8\r\n\r\n--mixed\r\nContent-Type: multipart/alternative;\r\n\tboundary=\"alt\"\r\n\r\n\r\n--mixed--\r\n",
			false,
		},
	}

	regex := regexp.MustCompile("\r?\n")

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := &MailYak{
				toAddrs:   tt.rtoAddrs,
				subject:   tt.rsubject,
				fromAddr:  tt.rfromAddr,
				fromName:  tt.rfromName,
				replyTo:   tt.rreplyTo,
				trimRegex: regex,
				date:      now,
			}
			m.HTML().Write(tt.rHTML)
			m.Plain().Write(tt.rPlain)

			buf := &bytes.Buffer{}
			err := m.buildMimeWithBoundaries(buf, "mixed", "alt")
			if (err != nil) != tt.wantErr {
				t.Fatalf("%q. MailYak.buildMime() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			if buf.String() != tt.want {
				t.Errorf("%q. MailYak.buildMime() = %v, want %v", tt.name, buf.String(), tt.want)
			}
		})
	}
}

// TestMailYakBuildMime_withAttachments ensures attachments are correctly added to the MIME message
func TestMailYakBuildMime_withAttachments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rHTML        []byte
		rPlain       []byte
		rtoAddrs     []string
		rsubject     string
		rfromAddr    string
		rfromName    string
		rreplyTo     string
		rattachments []attachment
		// Expected results.
		wantAttach []string
		wantErr    bool
	}{
		{
			"No attachment",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"",
			"",
			"",
			[]attachment{},
			[]string{},
			false,
		},
		{
			"One attachment",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"",
			"",
			"",
			[]attachment{
				{"test.txt", strings.NewReader("content"), false, ""},
			},
			[]string{"Y29udGVudA=="},
			false,
		},
		{
			"One inline attachment",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"",
			"",
			"",
			[]attachment{
				{"test.txt", strings.NewReader("content"), true, ""},
			},
			[]string{"Y29udGVudA=="},
			false,
		},
		{
			"Two attachments",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"",
			"",
			"",
			[]attachment{
				{"test.txt", strings.NewReader("content"), false, ""},
				{"another.txt", strings.NewReader("another"), false, ""},
			},
			[]string{"Y29udGVudA==", "YW5vdGhlcg=="},
			false,
		},
		{
			"Two inline attachments",
			[]byte{},
			[]byte{},
			[]string{""},
			"",
			"",
			"",
			"",
			[]attachment{
				{"test.txt", strings.NewReader("content"), true, ""},
				{"another.txt", strings.NewReader("another"), true, ""},
			},
			[]string{"Y29udGVudA==", "YW5vdGhlcg=="},
			false,
		},
	}

	regex := regexp.MustCompile("\r?\n")

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := &MailYak{
				toAddrs:     tt.rtoAddrs,
				subject:     tt.rsubject,
				fromAddr:    tt.rfromAddr,
				fromName:    tt.rfromName,
				replyTo:     tt.rreplyTo,
				attachments: tt.rattachments,
				trimRegex:   regex,
			}
			m.HTML().Write(tt.rHTML)
			m.Plain().Write(tt.rPlain)

			buf := &bytes.Buffer{}
			err := m.buildMimeWithBoundaries(buf, "mixed", "alt")
			if (err != nil) != tt.wantErr {
				t.Fatalf("%q. MailYak.buildMime() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			seen := 0
			mr := multipart.NewReader(buf, "mixed")

			// Itterate over the mime parts, look for attachments
			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Errorf("%q. MailYak.buildMime() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				}

				// Read the attachment data
				slurp, err := ioutil.ReadAll(p)
				if err != nil {
					t.Errorf("%q. MailYak.buildMime() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				}

				// Skip non-attachments
				if p.Header.Get("Content-Disposition") == "" {
					continue
				}

				// Run through our attachments looking for a match
				for i, attch := range tt.rattachments {
					// Check Disposition header
					var disp string
					if attch.inline {
						disp = "inline; filename=%q; name=%q"
					} else {
						disp = "attachment; filename=%q; name=%q"
					}
					if p.Header.Get("Content-Disposition") != fmt.Sprintf(disp, attch.filename, attch.filename) {
						continue
					}

					// Check data
					if !bytes.Equal(slurp, []byte(tt.wantAttach[i])) {
						fmt.Printf("Part %q: %q\n", p.Header.Get("Content-Disposition"), slurp)
						continue
					}

					seen++
				}

			}

			// Did we see all the expected attachments?
			if seen != len(tt.rattachments) {
				t.Errorf("%q. MailYak.buildMime() didn't find all attachments in mime body", tt.name)
			}
		})
	}
}
