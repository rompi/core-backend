package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// binder handles binding configuration values to struct fields.
type binder struct {
	values       map[string]any
	keyDelimiter string
}

// newBinder creates a new binder instance.
func newBinder(values map[string]any, keyDelimiter string) *binder {
	return &binder{
		values:       values,
		keyDelimiter: keyDelimiter,
	}
}

// Bind binds configuration values to a struct.
func (b *binder) Bind(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("bind target must be a non-nil pointer")
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("bind target must be a pointer to struct")
	}

	errors := &MultiBindError{}
	b.bindStruct(rv, "", errors)

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// bindStruct recursively binds values to a struct.
func (b *binder) bindStruct(rv reflect.Value, prefix string, errors *MultiBindError) {
	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fieldVal := rv.Field(i)

		// Skip unexported fields
		if !fieldVal.CanSet() {
			continue
		}

		// Get config key from tag
		tag := field.Tag.Get("config")
		if tag == "-" {
			continue
		}

		// Use field name if no tag
		if tag == "" {
			tag = strings.ToLower(field.Name)
		}

		// Build full key
		key := tag
		if prefix != "" {
			key = prefix + b.keyDelimiter + tag
		}

		// Handle nested structs
		if field.Type.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{}) {
			b.bindStruct(fieldVal, key, errors)
			continue
		}

		// Handle pointer to struct
		if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			if fieldVal.IsNil() {
				fieldVal.Set(reflect.New(field.Type.Elem()))
			}
			b.bindStruct(fieldVal.Elem(), key, errors)
			continue
		}

		// Get value from config
		value := b.getValue(key)

		// Check for environment variable override
		if envKey := field.Tag.Get("env"); envKey != "" {
			if envVal := b.getValue(envKey); envVal != nil {
				value = envVal
			}
		}

		// Apply default if value is nil
		if value == nil {
			if defaultVal := field.Tag.Get("default"); defaultVal != "" {
				value = defaultVal
			}
		}

		// Check required
		if value == nil {
			if required := field.Tag.Get("required"); required == "true" {
				errors.Add(BindError{
					Field:   field.Name,
					Tag:     "required",
					Message: fmt.Sprintf("required field %q is missing", key),
				})
			}
			continue
		}

		// Set the value
		if err := b.setFieldValue(fieldVal, value); err != nil {
			errors.Add(BindError{
				Field:   field.Name,
				Tag:     "config",
				Value:   value,
				Message: fmt.Sprintf("failed to set field %q: %v", field.Name, err),
			})
		}
	}
}

// getValue retrieves a value from the config map.
func (b *binder) getValue(key string) any {
	return b.values[strings.ToLower(key)]
}

// setFieldValue sets a struct field to the given value.
func (b *binder) setFieldValue(field reflect.Value, value any) error {
	if value == nil {
		return nil
	}

	fieldType := field.Type()

	// Handle pointer types
	if fieldType.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(fieldType.Elem()))
		}
		return b.setFieldValue(field.Elem(), value)
	}

	// Handle time.Duration
	if fieldType == reflect.TypeOf(time.Duration(0)) {
		return b.setDuration(field, value)
	}

	// Handle time.Time
	if fieldType == reflect.TypeOf(time.Time{}) {
		return b.setTime(field, value)
	}

	// Handle slices
	if fieldType.Kind() == reflect.Slice {
		return b.setSlice(field, value)
	}

	// Handle maps
	if fieldType.Kind() == reflect.Map {
		return b.setMap(field, value)
	}

	// Handle basic types
	switch fieldType.Kind() {
	case reflect.String:
		return b.setString(field, value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return b.setInt(field, value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return b.setUint(field, value)
	case reflect.Float32, reflect.Float64:
		return b.setFloat(field, value)
	case reflect.Bool:
		return b.setBool(field, value)
	default:
		// Try direct assignment
		rv := reflect.ValueOf(value)
		if rv.Type().AssignableTo(fieldType) {
			field.Set(rv)
			return nil
		}
		return fmt.Errorf("unsupported field type: %v", fieldType)
	}
}

func (b *binder) setString(field reflect.Value, value any) error {
	switch v := value.(type) {
	case string:
		field.SetString(v)
	case fmt.Stringer:
		field.SetString(v.String())
	default:
		field.SetString(fmt.Sprintf("%v", value))
	}
	return nil
}

func (b *binder) setInt(field reflect.Value, value any) error {
	var i int64

	switch v := value.(type) {
	case int:
		i = int64(v)
	case int8:
		i = int64(v)
	case int16:
		i = int64(v)
	case int32:
		i = int64(v)
	case int64:
		i = v
	case uint:
		i = int64(v)
	case uint8:
		i = int64(v)
	case uint16:
		i = int64(v)
	case uint32:
		i = int64(v)
	case uint64:
		i = int64(v)
	case float32:
		i = int64(v)
	case float64:
		i = int64(v)
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert %q to int: %w", v, err)
		}
		i = parsed
	default:
		return fmt.Errorf("cannot convert %T to int", value)
	}

	if field.OverflowInt(i) {
		return fmt.Errorf("value %d overflows %v", i, field.Type())
	}

	field.SetInt(i)
	return nil
}

