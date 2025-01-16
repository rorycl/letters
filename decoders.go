package letters

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"strings"

	"github.com/mnako/letters/base64toraw"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

func decodeHeader(s string) (string, error) {
	CharsetReader := func(label string, input io.Reader) (io.Reader, error) {
		enc, _ := charset.Lookup(label)
		if enc == nil {
			normalizedLabel := strings.Replace(label, "windows-", "cp", -1)
			enc, _ = charset.Lookup(normalizedLabel)
		}
		if enc == nil {
			return nil, fmt.Errorf(
				"letters.decoders.decodeHeader.CharsetReader: cannot lookup encoding %s",
				label)
		}
		return enc.NewDecoder().Reader(input), nil
	}
	mimeDecoder := mime.WordDecoder{CharsetReader: CharsetReader}
	decodedHeader, err := mimeDecoder.DecodeHeader(s)
	if err != nil {
		return decodedHeader, fmt.Errorf(
			"letters.decoders.decodeHeader: cannot decode MIME-word-encoded header %q: %w",
			s,
			err)
	}
	return decodedHeader, nil
}

// decodeContent wraps the content io.Reader (from a mime/multipart.Part) in either
// a base64 or quoted printable decoder if applicable. The function further wraps
// the reader in a transform character decoder if an encoding is supplied.
//
// Note that the base64 decoder "base64toraw.NewBase64ToRaw" decodes all base64
// content to data that is base64.RawStdEncoding encoded, i.e. without "=" padding.
func decodeContent(
	content io.Reader, e encoding.Encoding, cte ContentTransferEncoding,
) io.Reader {
	var contentReader io.Reader

	switch cte {
	case cteBase64:
		contentReader = base64.NewDecoder(base64.RawStdEncoding, base64toraw.NewBase64ToRaw(content))
	case cteQuotedPrintable:
		contentReader = quotedprintable.NewReader(content)
	default:
		contentReader = content
	}
	if e == nil {
		return contentReader
	}
	return transform.NewReader(contentReader, e.NewDecoder())
}

func decodeInlineFile(part *multipart.Part, cte ContentTransferEncoding) (InlineFile, error) {
	var ifl InlineFile

	cid, err := decodeHeader(part.Header.Get("Content-Id"))
	if err != nil {
		return ifl, fmt.Errorf(
			"letters.decoders.decodeInlineFile: cannot decode Content-ID header for inline attachment: %w",
			err)
	}

	ifl.ContentID = strings.Trim(cid, "<>")

	ifl.ContentType, err = parseContentTypeHeader(part.Header.Get("Content-Type"))
	if err != nil {
		return ifl, fmt.Errorf(
			"letters.decoders.decodeInlineFile: cannot parse Content-Type of inline attachment: %w",
			err)
	}

	ifl.ContentDisposition, err = parseContentDisposition(part.Header.Get("Content-Disposition"))
	if err != nil {
		return ifl, fmt.Errorf(
			"letters.decoders.decodeInlineFile: cannot parse Content-Disposition of inline attachment: %w",
			err)
	}

	ifl.setDefaultWriterFunc()
	err = ifl.Write(decodeContent(part, nil, cte))

	return ifl, err
}

func decodeAttachmentFileFromBody(body io.Reader, headers Headers, cte ContentTransferEncoding) (AttachedFile, error) {
	var afl AttachedFile

	afl.ContentType = headers.ContentType
	afl.ContentDisposition = headers.ContentDisposition

	afl.setDefaultWriterFunc()
	err := afl.Write(decodeContent(body, nil, cte))

	return afl, err
}

func decodeAttachedFileFromPart(part *multipart.Part, cte ContentTransferEncoding) (AttachedFile, error) {
	var afl AttachedFile

	var err error
	afl.ContentType, err = parseContentTypeHeader(part.Header.Get("Content-Type"))
	if err != nil {
		return afl, fmt.Errorf(
			"letters.decoders.decodeAttachedFileFromPart: cannot parse Content-Type of attached file: %w",
			err)
	}

	afl.ContentDisposition, err = parseContentDisposition(part.Header.Get("Content-Disposition"))
	if err != nil {
		return afl, fmt.Errorf(
			"letters.decoders.decodeAttachedFileFromPart: cannot parse Content-Disposition of attached file: %w",
			err)
	}

	afl.setDefaultWriterFunc()
	err = afl.Write(decodeContent(part, nil, cte))

	return afl, err
}
