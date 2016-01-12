package dataframe

import "fmt"

// IsNumeric checks whether x is a numeric type. Currently these
// consist only of integers and floats. Complex numbers and big
// numbers are not considered numeric.
func IsNumeric(x SimpleData) bool {
	switch x.(type) {
	case int: // numeric
	case int8: // numeric
	case int16: // numeric
	case int32: // numeric
	case int64: // numeric
	case uint: // numeric
	case uint8: // numeric
	case uint16: // numeric
	case uint32: // numeric
	case uint64: // numeric
	case uintptr: // numeric
	case float32: // numeric
	case float64: // numeric
	default:
		return false
	}
	return true
}

func ensureSimpleData(x SimpleData) {
	switch x.(type) {
	case string: // no error
	case bool: // no error
	case int: // no error
	case int8: // no error
	case int16: // no error
	case int32: // no error
	case int64: // no error
	case uint: // no error
	case uint8: // no error
	case uint16: // no error
	case uint32: // no error
	case uint64: // no error
	case uintptr: // no error
	case float32: // no error
	case float64: // no error
	default:
		panic(fmt.Sprintf("%s is not a valid SimpleData value", x))
	}
}

func slicePtrOf(dtype string) (interface{}, error) {
	switch dtype {
	case "empty":
		return new([]SimpleData), nil
	case "string":
		return new([]string), nil
	case "bool":
		return new([]bool), nil
	case "int":
		return new([]int), nil
	case "int8":
		return new([]int8), nil
	case "int16":
		return new([]int16), nil
	case "int32":
		return new([]int32), nil
	case "int64":
		return new([]int64), nil
	case "uint":
		return new([]uint), nil
	case "uint8":
		return new([]uint8), nil
	case "uint16":
		return new([]uint16), nil
	case "uint32":
		return new([]uint32), nil
	case "uint64":
		return new([]uint64), nil
	case "uintptr":
		return new([]uintptr), nil
	case "float32":
		return new([]float32), nil
	case "float64":
		return new([]float64), nil
	}
	return nil, fmt.Errorf("invalid data type %q for SimpleData", dtype)
}
