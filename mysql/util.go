package mysql

// EnumToGo is mysql ENUM value to golang value
func EnumToGo(enum interface{}) uint16 {
	num := enum.(uint16)
	return num - 1
}

// GoToEnum is golang value to mysql ENUM value
func GoToEnum(enum interface{}) uint16 {
	num := enum.(uint16)
	return num + 1
}
