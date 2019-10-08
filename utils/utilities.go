package utils

import (
	"reflect"
	"strings"
)

// ToCamelCase converts a string to camel case. The string 'FirstName' becomes 'firstName'
func ToCamelCase(str string) string {
	if len(str) == 0 {
		return str
	}

	first := strings.ToLower(string(str[0]))

	return first + str[1:]
}

// StructToMap converts any value into a map where the keys are field names
func StructToMap(v interface{}) map[string]interface{} {

	fieldMap := make(map[string]interface{})

	structType := reflect.TypeOf(v)
	structValue := reflect.ValueOf(v)

	for i := 0; i < structType.NumField(); i++ {
		fieldMap[ToCamelCase(structType.Field(i).Name)] = structValue.Field(i).Interface()
	}

	return fieldMap
}