func (b *binder) setUint(field reflect.Value, value any) error {
	var u uint64

	switch v := value.(type) {
	case uint:
		u = uint64(v)
	case uint8:
		u = uint64(v)
	case uint16:
		u = uint64(v)
	case uint32:
		u = uint64(v)
	case uint64:
		u = v
	case int:
		if v < 0 {
			return fmt.Errorf("cannot convert negative value to uint")
		}
		u = uint64(v)
	case int64:
		if v < 0 {
			return fmt.Errorf("cannot convert negative value to uint")
		}
		u = uint64(v)
	case float64:
		if v < 0 {
			return fmt.Errorf("cannot convert negative value to uint")
		}
		u = uint64(v)
	case string:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert %q to uint: %w", v, err)
		}
		u = parsed
	default:
		return fmt.Errorf("cannot convert %T to uint", value)
	}

	if field.OverflowUint(u) {
		return fmt.Errorf("value %d overflows %v", u, field.Type())
	}

	field.SetUint(u)
	return nil
}

func (b *binder) setFloat(field reflect.Value, value any) error {
	var f float64

	switch v := value.(type) {
	case float32:
		f = float64(v)
	case float64:
		f = v
	case int:
		f = float64(v)
	case int64:
		f = float64(v)
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("cannot convert %q to float: %w", v, err)
		}
		f = parsed
	default:
		return fmt.Errorf("cannot convert %T to float", value)
	}

	if field.OverflowFloat(f) {
		return fmt.Errorf("value %f overflows %v", f, field.Type())
	}

	field.SetFloat(f)
	return nil
}

func (b *binder) setBool(field reflect.Value, value any) error {
	switch v := value.(type) {
	case bool:
		field.SetBool(v)
	case string:
		field.SetBool(parseBool(v))
	case int:
		field.SetBool(v != 0)
	case int64:
		field.SetBool(v != 0)
	default:
		return fmt.Errorf("cannot convert %T to bool", value)
	}
	return nil
}

func (b *binder) setDuration(field reflect.Value, value any) error {
	var d time.Duration

	switch v := value.(type) {
	case time.Duration:
		d = v
	case string:
		parsed, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("cannot parse duration %q: %w", v, err)
		}
		d = parsed
	case int:
		d = time.Duration(v)
	case int64:
		d = time.Duration(v)
	case float64:
		d = time.Duration(v)
	default:
		return fmt.Errorf("cannot convert %T to duration", value)
	}

	field.Set(reflect.ValueOf(d))
	return nil
}

func (b *binder) setTime(field reflect.Value, value any) error {
	var t time.Time

	switch v := value.(type) {
	case time.Time:
		t = v
	case string:
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02",
			"2006-01-02 15:04:05",
		}
		var err error
		for _, format := range formats {
			t, err = time.Parse(format, v)
			if err == nil {
				break
			}
		}
		if err != nil {
			return fmt.Errorf("cannot parse time %q", v)
		}
	default:
		return fmt.Errorf("cannot convert %T to time", value)
	}

	field.Set(reflect.ValueOf(t))
	return nil
}

func (b *binder) setSlice(field reflect.Value, value any) error {
	rv := reflect.ValueOf(value)

	// Handle string -> slice conversion (comma-separated)
	if s, ok := value.(string); ok {
		parts := strings.Split(s, ",")
		slice := reflect.MakeSlice(field.Type(), len(parts), len(parts))

		for i, part := range parts {
			part = strings.TrimSpace(part)
			elem := slice.Index(i)
			if err := b.setFieldValue(elem, part); err != nil {
				return fmt.Errorf("cannot set slice element %d: %w", i, err)
			}
		}

		field.Set(slice)
		return nil
	}

	// Handle slice -> slice conversion
	if rv.Kind() == reflect.Slice {
		slice := reflect.MakeSlice(field.Type(), rv.Len(), rv.Len())

		for i := 0; i < rv.Len(); i++ {
			elem := slice.Index(i)
			if err := b.setFieldValue(elem, rv.Index(i).Interface()); err != nil {
				return fmt.Errorf("cannot set slice element %d: %w", i, err)
			}
		}

		field.Set(slice)
		return nil
	}

	return fmt.Errorf("cannot convert %T to slice", value)
}

func (b *binder) setMap(field reflect.Value, value any) error {
	rv := reflect.ValueOf(value)

	if rv.Kind() != reflect.Map {
		return fmt.Errorf("cannot convert %T to map", value)
	}

	mapType := field.Type()
	newMap := reflect.MakeMap(mapType)

	iter := rv.MapRange()
	for iter.Next() {
		k := iter.Key()
		v := iter.Value()

		// Convert key
		keyVal := reflect.New(mapType.Key()).Elem()
		if err := b.setFieldValue(keyVal, k.Interface()); err != nil {
			return fmt.Errorf("cannot convert map key: %w", err)
		}

		// Convert value
		valVal := reflect.New(mapType.Elem()).Elem()
		if err := b.setFieldValue(valVal, v.Interface()); err != nil {
			return fmt.Errorf("cannot convert map value: %w", err)
		}

		newMap.SetMapIndex(keyVal, valVal)
	}

	field.Set(newMap)
	return nil
}
