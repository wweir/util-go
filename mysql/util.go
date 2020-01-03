package mysql

import "fmt"

// EnumToGo is mysql ENUM value to golang value
func EnumToGo(enum interface{}) uint16 {
	num := enum.(uint16)
	return num - 1
}

// GoToEnum is golang value to mysql ENUM value
func GoToEnum(enum interface{}) uint16 {
	switch num := enum.(type) {
	case uint:
		return uint16(num) + 1
	case uint8:
		return uint16(num) + 1
	case uint16:
		return uint16(num) + 1
	case uint32:
		return uint16(num) + 1
	case uint64:
		return uint16(num) + 1
	case int:
		return uint16(num) + 1
	case int8:
		return uint16(num) + 1
	case int16:
		return uint16(num) + 1
	case int32:
		return uint16(num) + 1
	case int64:
		return uint16(num) + 1
	case interface{ EnumDescriptor() ([]byte, []int) }: //grpc
		_, nums := num.EnumDescriptor()
		return uint16(nums[0]) + 1
	default:
		panic(fmt.Sprintf("invalid type %T", enum))
	}
}
