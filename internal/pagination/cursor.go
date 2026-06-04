package pagination

import (
	"encoding/base64"
	"errors"
	"strings"
)

var ErrInvalidCursor = errors.New("invalid cursor")

const separator = "\x1f"

func EncodeCursor(parts ...string) string {
	joined := strings.Join(parts, separator)
	return base64.RawURLEncoding.EncodeToString([]byte(joined))
}

func DecodeCursor(raw string) ([]string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, ErrInvalidCursor
	}

	return strings.Split(string(decoded), separator), nil
}
