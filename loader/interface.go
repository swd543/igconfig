package loader

import (
	"context"
	"reflect"
	"time"
)

type DynamicRunner func(value []byte) error

type DynamicConfig struct {
	AppName string
	// FieldName that should be looked up.
	//
	// This is not exactly field name, rather how it will be named in Consul.
	// For example in Consul path /finops/adm0001s/loglevel FieldName MUST be set to "loglevel".
	FieldName       string
	RefreshInterval time.Duration
	// Runner will be executed when new, different value will be received.
	//
	// Returning error will stop dynamic updates, so it should be restarted manually.
	Runner DynamicRunner
}

type Loader interface {
	// Load will load all available data to at 'to' value.
	//
	// Even if particular loader type must implement ReflectLoader -
	// this interface still must be implemented as a proxy.
	Load(appName string, to interface{}) error
}

type ReflectLoader interface {
	// Load will load all available data to at 'to' value.
	// 'to' is reflect.Value as it will be easier to get field information from it.
	//
	// If unmarshaling is needed it can be done as
	//  err := json.Unmarshal(data, to.Interface())
	LoadReflect(appName string, to reflect.Value) error
}

type DynamicValuer interface {
	// DynamicValue polls config field value once in dynamicConfig.RefreshInterval.
	// If after poll value was changed - dynamicConfig.Runner function will be called with new value.
	//
	// If requested service supports better(native) listening for changes - it will be implemented instead.
	//
	// Error handling should be done in runner function.
	DynamicValue(ctx context.Context, dynamicConfig DynamicConfig) error
}