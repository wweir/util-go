package log_test

import (
	"io"

	"github.com/wweir/util-go/log"
)

func Example_If() {
	var a = 1
	var err error

	defer log.ErrPt(&err).Warnw("log defer ErrPt", "a", &a)

	log.Infow("log", "a", a, "err", err)
	log.Err(err).Warnw("log Err", "a", a)

	a = 2
	err = io.EOF

	log.Err(err).Warnw("log Err", "a", a)
	//Output:
}
