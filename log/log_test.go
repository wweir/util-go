package log_test

import (
	"errors"
	"fmt"

	"github.com/wweir/util-go/log"
)

func Example_If() {
	a := 1
	err := errors.New("error")

	fmt.Println(log.If == nil)
	defer log.If(err != nil).Errorw("log defer if", "a", &a, "err", err)

	log.Infow("log", "a", a, "err", err)
	a = 2
	log.If(err != nil).Errorw("log if", "a", a, "err", err)
}
