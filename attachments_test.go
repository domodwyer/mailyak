package mailyak

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/textproto"
	"strings"
	"testing"
)

type testAttachment struct {
	contentType string
	disposition string
	data        bytes.Buffer
}

// Satisfy the partCreator interface and keep track of wrote attachments
type testPartCreator struct {
	attachments []*testAttachment
}

func (t *testPartCreator) CreatePart(header textproto.MIMEHeader) (io.Writer, error) {
	a := &testAttachment{
		contentType: header.Get("Content-Type"),
		disposition: header.Get("Content-Disposition"),
	}

	t.attachments = append(t.attachments, a)
	return &a.data, nil
}

// nopSplitter - it does nothing!
type nopSplitter struct {
	w io.Writer
}

func (t nopSplitter) Write(p []byte) (int, error) {
	return t.w.Write(p)
}

// nopBuilder satisfies the writeWrapper interface
type nopBuilder struct{}

func (b nopBuilder) new(w io.Writer) io.Writer {
	return &nopSplitter{w: w}
}

// TestMailYakAttach calls Attach() and ensures the attachment slice is the
// correct length
func TestMailYakAttach(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rattachments []attachment
		// Parameters.
		pname string
		r     io.Reader
		// Expect
		count int
	}{
		{
			"From empty",
			[]attachment{},
			"test",
			&bytes.Buffer{},
			1,
		},
		{
			"From one",
			[]attachment{{"Existing", &bytes.Buffer{}, false, ""}},
			"test",
			&bytes.Buffer{},
			2,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := &MailYak{attachments: tt.rattachments}
			m.Attach(tt.pname, tt.r)

			if tt.count != len(m.attachments) {
				t.Errorf("%q. MailYak.Attach() len = %v, wantLen %v", tt.name, len(m.attachments), tt.count)
			}
		})
	}
}

// TestMailYakAttach calls AttachInline() and ensures the attachment slice is the
// correct length
func TestMailYakAttachInline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rattachments []attachment
		// Parameters.
		pname string
		r     io.Reader
		// Expect
		count int
	}{
		{
			"From empty",
			[]attachment{},
			"test",
			&bytes.Buffer{},
			1,
		},
		{
			"From one",
			[]attachment{{"Existing", &bytes.Buffer{}, false, ""}},
			"test",
			&bytes.Buffer{},
			2,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := &MailYak{attachments: tt.rattachments}
			m.AttachInline(tt.pname, tt.r)

			if tt.count != len(m.attachments) {
				t.Errorf("%q. MailYak.Attach() len = %v, wantLen %v", tt.name, len(m.attachments), tt.count)
			}
		})
	}
}

// TestMailYakAttachWithMimeType calls AttachWithMimeType() and ensures the attachment slice is the
// correct length
func TestMailYakAttachWithMimeType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rattachments []attachment
		// Parameters.
		pname string
		r     io.Reader
		mime  string
		// Expect
		count int
	}{
		{
			"From empty",
			[]attachment{},
			"test",
			&bytes.Buffer{},
			"text/csv; charset=utf-8",
			1,
		},
		{
			"From one",
			[]attachment{{"Existing", &bytes.Buffer{}, false, "text/csv; charset=utf-8"}},
			"test",
			&bytes.Buffer{},
			"text/csv; charset=utf-8",
			2,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := &MailYak{attachments: tt.rattachments}
			m.AttachWithMimeType(tt.pname, tt.r, tt.mime)

			if tt.count != len(m.attachments) {
				t.Errorf("%q. MailYak.Attach() len = %v, wantLen %v", tt.name, len(m.attachments), tt.count)
			}
		})
	}
}

// TestMailYakAttachWithMimeType calls AttachInlineWithMimeType() and ensures the attachment slice is the
// correct length
func TestMailYakAttachInlineWithMimeType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rattachments []attachment
		// Parameters.
		pname string
		r     io.Reader
		mime  string
		// Expect
		count int
	}{
		{
			"From empty",
			[]attachment{},
			"test",
			&bytes.Buffer{},
			"text/csv; charset=utf-8",
			1,
		},
		{
			"From one",
			[]attachment{{"Existing", &bytes.Buffer{}, false, "text/csv; charset=utf-8"}},
			"test",
			&bytes.Buffer{},
			"text/csv; charset=utf-8",
			2,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := &MailYak{attachments: tt.rattachments}
			m.AttachInlineWithMimeType(tt.pname, tt.r, tt.mime)

			if tt.count != len(m.attachments) {
				t.Errorf("%q. MailYak.Attach() len = %v, wantLen %v", tt.name, len(m.attachments), tt.count)
			}
		})
	}
}

