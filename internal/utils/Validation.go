package utils

import (
	"log"
	"reflect"
)

func CheckEmptyFields(info interface{}) {
	v := reflect.ValueOf(info)
	t := reflect.TypeOf(info)

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldName := t.Field(i).Name
		if fieldValue.Kind() == reflect.String && fieldValue.String() == "" {
			log.Printf("%s tidak ditemukan.\n", fieldName)
		}
	}
}
