package verboten

import (
	"github.com/reiver/go-erorr"
)

const (
	errNilEventWriter        = erorr.Error("nil event-writer")
	errNilHTTPSEEEventWriter = erorr.Error("nil HTTP-SSE event-writer")
)
