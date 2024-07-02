package assembler

import "strings"

func Assemble(source string) ([]byte, error) {
	lines := strings.Split(source, "\n") // Split the source code into lines
	var bytecode []byte
	labels := make(map[string]byte) // Map of labels to their line numbers (labels[line] = address)
	address := byte(0)

	// First pass: collect labels
	for _, line := range lines {
		line = strings.TrimSpace(line)                  // Remove leading and trailing whitespace
		if line == "" || strings.HasPrefix(line, ";") { // Skip empty lines and comments (comments start with a semicolon)
			continue
		}
		if strings.HasSuffix(line, ":") {
			label := strings.TrimSuffix(line, ":")
			labels[label] = address
		} else {
			address++
		}
	}
}
