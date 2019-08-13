package igconfig

import (
	"os"
	"reflect"
	"strings"
)

// loadEnv loads config values from the environment
func (m *localData) loadEnv() {
	t := m.userStruct.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !m.testEnv(field.Name, m.userStruct.FieldByName(field.Name), field.Name) {
			nn := strings.Split(field.Tag.Get("env"), ",")
			for _, n := range nn {
				if m.testEnv(field.Name, m.userStruct.FieldByName(field.Name), n) {
					break
				}
			}
		}
	}
}

// testEnv tests for an environment variable, and if found sets the field's value
func (m *localData) testEnv(fieldName string, val reflect.Value, n string) bool {
	if v, ok := os.LookupEnv(n); ok {
		m.setValue(fieldName, v)
		return true
	}
	if v, ok := os.LookupEnv(strings.ToUpper(n)); ok {
		m.setValue(fieldName, v)
		return true
	}
	if v, ok := os.LookupEnv(strings.ToLower(n)); ok {
		m.setValue(fieldName, v)
		return true
	}
	return false
}
