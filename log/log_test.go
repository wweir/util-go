package log_test

import (
	"errors"

	"github.com/wweir/util-go/log"
)

func Example_Check() {
	log.Check(nil).Infow("123")
	log.Check(errors.New("")).Infow("456")
	//Output:
}