// TestMailYakWriteAttachments ensures the correct headers are wrote, and the
// data is base64 encoded correctly
func TestMailYakWriteAttachments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rattachments []attachment
		// Expected results.
		ctype   string
		disp    string
		data    string
		wantErr bool
	}{
		{
			"Empty",
			[]attachment{{"Empty", &bytes.Buffer{}, false, ""}},
			"text/plain; charset=utf-8;\n\tfilename=\"Empty\"; name=\"Empty\"",
			"attachment;\n\tfilename=\"Empty\"; name=\"Empty\"",
			"",
			false,
		},
		{
			"Short string",
			[]attachment{{"advice", strings.NewReader("Don't Panic"), false, ""}},
			"text/plain; charset=utf-8;\n\tfilename=\"advice\"; name=\"advice\"",
			"attachment;\n\tfilename=\"advice\"; name=\"advice\"",
			"RG9uJ3QgUGFuaWM=",
			false,
		},
		{
			"Space in filename",
			[]attachment{{"Empty with spaces", &bytes.Buffer{}, false, ""}},
			"text/plain; charset=utf-8;\n\tfilename=\"Empty with spaces\"; name=\"Empty with spaces\"",
			"attachment;\n\tfilename=\"Empty with spaces\"; name=\"Empty with spaces\"",
			"",
			false,
		},
		{
			"With specified MIME type",
			[]attachment{{"Empty with spaces", &bytes.Buffer{}, false, "text/csv; charset=utf-8"}},
			"text/csv; charset=utf-8;\n\tfilename=\"Empty with spaces\"; name=\"Empty with spaces\"",
			"attachment;\n\tfilename=\"Empty with spaces\"; name=\"Empty with spaces\"",
			"",
			false,
		},
		{
			"Longer string",
			[]attachment{
				{
					"partyinvite.txt",
					strings.NewReader(
						"If Baldrick served a meal at HQ he would be arrested for the biggest " +
							"mass poisoning since Lucretia Borgia invited 500 friends for a Wine and Anthrax Party.",
					),
					false,
					"",
				},
			},
			"text/plain; charset=utf-8;\n\tfilename=\"partyinvite.txt\"; name=\"partyinvite.txt\"",
			"attachment;\n\tfilename=\"partyinvite.txt\"; name=\"partyinvite.txt\"",
			"SWYgQmFsZHJpY2sgc2VydmVkIGEgbWVhbCBhdCBIUSBoZSB3b3VsZCBiZSBhcnJlc3Rl" +
				"ZCBmb3IgdGhlIGJpZ2dlc3QgbWFzcyBwb2lzb25pbmcgc2luY2UgTHVjcmV0aWEgQm9y" +
				"Z2lhIGludml0ZWQgNTAwIGZyaWVuZHMgZm9yIGEgV2luZSBhbmQgQW50aHJheCBQYXJ0eS4=",
			false,
		},
		{
			"String >512 characters (content type sniff)",
			[]attachment{
				{
					"qed.txt",
					strings.NewReader(
						`Now it is such a bizarrely improbable coincidence that anything so mind-bogglingly ` +
							`useful could have evolved purely by chance that some thinkers have chosen to see it ` +
							`as the final and clinching proof of the non-existence of God. The argument goes something ` +
							`like this: "I refuse to prove that I exist," says God, "for proof denies faith, and ` +
							`without faith I am nothing." "But," says Man, "The Babel fish is a dead giveaway, ` +
							`isn't it? It could not have evolved by chance. It proves you exist, and so therefore, ` +
							`by your own arguments, you don't. QED." "Oh dear," says God, "I hadn't thought of ` +
							`that," and promptly vanishes in a puff of logic. "Oh, that was easy," says Man, and ` +
							`for an encore goes on to prove that black is white and gets himself killed on the next ` +
							`zebra crossing.`,
					),
					false,
					"",
				},
			},
			"text/plain; charset=utf-8;\n\tfilename=\"qed.txt\"; name=\"qed.txt\"",
			"attachment;\n\tfilename=\"qed.txt\"; name=\"qed.txt\"",
			"Tm93IGl0IGlzIHN1Y2ggYSBiaXphcnJlbHkgaW1wcm9iYWJsZSBjb2luY2lkZW5jZSB0a" +
				"GF0IGFueXRoaW5nIHNvIG1pbmQtYm9nZ2xpbmdseSB1c2VmdWwgY291bGQgaGF2ZSBldm" +
				"9sdmVkIHB1cmVseSBieSBjaGFuY2UgdGhhdCBzb21lIHRoaW5rZXJzIGhhdmUgY2hvc2V" +
				"uIHRvIHNlZSBpdCBhcyB0aGUgZmluYWwgYW5kIGNsaW5jaGluZyBwcm9vZiBvZiB0aGUg" +
				"bm9uLWV4aXN0ZW5jZSBvZiBHb2QuIFRoZSBhcmd1bWVudCBnb2VzIHNvbWV0aGluZyBsa" +
				"WtlIHRoaXM6ICJJIHJlZnVzZSB0byBwcm92ZSB0aGF0IEkgZXhpc3QsIiBzYXlzIEdvZC" +
				"wgImZvciBwcm9vZiBkZW5pZXMgZmFpdGgsIGFuZCB3aXRob3V0IGZhaXRoIEkgYW0gbm9" +
				"0aGluZy4iICJCdXQsIiBzYXlzIE1hbiwgIlRoZSBCYWJlbCBmaXNoIGlzIGEgZGVhZCBn" +
				"aXZlYXdheSwgaXNuJ3QgaXQ/IEl0IGNvdWxkIG5vdCBoYXZlIGV2b2x2ZWQgYnkgY2hhb" +
				"mNlLiBJdCBwcm92ZXMgeW91IGV4aXN0LCBhbmQgc28gdGhlcmVmb3JlLCBieSB5b3VyIG" +
				"93biBhcmd1bWVudHMsIHlvdSBkb24ndC4gUUVELiIgIk9oIGRlYXIsIiBzYXlzIEdvZCw" +
				"gIkkgaGFkbid0IHRob3VnaHQgb2YgdGhhdCwiIGFuZCBwcm9tcHRseSB2YW5pc2hlcyBp" +
				"biBhIHB1ZmYgb2YgbG9naWMuICJPaCwgdGhhdCB3YXMgZWFzeSwiIHNheXMgTWFuLCBhb" +
				"mQgZm9yIGFuIGVuY29yZSBnb2VzIG9uIHRvIHByb3ZlIHRoYXQgYmxhY2sgaXMgd2hpdG" +
				"UgYW5kIGdldHMgaGltc2VsZiBraWxsZWQgb24gdGhlIG5leHQgemVicmEgY3Jvc3Npbmcu",
			false,
		},
		{
			"HTML",
			[]attachment{{"name.html", strings.NewReader("<html><head></head></html>"), false, ""}},
			"text/html; charset=utf-8;\n\tfilename=\"name.html\"; name=\"name.html\"",
			"attachment;\n\tfilename=\"name.html\"; name=\"name.html\"",
			"PGh0bWw+PGhlYWQ+PC9oZWFkPjwvaHRtbD4=",
			false,
		},
		{
			"HTML - wrong extension",
			[]attachment{{"name.png", strings.NewReader("<html><head></head></html>"), false, ""}},
			"text/html; charset=utf-8;\n\tfilename=\"name.png\"; name=\"name.png\"",
			"attachment;\n\tfilename=\"name.png\"; name=\"name.png\"",
			"PGh0bWw+PGhlYWQ+PC9oZWFkPjwvaHRtbD4=",
			false,
		},

		// inline attachments
		{
			"Empty inline",
			[]attachment{{"Empty", &bytes.Buffer{}, true, ""}},
			"text/plain; charset=utf-8;\n\tfilename=\"Empty\"; name=\"Empty\"",
			"inline;\n\tfilename=\"Empty\"; name=\"Empty\"",
			"",
			false,
		},
		{
			"Short string inline",
			[]attachment{{"advice", strings.NewReader("Don't Panic"), true, ""}},
			"text/plain; charset=utf-8;\n\tfilename=\"advice\"; name=\"advice\"",
			"inline;\n\tfilename=\"advice\"; name=\"advice\"",
			"RG9uJ3QgUGFuaWM=",
			false,
		},
		{
			"Longer string inline",
			[]attachment{
				{
					"partyinvite.txt",
					strings.NewReader(
						"If Baldrick served a meal at HQ he would be arrested for the biggest " +
							"mass poisoning since Lucretia Borgia invited 500 friends for a Wine and Anthrax Party.",
					),
					true,
					"",
				},
			},
			"text/plain; charset=utf-8;\n\tfilename=\"partyinvite.txt\"; name=\"partyinvite.txt\"",
			"inline;\n\tfilename=\"partyinvite.txt\"; name=\"partyinvite.txt\"",
			"SWYgQmFsZHJpY2sgc2VydmVkIGEgbWVhbCBhdCBIUSBoZSB3b3VsZCBiZSBhcnJlc3Rl" +
				"ZCBmb3IgdGhlIGJpZ2dlc3QgbWFzcyBwb2lzb25pbmcgc2luY2UgTHVjcmV0aWEgQm9y" +
				"Z2lhIGludml0ZWQgNTAwIGZyaWVuZHMgZm9yIGEgV2luZSBhbmQgQW50aHJheCBQYXJ0eS4=",
			false,
		},
		{
			"String >512 characters (content type sniff) inline",
			[]attachment{
				{
					"qed.txt",
					strings.NewReader(
						`Now it is such a bizarrely improbable coincidence that anything so mind-bogglingly ` +
							`useful could have evolved purely by chance that some thinkers have chosen to see it ` +
							`as the final and clinching proof of the non-existence of God. The argument goes something ` +
							`like this: "I refuse to prove that I exist," says God, "for proof denies faith, and ` +
							`without faith I am nothing." "But," says Man, "The Babel fish is a dead giveaway, ` +
							`isn't it? It could not have evolved by chance. It proves you exist, and so therefore, ` +
							`by your own arguments, you don't. QED." "Oh dear," says God, "I hadn't thought of ` +
							`that," and promptly vanishes in a puff of logic. "Oh, that was easy," says Man, and ` +
							`for an encore goes on to prove that black is white and gets himself killed on the next ` +
							`zebra crossing.`,
					),
					true,
					"",
				},
			},
			"text/plain; charset=utf-8;\n\tfilename=\"qed.txt\"; name=\"qed.txt\"",
			"inline;\n\tfilename=\"qed.txt\"; name=\"qed.txt\"",
			"Tm93IGl0IGlzIHN1Y2ggYSBiaXphcnJlbHkgaW1wcm9iYWJsZSBjb2luY2lkZW5jZSB0a" +
				"GF0IGFueXRoaW5nIHNvIG1pbmQtYm9nZ2xpbmdseSB1c2VmdWwgY291bGQgaGF2ZSBldm" +
				"9sdmVkIHB1cmVseSBieSBjaGFuY2UgdGhhdCBzb21lIHRoaW5rZXJzIGhhdmUgY2hvc2V" +
				"uIHRvIHNlZSBpdCBhcyB0aGUgZmluYWwgYW5kIGNsaW5jaGluZyBwcm9vZiBvZiB0aGUg" +
				"bm9uLWV4aXN0ZW5jZSBvZiBHb2QuIFRoZSBhcmd1bWVudCBnb2VzIHNvbWV0aGluZyBsa" +
				"WtlIHRoaXM6ICJJIHJlZnVzZSB0byBwcm92ZSB0aGF0IEkgZXhpc3QsIiBzYXlzIEdvZC" +
				"wgImZvciBwcm9vZiBkZW5pZXMgZmFpdGgsIGFuZCB3aXRob3V0IGZhaXRoIEkgYW0gbm9" +
				"0aGluZy4iICJCdXQsIiBzYXlzIE1hbiwgIlRoZSBCYWJlbCBmaXNoIGlzIGEgZGVhZCBn" +
				"aXZlYXdheSwgaXNuJ3QgaXQ/IEl0IGNvdWxkIG5vdCBoYXZlIGV2b2x2ZWQgYnkgY2hhb" +
				"mNlLiBJdCBwcm92ZXMgeW91IGV4aXN0LCBhbmQgc28gdGhlcmVmb3JlLCBieSB5b3VyIG" +
				"93biBhcmd1bWVudHMsIHlvdSBkb24ndC4gUUVELiIgIk9oIGRlYXIsIiBzYXlzIEdvZCw" +
				"gIkkgaGFkbid0IHRob3VnaHQgb2YgdGhhdCwiIGFuZCBwcm9tcHRseSB2YW5pc2hlcyBp" +
				"biBhIHB1ZmYgb2YgbG9naWMuICJPaCwgdGhhdCB3YXMgZWFzeSwiIHNheXMgTWFuLCBhb" +
				"mQgZm9yIGFuIGVuY29yZSBnb2VzIG9uIHRvIHByb3ZlIHRoYXQgYmxhY2sgaXMgd2hpdG" +
				"UgYW5kIGdldHMgaGltc2VsZiBraWxsZWQgb24gdGhlIG5leHQgemVicmEgY3Jvc3Npbmcu",
			false,
		},
		{
			"HTML inline",
			[]attachment{{"name.html", strings.NewReader("<html><head></head></html>"), true, ""}},
			"text/html; charset=utf-8;\n\tfilename=\"name.html\"; name=\"name.html\"",
			"inline;\n\tfilename=\"name.html\"; name=\"name.html\"",
			"PGh0bWw+PGhlYWQ+PC9oZWFkPjwvaHRtbD4=",
			false,
		},
		{
			"HTML - wrong extension inline",
			[]attachment{{"name.png", strings.NewReader("<html><head></head></html>"), true, ""}},
			"text/html; charset=utf-8;\n\tfilename=\"name.png\"; name=\"name.png\"",
			"inline;\n\tfilename=\"name.png\"; name=\"name.png\"",
			"PGh0bWw+PGhlYWQ+PC9oZWFkPjwvaHRtbD4=",
			false,
		},
		{
			"String >512 characters (read full buffer)",
			[]attachment{
				{
					"qed.txt",
					base64.NewDecoder(base64.StdEncoding, strings.NewReader(
						"Tm93IGl0IGlzIHN1Y2ggYSBiaXphcnJlbHkgaW1wcm9iYWJsZSBjb2luY2lkZW5jZSB0a"+
							"GF0IGFueXRoaW5nIHNvIG1pbmQtYm9nZ2xpbmdseSB1c2VmdWwgY291bGQgaGF2ZSBldm"+
							"9sdmVkIHB1cmVseSBieSBjaGFuY2UgdGhhdCBzb21lIHRoaW5rZXJzIGhhdmUgY2hvc2V"+
							"uIHRvIHNlZSBpdCBhcyB0aGUgZmluYWwgYW5kIGNsaW5jaGluZyBwcm9vZiBvZiB0aGUg"+
							"bm9uLWV4aXN0ZW5jZSBvZiBHb2QuIFRoZSBhcmd1bWVudCBnb2VzIHNvbWV0aGluZyBsa"+
							"WtlIHRoaXM6ICJJIHJlZnVzZSB0byBwcm92ZSB0aGF0IEkgZXhpc3QsIiBzYXlzIEdvZC"+
							"wgImZvciBwcm9vZiBkZW5pZXMgZmFpdGgsIGFuZCB3aXRob3V0IGZhaXRoIEkgYW0gbm9"+
							"0aGluZy4iICJCdXQsIiBzYXlzIE1hbiwgIlRoZSBCYWJlbCBmaXNoIGlzIGEgZGVhZCBn"+
							"aXZlYXdheSwgaXNuJ3QgaXQ/IEl0IGNvdWxkIG5vdCBoYXZlIGV2b2x2ZWQgYnkgY2hhb"+
							"mNlLiBJdCBwcm92ZXMgeW91IGV4aXN0LCBhbmQgc28gdGhlcmVmb3JlLCBieSB5b3VyIG"+
							"93biBhcmd1bWVudHMsIHlvdSBkb24ndC4gUUVELiIgIk9oIGRlYXIsIiBzYXlzIEdvZCw"+
							"gIkkgaGFkbid0IHRob3VnaHQgb2YgdGhhdCwiIGFuZCBwcm9tcHRseSB2YW5pc2hlcyBp"+
							"biBhIHB1ZmYgb2YgbG9naWMuICJPaCwgdGhhdCB3YXMgZWFzeSwiIHNheXMgTWFuLCBhb"+
							"mQgZm9yIGFuIGVuY29yZSBnb2VzIG9uIHRvIHByb3ZlIHRoYXQgYmxhY2sgaXMgd2hpdG"+
							"UgYW5kIGdldHMgaGltc2VsZiBraWxsZWQgb24gdGhlIG5leHQgemVicmEgY3Jvc3Npbmcu",
					)),
					false,
					"",
				},
			},
			"text/plain; charset=utf-8;\n\tfilename=\"qed.txt\"; name=\"qed.txt\"",
			"attachment;\n\tfilename=\"qed.txt\"; name=\"qed.txt\"",
			"Tm93IGl0IGlzIHN1Y2ggYSBiaXphcnJlbHkgaW1wcm9iYWJsZSBjb2luY2lkZW5jZSB0a" +
				"GF0IGFueXRoaW5nIHNvIG1pbmQtYm9nZ2xpbmdseSB1c2VmdWwgY291bGQgaGF2ZSBldm" +
				"9sdmVkIHB1cmVseSBieSBjaGFuY2UgdGhhdCBzb21lIHRoaW5rZXJzIGhhdmUgY2hvc2V" +
				"uIHRvIHNlZSBpdCBhcyB0aGUgZmluYWwgYW5kIGNsaW5jaGluZyBwcm9vZiBvZiB0aGUg" +
				"bm9uLWV4aXN0ZW5jZSBvZiBHb2QuIFRoZSBhcmd1bWVudCBnb2VzIHNvbWV0aGluZyBsa" +
				"WtlIHRoaXM6ICJJIHJlZnVzZSB0byBwcm92ZSB0aGF0IEkgZXhpc3QsIiBzYXlzIEdvZC" +
				"wgImZvciBwcm9vZiBkZW5pZXMgZmFpdGgsIGFuZCB3aXRob3V0IGZhaXRoIEkgYW0gbm9" +
				"0aGluZy4iICJCdXQsIiBzYXlzIE1hbiwgIlRoZSBCYWJlbCBmaXNoIGlzIGEgZGVhZCBn" +
				"aXZlYXdheSwgaXNuJ3QgaXQ/IEl0IGNvdWxkIG5vdCBoYXZlIGV2b2x2ZWQgYnkgY2hhb" +
				"mNlLiBJdCBwcm92ZXMgeW91IGV4aXN0LCBhbmQgc28gdGhlcmVmb3JlLCBieSB5b3VyIG" +
				"93biBhcmd1bWVudHMsIHlvdSBkb24ndC4gUUVELiIgIk9oIGRlYXIsIiBzYXlzIEdvZCw" +
				"gIkkgaGFkbid0IHRob3VnaHQgb2YgdGhhdCwiIGFuZCBwcm9tcHRseSB2YW5pc2hlcyBp" +
				"biBhIHB1ZmYgb2YgbG9naWMuICJPaCwgdGhhdCB3YXMgZWFzeSwiIHNheXMgTWFuLCBhb" +
				"mQgZm9yIGFuIGVuY29yZSBnb2VzIG9uIHRvIHByb3ZlIHRoYXQgYmxhY2sgaXMgd2hpdG" +
				"UgYW5kIGdldHMgaGltc2VsZiBraWxsZWQgb24gdGhlIG5leHQgemVicmEgY3Jvc3Npbmcu",
			false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := MailYak{attachments: tt.rattachments}
			pc := testPartCreator{}

			if err := m.writeAttachments(&pc, nopBuilder{}); (err != nil) != tt.wantErr {
				t.Errorf("%q. MailYak.writeAttachments() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			// Ensure there's an attachment
			if len(pc.attachments) != 1 {
				t.Fatalf("%q. MailYak.writeAttachments() unexpected number of attachments = %v, want 1", tt.name, len(pc.attachments))
			}

			if pc.attachments[0].contentType != tt.ctype {
				t.Errorf("%q. MailYak.writeAttachments() content type = %v, want %v", tt.name, pc.attachments[0].contentType, tt.ctype)
			}

			if pc.attachments[0].disposition != tt.disp {
				t.Errorf("%q. MailYak.writeAttachments() disposition = %v, want %v", tt.name, pc.attachments[0].disposition, tt.disp)
			}

			if pc.attachments[0].data.String() != tt.data {
				t.Errorf("%q. MailYak.writeAttachments() data = %v, want %v", tt.name, pc.attachments[0].data.String(), tt.data)
			}
		})
	}
}

