// Package protocol holds the names that the installer and the key service share.
package protocol

// Label is the container label with the protocol name.
// The installer puts it on each protocol container,
// and the key service finds the containers by it.
const Label = "jao.protocol"

// Protocol names, the values of Label.
const (
	AWG = "awg"
)
