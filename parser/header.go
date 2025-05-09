package parser

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/rorycl/letters/decoders"
)

var (
	errorEmptyAddress error = errors.New("Empty Address")
	errorEmptyDate    error = errors.New("Empty Date")
)

// explicitHeaders are those headers stored in their own field in
// email.Headers, rather than in email.Headers.ExtraHeaders
var explicitHeaders = []string{
	"Date",
	"Sender",
	"From",
	"Reply-To",
	"To",
	"Cc",
	"Bcc",
	"Message-Id",
	"In-Reply-To",
	"References",
	"Received",
	"Subject",
	"Comments",
	"Keywords",
	"Resent-Date",
	"Resent-From",
	"Resent-Sender",
	"Resent-To",
	"Resent-Cc",
	"Resent-Bcc",
	"Resent-Message-Id",
	"Content-Transfer-Encoding",
	"Content-Type",
	"Content-Disposition",
}

// isExplicitHeader checks if the header is to be registered as a field.
// This slice search is much the same speed as a map lookup for small
// slices.
func isExplicitHeader(s string) bool {
	for _, e := range explicitHeaders {
		if e == s {
			return true
		}
	}
	return false
}

// idTrimCutset is the set of characters to trim around a message ID
const idTrimCutset string = "<> \n"

// parseAddresses parses a list of email addresses. Note that
// net/mail.Header[param] gets a list of addresses rather than slice.
func (se *stagedEmail) parseAddresses(s string) ([]*mail.Address, error) {
	if s == "" {
		return nil, errorEmptyAddress
	}
	addresses := []*mail.Address{}
	decodedHeader, err := decoders.DecodeHeader(s)
	if err != nil {
		return addresses, fmt.Errorf("cannot decode address %q: %w", s, err)
	}
	// plug point for custom address parsing
	return se.parser.addressesFunc(decodedHeader)
}

// parseAddress parses a single *mail.Address from a string using
// parseAddresses
func (se *stagedEmail) parseAddress(s string) (*mail.Address, error) {
	if s == "" {
		return nil, errorEmptyAddress
	}
	decodedHeader, err := decoders.DecodeHeader(s)
	if err != nil {
		return nil, fmt.Errorf("cannot decode address %q: %w", s, err)
	}
	// plug point for custom address parsing
	return se.parser.addressFunc(decodedHeader)
}