// TestMailYakWriteAttachments_multipleAttachments ensures multiple attachments
// are correctly handled
func TestMailYakWriteAttachments_multipleAttachments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rattachments []attachment
		// Expected results.
		want    []testAttachment
		wantErr bool
	}{
		{
			"Single Attachment",
			[]attachment{{"name.txt", strings.NewReader("test"), false, ""}},
			[]testAttachment{
				{
					contentType: "text/plain; charset=utf-8;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					disposition: "attachment;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					data:        *bytes.NewBufferString("dGVzdA=="),
				},
			},
			false,
		},
		{
			"Single Attachment with specified MIME type",
			[]attachment{{"name.txt", strings.NewReader("test"), false, "text/csv; charset=utf-8"}},
			[]testAttachment{
				{
					contentType: "text/csv; charset=utf-8;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					disposition: "attachment;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					data:        *bytes.NewBufferString("dGVzdA=="),
				},
			},
			false,
		},
		{
			"Multiple Attachment - same types",
			[]attachment{
				{"name.txt", strings.NewReader("test"), false, ""},
				{"different.txt", strings.NewReader("another"), false, ""},
			},
			[]testAttachment{
				{
					contentType: "text/plain; charset=utf-8;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					disposition: "attachment;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					data:        *bytes.NewBufferString("dGVzdA=="),
				},
				{
					contentType: "text/plain; charset=utf-8;\n\tfilename=\"different.txt\"; name=\"different.txt\"",
					disposition: "attachment;\n\tfilename=\"different.txt\"; name=\"different.txt\"",
					data:        *bytes.NewBufferString("YW5vdGhlcg=="),
				},
			},
			false,
		},
		{
			"Multiple Attachment - different types",
			[]attachment{
				{"name.txt", strings.NewReader("test"), false, ""},
				{"html.txt", strings.NewReader("<html><head></head></html>"), false, ""},
			},
			[]testAttachment{
				{
					contentType: "text/plain; charset=utf-8;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					disposition: "attachment;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					data:        *bytes.NewBufferString("dGVzdA=="),
				},
				{
					contentType: "text/html; charset=utf-8;\n\tfilename=\"html.txt\"; name=\"html.txt\"",
					disposition: "attachment;\n\tfilename=\"html.txt\"; name=\"html.txt\"",
					data:        *bytes.NewBufferString("PGh0bWw+PGhlYWQ+PC9oZWFkPjwvaHRtbD4="),
				},
			},
			false,
		},
		{
			"Multiple Attachment - different specified MIME types",
			[]attachment{
				{"name.txt", strings.NewReader("test"), false, "text/csv; charset=utf-8"},
				{"html.txt", strings.NewReader("<html><head></head></html>"), false, "application/xml"},
			},
			[]testAttachment{
				{
					contentType: "text/csv; charset=utf-8;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					disposition: "attachment;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					data:        *bytes.NewBufferString("dGVzdA=="),
				},
				{
					contentType: "application/xml;\n\tfilename=\"html.txt\"; name=\"html.txt\"",
					disposition: "attachment;\n\tfilename=\"html.txt\"; name=\"html.txt\"",
					data:        *bytes.NewBufferString("PGh0bWw+PGhlYWQ+PC9oZWFkPjwvaHRtbD4="),
				},
			},
			false,
		},
		{
			"Multiple Attachments - >512 bytes, longer first",
			[]attachment{
				{
					"550.txt",
					strings.NewReader(
						"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Mauris ut nisl felis. " +
							"Aenean felis justo, gravida eget leo aliquet, molestie aliquam risus. Vestibulum " +
							"et nibh rhoncus, malesuada tellus eget, pellentesque diam. Sed venenatis vitae " +
							"erat vel ullamcorper. Aenean rutrum pulvinar purus eget cursus. Integer at iaculis " +
							"arcu. Maecenas mollis nulla dolor, et ultricies massa posuere quis. Nulla facilisi. " +
							"Proin luctus nec nisl at imperdiet. Nulla dapibus purus ut lorem faucibus, at gravida " +
							"tellus euismod. Curabitur ex risus, egestas in porta amet.",
					),
					false,
					"",
				},
				{
					"520.txt", strings.NewReader(
						"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec eu vestibulum dolor. " +
							"Nunc ac posuere felis, a mattis leo. Duis elementum tempor leo, sed efficitur nunc. " +
							"Cras ornare feugiat vulputate. Maecenas sit amet felis lobortis ipsum dignissim euismod. " +
							"Vestibulum id ullamcorper nulla, tincidunt hendrerit justo. Donec vitae eros quam. Nulla " +
							"accumsan porta sapien, in consequat mauris fermentum ac. In at sem lobortis, auctor metus " +
							"rutrum, blandit ipsum. Praesent commodo porta semper. Etiam dignissim libero nullam.",
					),
					false,
					"",
				},
			},
			[]testAttachment{
				{
					contentType: "text/plain; charset=utf-8;\n\tfilename=\"550.txt\"; name=\"550.txt\"",
					disposition: "attachment;\n\tfilename=\"550.txt\"; name=\"550.txt\"",
					data: *bytes.NewBufferString(
						"TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2NpbmcgZWxpdC4gTWF1cmlzIHV0IG5pc" +
							"2wgZmVsaXMuIEFlbmVhbiBmZWxpcyBqdXN0bywgZ3JhdmlkYSBlZ2V0IGxlbyBhbGlxdWV0LCBtb2xlc3RpZSBhbGlxdW" +
							"FtIHJpc3VzLiBWZXN0aWJ1bHVtIGV0IG5pYmggcmhvbmN1cywgbWFsZXN1YWRhIHRlbGx1cyBlZ2V0LCBwZWxsZW50ZXN" +
							"xdWUgZGlhbS4gU2VkIHZlbmVuYXRpcyB2aXRhZSBlcmF0IHZlbCB1bGxhbWNvcnBlci4gQWVuZWFuIHJ1dHJ1bSBwdWx2" +
							"aW5hciBwdXJ1cyBlZ2V0IGN1cnN1cy4gSW50ZWdlciBhdCBpYWN1bGlzIGFyY3UuIE1hZWNlbmFzIG1vbGxpcyBudWxsY" +
							"SBkb2xvciwgZXQgdWx0cmljaWVzIG1hc3NhIHBvc3VlcmUgcXVpcy4gTnVsbGEgZmFjaWxpc2kuIFByb2luIGx1Y3R1cy" +
							"BuZWMgbmlzbCBhdCBpbXBlcmRpZXQuIE51bGxhIGRhcGlidXMgcHVydXMgdXQgbG9yZW0gZmF1Y2lidXMsIGF0IGdyYXZ" +
							"pZGEgdGVsbHVzIGV1aXNtb2QuIEN1cmFiaXR1ciBleCByaXN1cywgZWdlc3RhcyBpbiBwb3J0YSBhbWV0Lg==",
					),
				},
				{
					contentType: "text/plain; charset=utf-8;\n\tfilename=\"520.txt\"; name=\"520.txt\"",
					disposition: "attachment;\n\tfilename=\"520.txt\"; name=\"520.txt\"",
					data: *bytes.NewBufferString(
						"TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2NpbmcgZWxpdC4gRG9uZWMgZXUgdmVz" +
							"dGlidWx1bSBkb2xvci4gTnVuYyBhYyBwb3N1ZXJlIGZlbGlzLCBhIG1hdHRpcyBsZW8uIER1aXMgZWxlbWVudHVtIHRl" +
							"bXBvciBsZW8sIHNlZCBlZmZpY2l0dXIgbnVuYy4gQ3JhcyBvcm5hcmUgZmV1Z2lhdCB2dWxwdXRhdGUuIE1hZWNlbmFz" +
							"IHNpdCBhbWV0IGZlbGlzIGxvYm9ydGlzIGlwc3VtIGRpZ25pc3NpbSBldWlzbW9kLiBWZXN0aWJ1bHVtIGlkIHVsbGFt" +
							"Y29ycGVyIG51bGxhLCB0aW5jaWR1bnQgaGVuZHJlcml0IGp1c3RvLiBEb25lYyB2aXRhZSBlcm9zIHF1YW0uIE51bGxh" +
							"IGFjY3Vtc2FuIHBvcnRhIHNhcGllbiwgaW4gY29uc2VxdWF0IG1hdXJpcyBmZXJtZW50dW0gYWMuIEluIGF0IHNlbSBs" +
							"b2JvcnRpcywgYXVjdG9yIG1ldHVzIHJ1dHJ1bSwgYmxhbmRpdCBpcHN1bS4gUHJhZXNlbnQgY29tbW9kbyBwb3J0YSBz" +
							"ZW1wZXIuIEV0aWFtIGRpZ25pc3NpbSBsaWJlcm8gbnVsbGFtLg==",
					),
				},
			},
			false,
		},
		{
			"Multiple Attachments - >512 bytes, shorter first",
			[]attachment{
				{
					"520.txt",
					strings.NewReader(
						"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec eu vestibulum dolor. Nunc ac " +
							"posuere felis, a mattis leo. Duis elementum tempor leo, sed efficitur nunc. Cras ornare " +
							"feugiat vulputate. Maecenas sit amet felis lobortis ipsum dignissim euismod. Vestibulum " +
							"id ullamcorper nulla, tincidunt hendrerit justo. Donec vitae eros quam. Nulla accumsan " +
							"porta sapien, in consequat mauris fermentum ac. In at sem lobortis, auctor metus rutrum, " +
							"blandit ipsum. Praesent commodo porta semper. Etiam dignissim libero nullam.",
					),
					false,
					"",
				},
				{
					"550.txt",
					strings.NewReader(
						"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Mauris ut nisl felis. Aenean felis " +
							"justo, gravida eget leo aliquet, molestie aliquam risus. Vestibulum et nibh rhoncus, " +
							"malesuada tellus eget, pellentesque diam. Sed venenatis vitae erat vel ullamcorper. " +
							"Aenean rutrum pulvinar purus eget cursus. Integer at iaculis arcu. Maecenas mollis " +
							"nulla dolor, et ultricies massa posuere quis. Nulla facilisi. Proin luctus nec nisl " +
							"at imperdiet. Nulla dapibus purus ut lorem faucibus, at gravida tellus euismod. Curabitur " +
							"ex risus, egestas in porta amet.",
					),
					false,
					"",
				},
			},
			[]testAttachment{
				{
					contentType: "text/plain; charset=utf-8;\n\tfilename=\"520.txt\"; name=\"520.txt\"",
					disposition: "attachment;\n\tfilename=\"520.txt\"; name=\"520.txt\"",
					data: *bytes.NewBufferString(
						"TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2NpbmcgZWxpdC4gRG9uZWMgZXUgdmVz" +
							"dGlidWx1bSBkb2xvci4gTnVuYyBhYyBwb3N1ZXJlIGZlbGlzLCBhIG1hdHRpcyBsZW8uIER1aXMgZWxlbWVudHVtIHRl" +
							"bXBvciBsZW8sIHNlZCBlZmZpY2l0dXIgbnVuYy4gQ3JhcyBvcm5hcmUgZmV1Z2lhdCB2dWxwdXRhdGUuIE1hZWNlbmFz" +
							"IHNpdCBhbWV0IGZlbGlzIGxvYm9ydGlzIGlwc3VtIGRpZ25pc3NpbSBldWlzbW9kLiBWZXN0aWJ1bHVtIGlkIHVsbGFt" +
							"Y29ycGVyIG51bGxhLCB0aW5jaWR1bnQgaGVuZHJlcml0IGp1c3RvLiBEb25lYyB2aXRhZSBlcm9zIHF1YW0uIE51bGxh" +
							"IGFjY3Vtc2FuIHBvcnRhIHNhcGllbiwgaW4gY29uc2VxdWF0IG1hdXJpcyBmZXJtZW50dW0gYWMuIEluIGF0IHNlbSBs" +
							"b2JvcnRpcywgYXVjdG9yIG1ldHVzIHJ1dHJ1bSwgYmxhbmRpdCBpcHN1bS4gUHJhZXNlbnQgY29tbW9kbyBwb3J0YSBz" +
							"ZW1wZXIuIEV0aWFtIGRpZ25pc3NpbSBsaWJlcm8gbnVsbGFtLg==",
					),
				},
				{
					contentType: "text/plain; charset=utf-8;\n\tfilename=\"550.txt\"; name=\"550.txt\"",
					disposition: "attachment;\n\tfilename=\"550.txt\"; name=\"550.txt\"",
					data: *bytes.NewBufferString(
						"TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2NpbmcgZWxpdC4gTWF1cmlzIHV0IG5p" +
							"c2wgZmVsaXMuIEFlbmVhbiBmZWxpcyBqdXN0bywgZ3JhdmlkYSBlZ2V0IGxlbyBhbGlxdWV0LCBtb2xlc3RpZSBhbGlx" +
							"dWFtIHJpc3VzLiBWZXN0aWJ1bHVtIGV0IG5pYmggcmhvbmN1cywgbWFsZXN1YWRhIHRlbGx1cyBlZ2V0LCBwZWxsZW50" +
							"ZXNxdWUgZGlhbS4gU2VkIHZlbmVuYXRpcyB2aXRhZSBlcmF0IHZlbCB1bGxhbWNvcnBlci4gQWVuZWFuIHJ1dHJ1bSBw" +
							"dWx2aW5hciBwdXJ1cyBlZ2V0IGN1cnN1cy4gSW50ZWdlciBhdCBpYWN1bGlzIGFyY3UuIE1hZWNlbmFzIG1vbGxpcyBu" +
							"dWxsYSBkb2xvciwgZXQgdWx0cmljaWVzIG1hc3NhIHBvc3VlcmUgcXVpcy4gTnVsbGEgZmFjaWxpc2kuIFByb2luIGx1" +
							"Y3R1cyBuZWMgbmlzbCBhdCBpbXBlcmRpZXQuIE51bGxhIGRhcGlidXMgcHVydXMgdXQgbG9yZW0gZmF1Y2lidXMsIGF0" +
							"IGdyYXZpZGEgdGVsbHVzIGV1aXNtb2QuIEN1cmFiaXR1ciBleCByaXN1cywgZWdlc3RhcyBpbiBwb3J0YSBhbWV0Lg==",
					),
				},
			},
			false,
		},

		// inline attachments
		{
			"Single Inline Attachment",
			[]attachment{{"name.txt", strings.NewReader("test"), true, ""}},
			[]testAttachment{
				{
					contentType: "text/plain; charset=utf-8;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					disposition: "inline;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					data:        *bytes.NewBufferString("dGVzdA=="),
				},
			},
			false,
		},
		{
			"Single Inline Attachment with specified MIME type",
			[]attachment{{"name.txt", strings.NewReader("test"), true, "text/csv; charset=utf-8"}},
			[]testAttachment{
				{
					contentType: "text/csv; charset=utf-8;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					disposition: "inline;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					data:        *bytes.NewBufferString("dGVzdA=="),
				},
			},
			false,
		},
		{
			"Multiple Inline Attachments - same types",
			[]attachment{
				{"name.txt", strings.NewReader("test"), true, ""},
				{"different.txt", strings.NewReader("another"), true, ""},
			},
			[]testAttachment{
				{
					contentType: "text/plain; charset=utf-8;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					disposition: "inline;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					data:        *bytes.NewBufferString("dGVzdA=="),
				},
				{
					contentType: "text/plain; charset=utf-8;\n\tfilename=\"different.txt\"; name=\"different.txt\"",
					disposition: "inline;\n\tfilename=\"different.txt\"; name=\"different.txt\"",
					data:        *bytes.NewBufferString("YW5vdGhlcg=="),
				},
			},
			false,
		},
		{
			"Multiple Attachments - One Inline, One not",
			[]attachment{
				{"name.txt", strings.NewReader("test"), false, ""},
				{"different.txt", strings.NewReader("another"), true, ""},
			},
			[]testAttachment{
				{
					contentType: "text/plain; charset=utf-8;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					disposition: "attachment;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					data:        *bytes.NewBufferString("dGVzdA=="),
				},
				{
					contentType: "text/plain; charset=utf-8;\n\tfilename=\"different.txt\"; name=\"different.txt\"",
					disposition: "inline;\n\tfilename=\"different.txt\"; name=\"different.txt\"",
					data:        *bytes.NewBufferString("YW5vdGhlcg=="),
				},
			},
			false,
		},
		{
			"Multiple Inline Attachments - different types",
			[]attachment{
				{"name.txt", strings.NewReader("test"), true, ""},
				{"html.txt", strings.NewReader("<html><head></head></html>"), true, ""},
			},
			[]testAttachment{
				{
					contentType: "text/plain; charset=utf-8;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					disposition: "inline;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					data:        *bytes.NewBufferString("dGVzdA=="),
				},
				{
					contentType: "text/html; charset=utf-8;\n\tfilename=\"html.txt\"; name=\"html.txt\"",
					disposition: "inline;\n\tfilename=\"html.txt\"; name=\"html.txt\"",
					data:        *bytes.NewBufferString("PGh0bWw+PGhlYWQ+PC9oZWFkPjwvaHRtbD4="),
				},
			},
			false,
		},
		{
			"Multiple Inline Attachments - specified MIME types",
			[]attachment{
				{"name.txt", strings.NewReader("test"), true, "text/csv; charset=utf-8"},
				{"different.txt", strings.NewReader("<html><head></head></html>"), true, "application/xml"},
			},
			[]testAttachment{
				{
					contentType: "text/csv; charset=utf-8;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					disposition: "inline;\n\tfilename=\"name.txt\"; name=\"name.txt\"",
					data:        *bytes.NewBufferString("dGVzdA=="),
				},
				{
					contentType: "application/xml;\n\tfilename=\"different.txt\"; name=\"different.txt\"",
					disposition: "inline;\n\tfilename=\"different.txt\"; name=\"different.txt\"",
					data:        *bytes.NewBufferString("PGh0bWw+PGhlYWQ+PC9oZWFkPjwvaHRtbD4="),
				},
			},
			false,
		},
		{
			"Multiple Inline Attachments - >512 bytes, longer first",
			[]attachment{
				{
					"550.txt",
					strings.NewReader(
						"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Mauris ut nisl felis. " +
							"Aenean felis justo, gravida eget leo aliquet, molestie aliquam risus. Vestibulum " +
							"et nibh rhoncus, malesuada tellus eget, pellentesque diam. Sed venenatis vitae " +
							"erat vel ullamcorper. Aenean rutrum pulvinar purus eget cursus. Integer at iaculis " +
							"arcu. Maecenas mollis nulla dolor, et ultricies massa posuere quis. Nulla facilisi. " +
							"Proin luctus nec nisl at imperdiet. Nulla dapibus purus ut lorem faucibus, at gravida " +
							"tellus euismod. Curabitur ex risus, egestas in porta amet.",
					),
					true,
					"",
				},
				{
					"520.txt", strings.NewReader(
						"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec eu vestibulum dolor. " +
							"Nunc ac posuere felis, a mattis leo. Duis elementum tempor leo, sed efficitur nunc. " +
							"Cras ornare feugiat vulputate. Maecenas sit amet felis lobortis ipsum dignissim euismod. " +
							"Vestibulum id ullamcorper nulla, tincidunt hendrerit justo. Donec vitae eros quam. Nulla " +
							"accumsan porta sapien, in consequat mauris fermentum ac. In at sem lobortis, auctor metus " +
							"rutrum, blandit ipsum. Praesent commodo porta semper. Etiam dignissim libero nullam.",
					),
					true,
					"",
				},
			},
			[]testAttachment{
				{
					contentType: "text/plain; charset=utf-8;\n\tfilename=\"550.txt\"; name=\"550.txt\"",
					disposition: "inline;\n\tfilename=\"550.txt\"; name=\"550.txt\"",
					data: *bytes.NewBufferString(
						"TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2NpbmcgZWxpdC4gTWF1cmlzIHV0IG5pc" +
							"2wgZmVsaXMuIEFlbmVhbiBmZWxpcyBqdXN0bywgZ3JhdmlkYSBlZ2V0IGxlbyBhbGlxdWV0LCBtb2xlc3RpZSBhbGlxdW" +
							"FtIHJpc3VzLiBWZXN0aWJ1bHVtIGV0IG5pYmggcmhvbmN1cywgbWFsZXN1YWRhIHRlbGx1cyBlZ2V0LCBwZWxsZW50ZXN" +
							"xdWUgZGlhbS4gU2VkIHZlbmVuYXRpcyB2aXRhZSBlcmF0IHZlbCB1bGxhbWNvcnBlci4gQWVuZWFuIHJ1dHJ1bSBwdWx2" +
							"aW5hciBwdXJ1cyBlZ2V0IGN1cnN1cy4gSW50ZWdlciBhdCBpYWN1bGlzIGFyY3UuIE1hZWNlbmFzIG1vbGxpcyBudWxsY" +
							"SBkb2xvciwgZXQgdWx0cmljaWVzIG1hc3NhIHBvc3VlcmUgcXVpcy4gTnVsbGEgZmFjaWxpc2kuIFByb2luIGx1Y3R1cy" +
							"BuZWMgbmlzbCBhdCBpbXBlcmRpZXQuIE51bGxhIGRhcGlidXMgcHVydXMgdXQgbG9yZW0gZmF1Y2lidXMsIGF0IGdyYXZ" +
							"pZGEgdGVsbHVzIGV1aXNtb2QuIEN1cmFiaXR1ciBleCByaXN1cywgZWdlc3RhcyBpbiBwb3J0YSBhbWV0Lg==",
					),
				},
				{
					contentType: "text/plain; charset=utf-8;\n\tfilename=\"520.txt\"; name=\"520.txt\"",
					disposition: "inline;\n\tfilename=\"520.txt\"; name=\"520.txt\"",
					data: *bytes.NewBufferString(
						"TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2NpbmcgZWxpdC4gRG9uZWMgZXUgdmVz" +
							"dGlidWx1bSBkb2xvci4gTnVuYyBhYyBwb3N1ZXJlIGZlbGlzLCBhIG1hdHRpcyBsZW8uIER1aXMgZWxlbWVudHVtIHRl" +
							"bXBvciBsZW8sIHNlZCBlZmZpY2l0dXIgbnVuYy4gQ3JhcyBvcm5hcmUgZmV1Z2lhdCB2dWxwdXRhdGUuIE1hZWNlbmFz" +
							"IHNpdCBhbWV0IGZlbGlzIGxvYm9ydGlzIGlwc3VtIGRpZ25pc3NpbSBldWlzbW9kLiBWZXN0aWJ1bHVtIGlkIHVsbGFt" +
							"Y29ycGVyIG51bGxhLCB0aW5jaWR1bnQgaGVuZHJlcml0IGp1c3RvLiBEb25lYyB2aXRhZSBlcm9zIHF1YW0uIE51bGxh" +
							"IGFjY3Vtc2FuIHBvcnRhIHNhcGllbiwgaW4gY29uc2VxdWF0IG1hdXJpcyBmZXJtZW50dW0gYWMuIEluIGF0IHNlbSBs" +
							"b2JvcnRpcywgYXVjdG9yIG1ldHVzIHJ1dHJ1bSwgYmxhbmRpdCBpcHN1bS4gUHJhZXNlbnQgY29tbW9kbyBwb3J0YSBz" +
							"ZW1wZXIuIEV0aWFtIGRpZ25pc3NpbSBsaWJlcm8gbnVsbGFtLg==",
					),
				},
			},
			false,
		},
		{
			"Multiple Inline Attachments - >512 bytes, shorter first",
			[]attachment{
				{
					"520.txt",
					strings.NewReader(
						"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec eu vestibulum dolor. Nunc ac " +
							"posuere felis, a mattis leo. Duis elementum tempor leo, sed efficitur nunc. Cras ornare " +
							"feugiat vulputate. Maecenas sit amet felis lobortis ipsum dignissim euismod. Vestibulum " +
							"id ullamcorper nulla, tincidunt hendrerit justo. Donec vitae eros quam. Nulla accumsan " +
							"porta sapien, in consequat mauris fermentum ac. In at sem lobortis, auctor metus rutrum, " +
							"blandit ipsum. Praesent commodo porta semper. Etiam dignissim libero nullam.",
					),
					true,
					"",
				},
				{
					"550.txt",
					strings.NewReader(
						"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Mauris ut nisl felis. Aenean felis " +
							"justo, gravida eget leo aliquet, molestie aliquam risus. Vestibulum et nibh rhoncus, " +
							"malesuada tellus eget, pellentesque diam. Sed venenatis vitae erat vel ullamcorper. " +
							"Aenean rutrum pulvinar purus eget cursus. Integer at iaculis arcu. Maecenas mollis " +
							"nulla dolor, et ultricies massa posuere quis. Nulla facilisi. Proin luctus nec nisl " +
							"at imperdiet. Nulla dapibus purus ut lorem faucibus, at gravida tellus euismod. Curabitur " +
							"ex risus, egestas in porta amet.",
					),
					true,
					"",
				},
			},
			[]testAttachment{
				{
					contentType: "text/plain; charset=utf-8;\n\tfilename=\"520.txt\"; name=\"520.txt\"",
					disposition: "inline;\n\tfilename=\"520.txt\"; name=\"520.txt\"",
					data: *bytes.NewBufferString(
						"TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2NpbmcgZWxpdC4gRG9uZWMgZXUgdmVz" +
							"dGlidWx1bSBkb2xvci4gTnVuYyBhYyBwb3N1ZXJlIGZlbGlzLCBhIG1hdHRpcyBsZW8uIER1aXMgZWxlbWVudHVtIHRl" +
							"bXBvciBsZW8sIHNlZCBlZmZpY2l0dXIgbnVuYy4gQ3JhcyBvcm5hcmUgZmV1Z2lhdCB2dWxwdXRhdGUuIE1hZWNlbmFz" +
							"IHNpdCBhbWV0IGZlbGlzIGxvYm9ydGlzIGlwc3VtIGRpZ25pc3NpbSBldWlzbW9kLiBWZXN0aWJ1bHVtIGlkIHVsbGFt" +
							"Y29ycGVyIG51bGxhLCB0aW5jaWR1bnQgaGVuZHJlcml0IGp1c3RvLiBEb25lYyB2aXRhZSBlcm9zIHF1YW0uIE51bGxh" +
							"IGFjY3Vtc2FuIHBvcnRhIHNhcGllbiwgaW4gY29uc2VxdWF0IG1hdXJpcyBmZXJtZW50dW0gYWMuIEluIGF0IHNlbSBs" +
							"b2JvcnRpcywgYXVjdG9yIG1ldHVzIHJ1dHJ1bSwgYmxhbmRpdCBpcHN1bS4gUHJhZXNlbnQgY29tbW9kbyBwb3J0YSBz" +
							"ZW1wZXIuIEV0aWFtIGRpZ25pc3NpbSBsaWJlcm8gbnVsbGFtLg==",
					),
				},
				{
					contentType: "text/plain; charset=utf-8;\n\tfilename=\"550.txt\"; name=\"550.txt\"",
					disposition: "inline;\n\tfilename=\"550.txt\"; name=\"550.txt\"",
					data: *bytes.NewBufferString(
						"TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2NpbmcgZWxpdC4gTWF1cmlzIHV0IG5p" +
							"c2wgZmVsaXMuIEFlbmVhbiBmZWxpcyBqdXN0bywgZ3JhdmlkYSBlZ2V0IGxlbyBhbGlxdWV0LCBtb2xlc3RpZSBhbGlx" +
							"dWFtIHJpc3VzLiBWZXN0aWJ1bHVtIGV0IG5pYmggcmhvbmN1cywgbWFsZXN1YWRhIHRlbGx1cyBlZ2V0LCBwZWxsZW50" +
							"ZXNxdWUgZGlhbS4gU2VkIHZlbmVuYXRpcyB2aXRhZSBlcmF0IHZlbCB1bGxhbWNvcnBlci4gQWVuZWFuIHJ1dHJ1bSBw" +
							"dWx2aW5hciBwdXJ1cyBlZ2V0IGN1cnN1cy4gSW50ZWdlciBhdCBpYWN1bGlzIGFyY3UuIE1hZWNlbmFzIG1vbGxpcyBu" +
							"dWxsYSBkb2xvciwgZXQgdWx0cmljaWVzIG1hc3NhIHBvc3VlcmUgcXVpcy4gTnVsbGEgZmFjaWxpc2kuIFByb2luIGx1" +
							"Y3R1cyBuZWMgbmlzbCBhdCBpbXBlcmRpZXQuIE51bGxhIGRhcGlidXMgcHVydXMgdXQgbG9yZW0gZmF1Y2lidXMsIGF0" +
							"IGdyYXZpZGEgdGVsbHVzIGV1aXNtb2QuIEN1cmFiaXR1ciBleCByaXN1cywgZWdlc3RhcyBpbiBwb3J0YSBhbWV0Lg==",
					),
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := MailYak{attachments: tt.rattachments}
			pc := testPartCreator{}

			if err := m.writeAttachments(&pc, nopBuilder{}); (err != nil) != tt.wantErr {
				t.Errorf("%q. MailYak.writeAttachments() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			// Did we get enough attachments?
			if len(tt.want) != len(pc.attachments) {
				t.Fatalf("%q. MailYak.writeAttachments() unexpected number of attachments = %v, want %v", tt.name, len(pc.attachments), len(tt.want))
			}

			for i, want := range tt.want {
				got := pc.attachments[i]

				if want.contentType != got.contentType {
					t.Errorf("%q. MailYak.writeAttachments() content type = %v, want %v", tt.name, want.contentType, got.contentType)
				}

				if want.disposition != got.disposition {
					t.Errorf("%q. MailYak.writeAttachments() disposition = %v, want %v", tt.name, want.disposition, got.disposition)
				}

				if !bytes.Equal(want.data.Bytes(), got.data.Bytes()) {
					t.Errorf("%q. MailYak.writeAttachments() data = %v, want %v", tt.name, want.data.String(), got.data.String())
				}
			}
		})
	}
}
