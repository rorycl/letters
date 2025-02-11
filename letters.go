// # Letters, an email parsing package for go
//
// This is a fork of [mnako/letters](https://github.com/mnako/letters), a
// minimalistic Golang library for parsing plaintext and MIME emails.
//
// Thanks to @mnako and contributors, letters has great support for
// languages other than English, text encodings and transfer-encodings.
//
// This fork focuses on performance, memory efficiency and extensibility
// through modularisation.
//
// ## Quickstart
//
// # Install
//
// ```
// go get github.com/rorycl/letters@latest
// ```
//
// Parse an email:
//
//	```go
//	p := letters.NewParser()
//	parsedEmail, err := p.Parse(reader)
//
//	// &email.Email{
//	// 	Headers: email.Headers{
//	// 		Date: time.Time(time.Date(2019, 4, 1, 0, 55, 0, 0, time.UTC)),
//	// 		Sender: &mail.Address{
//	// 			Name:    "อลิซ ผู้ส่งจดหมาย",
//	// 			Address: "alis.phusngcdhmay@example.com",
//	// 		},
//	// 		From: []*mail.Address{
//	// 			&mail.Address{
//	// 				Name:    "อลิซ ผู้ส่งจดหมาย",
//	// 				Address: "alis.phusngcdhmay@example.com",
//	// 			},
//	// 		},
//	// 		To: []*mail.Address{
//	// 			&mail.Address{
//	// 				Name:    "บ๊อบ ผู้รับ",
//	// 				Address: "bob.phurab@example.com",
//	// 			},
//	// 		},
//	// 		Cc: []*mail.Address{
//	// 			&mail.Address{
//	// 				Name:    "แดน ผู้รับ",
//	// 				Address: "dan.phurab@example.com",
//	// 			},
//	// 		},
//	// 		MessageID: "Message-Id-1@example.com",
//	// 		InReplyTo: []string{
//	// 			"Message-Id-0@example.com",
//	// 		},
//	// 		References: []string{
//	// 			"Message-Id-0@example.com",
//	// 		},
//	// 		Subject:  "📧 Test แพนแกรมภาษาไทย",
//	// 		Comments: "Message Header Comment",
//	// 		Keywords: []string{
//	// 			"Keyword 1",
//	// 			"Keyword 2",
//	// 		},
//	// 		ResentDate: time.Time(time.Date(2019, 4, 1, 0, 55, 0, 0, time.UTC)),
//	// 		ExtraHeaders: map[string][]string{
//	// 			"X-Clacks-Overhead": []string{
//	// 				"GNU Terry Pratchett",
//	// 			},
//	// 		},
//	// 		ContentInfo: &email.ContentInfo{
//	// 			Type: "multipart/mixed",
//	// 			TypeParams: map[string]string{
//	// 				"boundary": "MixedBoundaryString",
//	// 				"charset":  "tis-620",
//	// 			},
//	// 			Disposition:       "",
//	// 			DispositionParams: map[string]string(nil), // p0
//	// 			TransferEncoding:  "base64",
//	// 			ID:                "",
//	// 			Charset:           "tis-620",
//	// 		},
//	// 		Received: nil,
//	// 	},
//	// 	Text: ""
//	// 	EnrichedText: "<bold>เป็นมนุษย์สุดประเสริฐเลิศคุณค่า</bold> ..."
//	// 	HTML: ""
//	// 	Files: []*email.File{
//	// 		&email.File{
//	// 			FileType: "inline",
//	// 			Name:     "inline-jpg-image-without-disposition.jpg",
//	// 			ContentInfo: &email.ContentInfo{
//	// 				Type: "image/jpeg",
//	// 				TypeParams: map[string]string{
//	// 					"name": "inline-jpg-image-without-disposition.jpg",
//	// 				},
//	// 				Disposition:       "",
//	// 				DispositionParams: map[string]string(nil), // p0
//	// 				TransferEncoding:  "base64",
//	// 				ID:                "",
//	// 				Charset:           "tis-620",
//	// 			},
//	// 			Data: []byte{
//	// 				239, 191, 189, 224, 184, 184, 239, 191, 189, ...
//	// 			},
//	// 		},
//	// 		},
//	// 		&email.File{
//	// 			FileType: "attachment",
//	// 			Name:     "attached-pdf-filename.pdf",
//	// 			ContentInfo: &email.ContentInfo{
//	// 				Type: "application/pdf",
//	// 				TypeParams: map[string]string{
//	// 					"name": "attached-pdf-name.pdf",
//	// 				},
//	// 				Disposition: "attachment",
//	// 				DispositionParams: map[string]string{
//	// 					"filename": "attached-pdf-filename.pdf",
//	// 				},
//	// 				TransferEncoding: "base64",
//	// 				ID:               "",
//	// 				Charset:          "tis-620",
//	// 			},
//	// 			Data: []byte{
//	// 				37, 80, 68, 70, 45, 49, 46, 13, 116, 114, 97, ...
//	// 			},
//	// 		},
//	// }
//	```
//
// # Options
//
// Various options are provided for customising the Parser, including:
//
//	// skip content types
//	func WithSkipContentTypes(skipContentTypes []string) Opt
//	// provide a custom address processing function
//	func WithCustomAddressFunc(af func(string) (*mail.Address, error)) Opt
//	// provide a custom processing function for string lists of addresses
//	func WithCustomAddressesFunc(af func(list string) ([]*mail.Address, error)) Opt
//	// provide a custom date processing function
//	func WithCustomDateFunc(df func(string) (time.Time, error)) Opt
//	// provide a custom file processing function
//	func WithCustomFileFunc(ff func(*email.File) error) Opt
//	// save files to the stated directory (an example of WithCustomFileFunc)
//	func WithSaveFilesToDirectory(dir string) Opt
//	// only process headers
//	func WithHeadersOnly() Opt
//	// skip processing attachments
//	func WithoutAttachments() Opt
//	// show verbose processing info (currently a noop)
//	func WithVerbose() Opt
//
// The `WithoutAttachments` and `WithHeadersOnly` options determine if
// only part of an email will be processed.
//
// The `WithSkipContentTypes` allows the user to skip processing MIME
// message parts with the supplied content-types.
//
// The date and address "With" options allow the provision of custom
// funcs to override the [net/mail] funcs normally used. For example it
// might be necessary to extend the date parsing capabilities to deal
// with poorly formatted date strings produced by older SMTP servers.
//
// The `WithCustomFileFunc` allows the provision of a custom func for
// saving, filtering and/or processing of inline or attached files
// without reading them first into an `email.File.Data` []byte slice
// first, which is the default behaviour. The `WithSaveFilesToDirectory`
// option is an example of such a custom func.
//
// As shown in the [parser/optspkg_test.go](parser/optspkg_test.go)
// package test, `WithCustomFileFunc` can be used to, for example, only
// process `image/jpeg` files. More examples are shown in
// [parser/opts_test.go](parser/opts_test.go), for example:
//
//	opt := parser.WithHeadersOnly() // the headers only option
//	p := letters.NewParser(opt, parser.WithVerbose()) // options can be chained
//	parsedEmail, err := p.Parse(emailReader)
//	if err != nil {
//		return fmt.Errorf("error while parsing email headers: %s", err)
//	}
package letters

import (
	"github.com/rorycl/letters/parser"
)

func NewParser(options ...parser.Opt) *parser.Parser {
	return parser.NewParser(options...)
}
