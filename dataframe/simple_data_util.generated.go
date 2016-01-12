package dataframe

import "fmt"

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
