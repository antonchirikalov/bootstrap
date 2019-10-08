package utils

import "reflect"

// IncludesAnyStr looks for the presence of any in a set of needles
// for a given haystack
func IncludesAnyStr(needles []string, haystack []string) bool {
	for _, needle := range needles {
		if ok, _ := InArray(needle, haystack); ok {
			return true
		}
	}

	return false
}

// IncludesAnyInt looks for the presence of any in a set of needles
// for a given haystack
func IncludesAnyInt(needles []int64, haystack []int64) bool {
	for _, needle := range needles {
		if ok, _ := InArray(needle, haystack); ok {
			return true
		}
	}

	return false
}

// InArray is a generic in_array checker
func InArray(v interface{}, in interface{}) (ok bool, i int) {
	haystack := reflect.Indirect(reflect.ValueOf(in))
	switch haystack.Kind() {
	case reflect.Slice, reflect.Array:
		for ; i < haystack.Len(); i++ {
			if ok = v == haystack.Index(i).Interface(); ok {
				return
			}
		}
	}
	return
}
