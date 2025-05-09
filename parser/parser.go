// Package parser provides the capabilities for parsing an email
// io.Reader into an [email.Email]. The parser can receive options of
// type `Opt` which alter the parsing process.
package parser

import (
	"fmt"
	"io"
	"net/mail"
	"strings"
	"time"

	"github.com/rorycl/letters/email"
)

// UnknownContentTypeError reports an unknown Content Type
type UnknownContentTypeError struct {
	contentType string
}

func (e *UnknownContentTypeError) Error() string {
	return fmt.Sprintf("unknown Content-Type %q", e.contentType)
}

// typeOfProcessing determines the type of processing to be done by the
// Parser. If processing many emails it will be much more efficient to
// use the `noAttachments` or `headersOnly` processing types if the
// whole email isn't needed.
type typeOfProcessing string

const (
	wholeEmail    typeOfProcessing = "wholeEmail"
	headersOnly   typeOfProcessing = "headersOnly"
	noAttachments typeOfProcessing = "noAttachments"
)

// Opt is a parser option type provided as a closure to add options to a
// parser default instance instantiated by NewParser. The options are
// held in opts.go providing closures returning an Opt such as
// WithHeadersOnly().
type Opt func(p *Parser)

// Parser is the structure holding the parser settings including
// processType, skipContentTypes and address, file and date processing
// funcs, all of which should be set by Opt closures if the defaults
// aren't to be used.
//
// The default address and date parsers are provided by net/mail
// (mail.ParseAddress and mail.ParseAddressList, mail.ParseDate) while
// the default attachment func is to simply ready each attachment into
// the slice of email.File.Data.
type Parser struct {
	// what parts of the email to process (default all)
	processType typeOfProcessing
	// skipContentTypes is a list of content types to skip
	skipContentTypes []string

	// funcs that can be overridden by the user; defaults are set
	// attached by NewParser.
	// addressFunc : the function for processing email header addresses
	addressFunc func(string) (*mail.Address, error)
	// addressesFunc: the functionfor processing a list of email header
	// addresses
	addressesFunc func(list string) ([]*mail.Address, error)
	// dateFunc : the function for processing the email header Date
	dateFunc func(string) (time.Time, error)
	// fileFunc : a function for processing inline and attached files
	fileFunc func(*email.File) error

	// debugging, for future use
	verbose bool
}

// NewParser initialises a new Parser. The default parser can be changed
// using options returning an Opt.
func NewParser(options ...Opt) *Parser {
	p := &Parser{
		// initialise main fields
		processType: wholeEmail,

		// initialise overrideable funcs
		// use net/mail.ParseAddress and ParseAddressList  as default
		// address parsers
		addressFunc:   mail.ParseAddress,
		addressesFunc: mail.ParseAddressList,
		// use net/mail.ParseDate as the default date parser
		dateFunc: mail.ParseDate,
		// by default write file io.Readers to email.File.Data.
		// User-supplied funcs might write files directly to disk, for
		// example, bypassing this step.
		fileFunc: func(f *email.File) error {
			var err error
			f.Data, err = io.ReadAll(f.Reader)
			return err
		},

		// debugging
		verbose: false,
	}

	for _, opt := range options {
		opt(p)
	}
	return p
}

// Parse is the main entry point of letters.
func (p *Parser) Parse(r io.Reader) (*email.Email, error) {
	var err error
	se := newStagedEmail(p)

	// read the message into a *mail.Message
	se.msg, err = mail.ReadMessage(r)
	if err != nil {
		return nil, fmt.Errorf("cannot read message: %w", err)
	}

	// extract content information
	se.contentInfo, err = email.ExtractContentInfo(se.msg.Header, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot extract content: %w", err)
	}

	// parse headers
	err = se.parseHeaders()
	if err != nil {
		return nil, fmt.Errorf("cannot parse headers: %w", err)
	}
	if p.processType == headersOnly {
		return se.email, nil
	}

	switch ct := se.contentInfo.Type; { // true switch

	case ct == "text/plain", ct == "text/enriched", ct == "text/html":
		// parse body
		err = se.parseBody()
		if err != nil {
			return nil, err
		}

	case strings.HasPrefix(ct, "multipart/"):
		// parse parts
		err = se.parsePart(
			se.msg.Body,
			se.contentInfo,
			se.contentInfo.TypeParams["boundary"],
		)
		if err != nil {
			return nil, err
		}

	default:
		// parse attachment
		err = se.parseFile(se.msg.Body, se.contentInfo)
		if err != nil {
			return nil, err
		}
	}
	return se.email, err
}
