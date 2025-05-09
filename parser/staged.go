package parser

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/mail"
	"strings"

	"github.com/rorycl/letters/email"
)

// "staged" sets out the stagedEmail that the parsing process uses to create
// output and related methods.

// stagedEmail is the structure used to incrementally build up an
// *email.Email for returning to parser.Parse.
type stagedEmail struct {
	// parser is a reference to the parsing setting struct
	parser *Parser

	// msg is the net/mail.Message used for deriving parts to build
	// the output email.
	msg *mail.Message

	// the main email content info (note that parent contentinfo is also
	// used elsewhere in processing)
	contentInfo *email.ContentInfo

	// email to be built and returned, for incremental processing
	email *email.Email
}

// newStagedEmail returns an initialised *stagedEmail
func newStagedEmail(p *Parser) *stagedEmail {
	return &stagedEmail{
		parser: p,
		email:  &email.Email{},
		msg:    &mail.Message{},
	}
}

// parsePart parses the parts of a multipart message and may be called
// recursively.
func (se *stagedEmail) parsePart(msg io.Reader, parentCI *email.ContentInfo, boundary string) error {

	multipartReader := multipart.NewReader(msg, boundary)
	if multipartReader == nil {
		return nil
	}

	for {
		part, err := multipartReader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("cannot read part: %w", err)
		}

		// extract content information
		contentInfo, err := email.ExtractContentInfo(part.Header, se.contentInfo)
		if err != nil {
			return fmt.Errorf("content extraction error: %w", err)
		}

		// skip part if the content type is in parser.skipContentTypes
		if se.parser.inSkipContentTypes(contentInfo.Type) {
			continue
		}

		// commence extraction of data with attached file
		if contentInfo.Disposition == "attachment" {
			err = se.parseFile(
				part,
				contentInfo,
			)
			if err != nil {
				return fmt.Errorf("cannot parse attached file: %w", err)
			}
			continue
		}

		// process text plain content
		if contentInfo.Type == "text/plain" {
			partTextBody, err := se.parseText(part, contentInfo)
			if err != nil {
				return fmt.Errorf("cannot parse plain text: %w", err)
			}
			if len(se.email.Text) > 0 { // add separator
				se.email.Text += "\n\n"
			}
			se.email.Text += partTextBody
			continue
		}

		// process text enriched content
		if contentInfo.Type == "text/enriched" {
			partEnrichedText, err := se.parseText(part, contentInfo)
			if err != nil {
				return fmt.Errorf("cannot parse enriched text: %w", err)
			}
			se.email.EnrichedText += partEnrichedText
			continue
		}

		// process html content
		if contentInfo.Type == "text/html" {
			partHtmlBody, err := se.parseText(part, contentInfo)
			if err != nil {
				return fmt.Errorf("cannot parse html text: %w", err)
			}
			se.email.HTML += partHtmlBody
			continue
		}

		// recursive call to parsePart
		if strings.HasPrefix(contentInfo.Type, "multipart") {
			err := se.parsePart(part, contentInfo, contentInfo.TypeParams["boundary"])
			if err != nil {
				return fmt.Errorf("cannot parse nested part: %w", err)
			}
			continue
		}

		// process inline file
		if contentInfo.IsInlineFile(contentInfo) {
			if se.parser.processType != wholeEmail {
				continue
			}
			err = se.parseFile(part, contentInfo)
			if err != nil {
				return fmt.Errorf("cannot parse inline file: %w", err)
			}
			continue
		}

		// process attached file
		if contentInfo.IsAttachedFile(contentInfo) {
			if se.parser.processType != wholeEmail {
				continue
			}
			err := se.parseFile(part, contentInfo)
			if err != nil {
				return fmt.Errorf("cannot parse attached file: %w", err)
			}
			continue
		}

		// types to ignore
		// Todo/fixme
		// This section needs to be expanded or, alternatively and more
		// sensibly, expanded and moved to contentInfo

		// unhandled types fixme
		switch contentInfo.Type {
		case "text/calendar":
			fmt.Println("skipping text/calendar content-type")
			continue
		}

		// fallthrough error
		return &UnknownContentTypeError{contentType: contentInfo.Type}
	}

	return nil
}
