module cima

go 1.26

require (
	github.com/ovav/ovav/go-runtime/cmd/sheets v0.0.0
)

// replace directive points CIMA to the local OVAV repo for the sheets bridge.
// This is a runtime dependency — CIMA consumes OVAV's sheets implementation
// as a library, not a fork.
replace github.com/ovav/ovav/go-runtime/cmd/sheets => /home/braka/Systems/ovav/go-runtime/cmd/sheets