package gokasa

const (
	initial_key byte = 171
)

func rotorCommand(command, output []byte) error {
	key := initial_key
	if len(output) < len(command) {
		return ErrOutputTooShort
	}

	for i, b := range command {
		a := key ^ b
		output[i] = a
		key = a
	}

	return nil
}

func derotorCommand(command []byte) []byte {
	key := initial_key
	output := make([]byte, len(command))

	for i, b := range command {
		a := key ^ b
		output[i] = a
		key = b
	}

	return output
}