// parseHeaders parses the headers in the net/mail.Header at se.msg into
// se.email.Headers field values.
func (se *stagedEmail) parseHeaders() error {

	// get is a shortcut to net/mail.Header.Get, which returns the first
	// value (if any) for a header field. Note that all lists of email
	// addresses are returned as single string, so should be retrieved
	// using "Get" rather than by map lookup.
	get := func(field string) string {
		return se.msg.Header.Get(field)
	}

	// getAll is shortcut to get the net/mail.Header []string elements
	getAll := func(field string) []string {
		return se.msg.Header[field]
	}

	// getID returns a cleaned message id
	getID := func(s string) string { return strings.Trim(s, idTrimCutset) }

	// getIDs returns a slice of cleaned message ids
	getIDs := func(s string) []string {
		ids := []string{}
		for _, id := range strings.Split(s, " ") {
			id := strings.TrimSpace(strings.Trim(id, idTrimCutset))
			if id == "" {
				continue
			}
			ids = append(ids, id)
		}
		return ids
	}

	callDateFunc := func(s string) (time.Time, error) {
		if s == "" {
			return time.Time{}, errorEmptyDate
		}
		// plug point for custom address parsing
		return se.parser.dateFunc(s)
	}

	// getDecodedString decodes and trims a string header
	getDecodedString := func(s string) (string, error) {
		return decoders.DecodeHeader(strings.TrimSpace(s))
	}

	// getCSV gets parts of a comma delimited string
	getCSV := func(s string) []string {
		o := []string{}
		parts := strings.Split(s, ",")
		for _, pa := range parts {
			pp := strings.TrimSpace(pa)
			if len(pp) > 0 {
				o = append(o, pp)
			}
		}
		return o
	}

	// alias headers for easy reference
	h := &se.email.Headers

	// set contentInfo from stagedEmail
	h.ContentInfo = se.contentInfo

	h.ExtraHeaders = map[string][]string{}
	for key, value := range se.msg.Header {
		if isExplicitHeader(key) {
			continue
		}
		h.ExtraHeaders[key] = []string{}
		for _, val := range value {
			val, _ := decoders.DecodeHeader(val)
			h.ExtraHeaders[key] = append(h.ExtraHeaders[key], val)
		}
	}

	var err error
	if h.Sender, err = se.parseAddress(get("Sender")); err != nil {
		if !errors.Is(errorEmptyAddress, err) {
			return fmt.Errorf("cannot parse Sender header: %w", err)
		}
	}

	// Get email address lists via get. See get function comments.
	if h.From, err = se.parseAddresses(get("From")); err != nil {
		if !errors.Is(errorEmptyAddress, err) {
			return fmt.Errorf("From header: (%s) %w", get("From"), err)
		}
	}

	if h.ReplyTo, err = se.parseAddresses(get("Reply-To")); err != nil {
		if !errors.Is(errorEmptyAddress, err) {
			return fmt.Errorf("Reply-To header: (%s) %w", get("Reply-To"), err)
		}
	}

	if h.To, err = se.parseAddresses(get("To")); err != nil {
		if !errors.Is(errorEmptyAddress, err) {
			return fmt.Errorf("To header: (%s) %w", get("To"), err)
		}
	}

	if h.Cc, err = se.parseAddresses(get("Cc")); err != nil {
		if !errors.Is(errorEmptyAddress, err) {
			return fmt.Errorf("Cc header: (%s) %w", get("Cc"), err)
		}
	}

	if h.Bcc, err = se.parseAddresses(get("Bcc")); err != nil {
		if !errors.Is(errorEmptyAddress, err) {
			return fmt.Errorf("Bcc header: (%s) %w", get("Bcc"), err)
		}
	}

	if h.ResentFrom, err = se.parseAddresses(get("Resent-From")); err != nil {
		if !errors.Is(errorEmptyAddress, err) {
			return fmt.Errorf("Resent-From header: (%s) %w", get("Resent-From"), err)
		}
	}

	if h.ResentSender, err = se.parseAddress(get("Resent-Sender")); err != nil {
		if !errors.Is(errorEmptyAddress, err) {
			return fmt.Errorf("Resent-Sender header: (%s) %w", get("Resent-Sender"), err)
		}
	}

	if h.ResentTo, err = se.parseAddresses(get("Resent-To")); err != nil {
		if !errors.Is(errorEmptyAddress, err) {
			return fmt.Errorf("Resent-To header: (%s) %w", get("Resent-To"), err)
		}
	}

	if h.ResentCc, err = se.parseAddresses(get("Resent-Cc")); err != nil {
		if !errors.Is(errorEmptyAddress, err) {
			return fmt.Errorf("Resent-Cc header: (%s) %w", get("Resent-Cc"), err)
		}
	}

	if h.ResentBcc, err = se.parseAddresses(get("Resent-Bcc")); err != nil {
		if !errors.Is(errorEmptyAddress, err) {
			return fmt.Errorf("Resent-Bcc header: (%s) %w", get("Resent-Bcc"), err)
		}
	}

	if h.Date, err = callDateFunc(get("Date")); err != nil {
		if !errors.Is(errorEmptyDate, err) {
			return fmt.Errorf("Date header: (%s) %w", get("Date"), err)
		}
	}

	if h.ResentDate, err = callDateFunc(get("Resent-Date")); err != nil {
		if !errors.Is(errorEmptyDate, err) {
			return fmt.Errorf("Resent-Date header: (%s) %w", get("Resent-Date"), err)
		}
	}

	if h.Subject, err = getDecodedString(get("Subject")); err != nil {
		return fmt.Errorf("Subject header: (%s) %w", get("Subject"), err)
	}

	if h.Comments, err = getDecodedString(get("Comments")); err != nil {
		return fmt.Errorf("Comments header: (%s) %w", get("Comments"), err)
	}

	// consider parsing this into []Received
	if re := getAll("Received"); len(re) > 0 {
		h.Received = re
	}

	if id := getID(get("Message-ID")); id != "" {
		h.MessageID = id
	}

	if ids := getIDs(get("In-Reply-To")); len(ids) > 0 {
		h.InReplyTo = ids
	}

	if ids := getIDs(get("References")); len(ids) > 0 {
		h.References = ids
	}

	if kw := getCSV(get("Keywords")); len(kw) > 0 {
		h.Keywords = kw
	}

	if id := getID(get("Resent-Message-ID")); id != "" {
		h.ResentMessageID = id
	}

	return nil
}
