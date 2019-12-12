package util

import (
	"errors"
	"fmt"
)

func Example_MergeErr() {
	err := MergeErr(errors.New("1"), errors.New("1"), errors.New("2"))

	fmt.Println(err)
}
