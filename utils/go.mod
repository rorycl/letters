module main

go 1.22.5

replace github.com/rorycl/letters => ../

replace github.com/rorycl/letters/email => ../email/

require (
	github.com/google/go-cmp v0.7.0
	github.com/rorycl/letters v0.1.2
	github.com/sanity-io/litter v1.5.8
)

require (
	github.com/rorycl/base64toraw v0.0.1 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)
