package mailyak

import (
	"bytes"
	"fmt"
	"io"
	"net/smtp"
	"strings"
	"testing"
)

// TestHTML ensures we can write to HTML as an io.Writer
func TestHTML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		// Parameters.
		data []string
	}{
		{
			"Writer test",
			[]string{"Worst idea since someone said ‘yeah let’s take this suspiciously large " +
				"wooden horse into Troy, statues are all the rage this season’.",
			},
		},
		{
			"Writer test multiple",
			[]string{
				"Worst idea since someone said ‘yeah let’s take this suspiciously large " +
					"wooden horse into Troy, statues are all the rage this season’.",
				"Am I jumping the gun, Baldrick, or are the words 'I have a cunning plan' " +
					"marching with ill-deserved confidence in the direction of this conversation?",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mail := New("", smtp.PlainAuth("", "", "", ""))

			for _, data := range tt.data {
				if _, err := io.WriteString(mail.HTML(), data); err != nil {
					t.Errorf("%q. HTML() error = %v", tt.name, err)
					continue
				}
			}

			if !bytes.Equal([]byte(strings.Join(tt.data, "")), mail.html.Bytes()) {
				t.Errorf("%q. HTML() = %v, want %v", tt.name, mail.html.String(), tt.data)
			}
		})
	}
}

// TestPlain ensures we can write to Plain as an io.Writer
func TestPlain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		// Parameters.
		data string
	}{
		{
			"Writer test",
			"Am I jumping the gun, Baldrick, or are the words 'I have a cunning plan' " +
				"marching with ill-deserved confidence in the direction of this conversation?",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mail := New("", smtp.PlainAuth("", "", "", ""))

			if _, err := io.WriteString(mail.Plain(), tt.data); err != nil {
				t.Fatalf("%q. Plain() error = %v", tt.name, err)
			}

			if !bytes.Equal([]byte(tt.data), mail.plain.Bytes()) {
				t.Errorf("%q. Plain() = %v, want %v", tt.name, mail.plain.String(), tt.data)
			}
		})
	}
}

// TestWritableString ensures the writable type returns a string when called
// with fmt.Printx(), etc
func TestWritableString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		// Parameters.
		data string
	}{
		{
			"String test",
			"Baldrick, does it have to be this way? " +
				"Our valued friendship ending with me cutting you up into strips and telling " +
				"the prince that you walked over a very sharp cattle grid in an extremely heavy hat?",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mail := New("", smtp.PlainAuth("", "", "", ""))

			if _, err := io.WriteString(mail.Plain(), tt.data); err != nil {
				t.Fatalf("%q. Plain() error = %v", tt.name, err)
			}

			if tt.data != mail.plain.String() {
				t.Errorf("%q. writable.String() = %v, want %v", tt.name, mail.plain.String(), tt.data)
			}

			if out := fmt.Sprintf("%v", mail.plain.String()); out != tt.data {
				t.Errorf("%q. writable.String() via fmt.Sprintf = %v, want %v", tt.name, out, tt.data)
			}
		})
	}
}

// TestPlain_String ensures we can use the string setter
func TestPlain_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		// Parameters.
		data string
	}{
		{
			"Writer test",
			"Am I jumping the gun, Baldrick, or are the words 'I have a cunning plan' " +
				"marching with ill-deserved confidence in the direction of this conversation?",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mail := New("", smtp.PlainAuth("", "", "", ""))

			if _, err := io.WriteString(mail.Plain(), tt.data); err != nil {
				t.Fatalf("%q. Plain() error = %v", tt.name, err)
			}

			if !bytes.Equal([]byte(tt.data), mail.plain.Bytes()) {
				t.Errorf("%q. Plain() = %v, want %v", tt.name, mail.plain.String(), tt.data)
			}
		})
	}
}
