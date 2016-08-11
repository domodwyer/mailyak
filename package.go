// Package mailyak provides a simple interface for sending MIME compliant emails
// over SMTP.
//
// Attachments are fully supported (attach anything that implements io.Reader)
// and the code is tried and tested in a production setting.
//
// For convenience the HTML and Plain methods return a type implementing io.Writer,
// allowing email bodies to be composed directly from templating engines.
package mailyak
