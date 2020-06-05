package log_test

import (
	"io"

	"github.com/wweir/util-go/log"
)

func Example_If() {
	var a = 1
	var err error

	defer log.NotNil(&err).Warnw("log defer NotNil", "a", &a)

	log.Infow("log", "a", a, "err", err)
	log.NotNil(&err).Warnw("log NotNil", "a", a)

	a = 2
	err = io.EOF

	log.NotNil(&err).Warnw("log NotNil", "a", a)
	//Output:
}
