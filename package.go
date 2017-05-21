// Package mailyak provides a simple interface for generating MIME compliant
// emails, and optionally sending them over SMTP.
//
// Both plain-text and HTML email body content is supported, and their types
// implement io.Writer allowing easy composition directly from templating
// engines, etc.
//
// Attachments are fully supported (attach anything that implements io.Reader).
//
// The raw MIME content can be retrieved using MimeBuf(), typically used with an
// API service such as Amazon SES that does not require using an SMTP interface.
package mailyak
