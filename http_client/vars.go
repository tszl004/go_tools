package http_client

import (
	"errors"
)

var (
	NotRedirectErr = errors.New("It doesn't follow redirect")
)
