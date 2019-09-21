// This package provides a easy to use MIME email composer with support for
// attachments.
package mailyak

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/smtp"
	"text/template"
)

func Example() {
	// Create a new email - specify the SMTP host and auth
	mail := New("mail.host.com:25", smtp.PlainAuth("", "user", "pass", "mail.host.com"))
	// You can also use NewWithTLS to provide TLS configuration
	mail = NewWithTLS("mail.host.com:25", smtp.PlainAuth("", "user", "pass", "mail.host.com"), &tls.Config{InsecureSkipVerify: true})

	mail.To("dom@itsallbroken.com")
	mail.From("jsmith@example.com")
	mail.FromName("Prince Anybody")

	mail.Subject("Business proposition")

	// Add a custom header
	mail.AddHeader("X-TOTALLY-NOT-A-SCAM", "true")

	// mail.HTMLWriter() and mail.PlainWriter() implement io.Writer, so you can
	// do handy things like parse a template directly into the email body - here
	// we just use io.WriteString()
	if _, err := io.WriteString(mail.HTML(), "So long, and thanks for all the fish."); err != nil {
		panic(" :( ")
	}

	// Or set the body using a string helper
	mail.Plain().Set("Get a real email client")

	// And you're done!
	if err := mail.Send(); err != nil {
		panic(" :( ")
	}
}

func Example_attachments() {
	// This will be our attachment data
	buf := &bytes.Buffer{}
	io.WriteString(buf, "We're in the stickiest situation since Sticky the Stick Insect got stuck on a sticky bun.")

	// Create a new email - specify the SMTP host and auth
	mail := New("mail.host.com:25", smtp.PlainAuth("", "user", "pass", "mail.host.com"))

	mail.To("dom@itsallbroken.com")
	mail.From("jsmith@example.com")
	mail.HTML().Set("I am an email")

	// buf could be anything that implements io.Reader
	mail.Attach("sticky.txt", buf)

	if err := mail.Send(); err != nil {
		panic(" :( ")
	}
}

func ExampleBodyPart_string() {
	// Create a new email - specify the SMTP host and auth
	mail := New("mail.host.com:25", smtp.PlainAuth("", "user", "pass", "mail.host.com"))

	// Set the plain text email content using a string
	mail.Plain().Set("Get a real email client")
}

func ExampleBodyPart_templates() {
	// Create a new email
	mail := New("mail.host.com:25", smtp.PlainAuth("", "user", "pass", "mail.host.com"))

	// Our pretend template data
	tmplData := struct {
		Language string
	}{"Go"}

	// Compile a template
	tmpl, err := template.New("html").Parse("I am an email template in {{ .Language }}")
	if err != nil {
		panic(" :( ")
	}

	// Execute the template directly into the email body
	if err := tmpl.Execute(mail.HTML(), tmplData); err != nil {
		panic(" :( ")
	}
}
