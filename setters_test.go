package mailyak

import (
	"reflect"
	"regexp"
	"testing"
)

func TestMailYakTo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// Test description.
		name string
		// Parameters.
		addrs []string
		// Want
		want []string
	}{
		{
			"Single email",
			[]string{"dom@itsallbroken.com"},
			[]string{"dom@itsallbroken.com"},
		},
		{
			"Multiple email",
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
		},
		{
			"Empty last",
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com", ""},
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
		},
		{
			"Empty Middle",
			[]string{"dom@itsallbroken.com", "", "ohnoes@itsallbroken.com"},
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := &MailYak{
				toAddrs:   []string{},
				trimRegex: regexp.MustCompile("\r?\n"),
			}
			m.To(tt.addrs...)

			if !reflect.DeepEqual(m.toAddrs, tt.want) {
				t.Errorf("%q. MailYak.To() = %v, want %v", tt.name, m.toAddrs, tt.want)
			}
		})
	}
}

func TestMailYakBcc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// Test description.
		name string
		// Parameters.
		addrs []string
		// Want
		want []string
	}{
		{
			"Single email",
			[]string{"dom@itsallbroken.com"},
			[]string{"dom@itsallbroken.com"},
		},
		{
			"Multiple email",
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
		},
		{
			"Empty last",
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com", ""},
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
		},
		{
			"Empty Middle",
			[]string{"dom@itsallbroken.com", "", "ohnoes@itsallbroken.com"},
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := &MailYak{
				bccAddrs:  []string{},
				trimRegex: regexp.MustCompile("\r?\n"),
			}
			m.Bcc(tt.addrs...)

			if !reflect.DeepEqual(m.bccAddrs, tt.want) {
				t.Errorf("%q. MailYak.Bcc() = %v, want %v", tt.name, m.bccAddrs, tt.want)
			}
		})
	}
}
