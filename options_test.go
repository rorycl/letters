// This is a test showing how ParserOptions could be implemented using
// the golang functional options pattern.
//
// A usefulblog post on the subject is at
// https://davidbacisin.com/writing/golang-options-pattern
// See also the definitive blog post at
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
//
// RCL 11 January 2025

package letters

import (
	"fmt"
	"io"
	"net/mail"
	"strings"
	"testing"
	"time"
)

type Option func(*ParserOptions)

type parserType int

const (
	wholeEmail parserType = iota
	headersOnly
	skipAttachments
)

type ParserOptions struct {
	parser              parserType
	parseAddressHeader  func(mail.Header, string) (*mail.Address, error)
	parseDateHeaderFunc func(string) time.Time
	parserFunc          func(io.Reader) (Email, error)
}

func NewParser(options ...Option) (*ParserOptions, error) {
	// defaults
	p := &ParserOptions{
		parser:              wholeEmail,
		parseAddressHeader:  parseAddressHeader,
		parseDateHeaderFunc: parseDateHeader,
		parserFunc:          ParseEmail,
	}
	for _, opt := range options {
		opt(p)
	}
	return p, nil
}

// new main entry point
func (p *ParserOptions) Parse(r io.Reader) (Email, error) {
	return p.parserFunc(r)
}

// some options

func WithHeadersOnly() Option {
	return func(p *ParserOptions) {
		p.parser = headersOnly
	}
}

func WithSkipAttachments() Option {
	return func(p *ParserOptions) {
		p.parser = skipAttachments
	}
}

func WithAddressHeaderFunc(f func(mail.Header, string) (*mail.Address, error)) Option {
	return func(p *ParserOptions) {
		p.parseAddressHeader = f
	}
}

func TestOptionParse(t *testing.T) {

	messageReader := strings.NewReader(`From: John Doe <jdoe@machine.example>
To: Mary Smith <mary@example.net>
Subject: Saying Hello
Date: Fri, 21 Nov 1997 09:55:06 -0600
Message-ID: <1234@local.machine.example>

This is a message just to say hello.
So, "Hello".
`)

	parser, _ := NewParser(
		WithSkipAttachments(),
		WithAddressHeaderFunc(
			func(mail.Header, string) (*mail.Address, error) {
				return &mail.Address{Name: "Joe Bloggs", Address: "joe@bloggs.com"}, nil
			},
		),
	)

	// testing override of main parserFun (normally ParseEmail)
	parser.parserFunc = func(r io.Reader) (Email, error) {
		e := Email{}
		email, err := mail.ReadMessage(r)
		if err != nil {
			return e, err
		}
		e.Headers, err = ParseHeaders(email.Header)
		if err != nil {
			return e, err
		}
		e.Headers.From[0], err = parser.parseAddressHeader(email.Header, "From")
		if err != nil {
			return e, err
		}
		text, err := io.ReadAll(email.Body)
		if err != nil {
			return e, err
		}
		e.Text = string(text)
		return e, nil
	}

	// run
	email, err := parser.Parse(messageReader)

	// view and validate
	fmt.Printf("email from    : %#v\n", email.Headers.From[0])
	fmt.Printf("email subject : %s\n", email.Headers.Subject)
	fmt.Printf("email text    : %s", email.Text)
	fmt.Printf("email err     : %t \n", err != nil)

	if got, want := email.Headers.From[0].Address, "joe@bloggs.com"; got != want {
		t.Errorf("Address.Address got %s want %s", got, want)
	}

	tpl := `parser
	parser:              %s
	parseAddressHeader:  %T
	parseDateHeaderFunc: %T
	parserFunc:          %T
`
	pT := ""
	switch parser.parser {
	case wholeEmail:
		pT = "wholeEmail"
	case headersOnly:
		pT = "headersOnly"
	case skipAttachments:
		pT = "skipAttachments"
	}
	fmt.Printf(tpl, pT, parser.parseAddressHeader, parser.parseDateHeaderFunc, parser.parserFunc)

	if got, want := pT, "skipAttachments"; got != want {
		t.Errorf("parser mode got %s want %s", got, want)
	}

	if err != nil {
		t.Fatal(err)
	}

}
