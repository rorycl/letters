module main

go 1.22.5

replace github.com/rorycl/letters => ../

replace github.com/rorycl/letters/email => ../email/

require (
	github.com/rorycl/base64toraw v0.0.1 // indirect
	github.com/rorycl/letters v0.0.1 // indirect
	github.com/sanity-io/litter v1.5.6 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)
