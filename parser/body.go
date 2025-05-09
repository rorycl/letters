package parser

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/rorycl/letters/decoders"
	"github.com/rorycl/letters/email"
)

// parseBody parses the body of an email
func (se *stagedEmail) parseBody() error {

	var err error
	switch se.contentInfo.Type {
	case "text/plain":
		se.email.Text, err = se.parseText(se.msg.Body, se.contentInfo)
		if err != nil {
			return fmt.Errorf("cannot parse plain text: %w", err)
		}
		return nil

	case "text/enriched":
		se.email.EnrichedText, err = se.parseText(se.msg.Body, se.contentInfo)
		if err != nil {
			return fmt.Errorf("cannot parse enriched text: %w", err)
		}
		return nil

	case "text/html":
		se.email.HTML, err = se.parseText(se.msg.Body, se.contentInfo)
		if err != nil {
			return fmt.Errorf("cannot parse html text: %w", err)
		}
		return nil
	}
	return fmt.Errorf("parse body content type %q not known", se.contentInfo.Type)

}

// parseText parses the text content of an email body or mime part. Note
// that mime parts can be nested inside other mime parts.
func (se *stagedEmail) parseText(t io.Reader, ci *email.ContentInfo) (string, error) {
	reader := decoders.DecodeContent(t, ci)
	textBody, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("cannot read plain text content: %w", err)
	}
	textBody = bytes.ReplaceAll(textBody, []byte("\r\n"), []byte("\n"))
	return strings.TrimSpace(string(textBody)), nil
}
