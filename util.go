package util

func If(ok bool, fn func()) {
	if ok {
		fn()
	}
}
