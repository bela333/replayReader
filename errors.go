package replayReader

import "errors"

var (
	VarIntTooBigError = errors.New("VarInt is too big")
)
