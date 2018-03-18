package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const tagName = "db"

var (
	notReadableValueError     = errors.New("Value not addressable or interfaceable")
	notSettableError          = errors.New("Passed in variable is not settable")
	unsupportedValueTypeError = errors.New("Unsupported unmarshal type")
)

type Scanner interface {
	Columns() ([]string, error)
	Err() error
	Next() bool
	Scan(v ...interface{}) error
}

func UnmarshalRow(v interface{}, scanner Scanner) error {
	if !scanner.Next() {
		if err := scanner.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}

	rv := reflect.ValueOf(v)
	if err := validatePtr(&rv); err != nil {
		return err
	}

	rte := reflect.TypeOf(v).Elem()
	rve := rv.Elem()

	switch rte.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		if rve.CanSet() {
			return scanner.Scan(v)
		} else {
			return notSettableError
		}
	case reflect.Struct:
		columns, err := scanner.Columns()
		if err != nil {
			return err
		}

		if values, err := mapStructFieldsIntoSlice(rve, columns); err != nil {
			return err
		} else {
			return scanner.Scan(values...)
		}
	default:
		return unsupportedValueTypeError
	}
}

func UnmarshalRows(v interface{}, scanner Scanner) error {
	rv := reflect.ValueOf(v)
	if err := validatePtr(&rv); err != nil {
		return err
	}

	rt := reflect.TypeOf(v)
	rte := rt.Elem()
	rve := rv.Elem()
	switch rte.Kind() {
	case reflect.Slice:
		if rve.CanSet() {
			ptr := rte.Elem().Kind() == reflect.Ptr
			appendFn := func(item reflect.Value) {
				if ptr {
					rve.Set(reflect.Append(rve, item))
				} else {
					rve.Set(reflect.Append(rve, reflect.Indirect(item)))
				}
			}
			fillFn := func(value interface{}) error {
				if rve.CanSet() {
					if err := scanner.Scan(value); err != nil {
						return err
					} else {
						appendFn(reflect.ValueOf(value))
						return nil
					}
				} else {
					return notSettableError
				}
			}

			base := Deref(rte.Elem())
			switch base.Kind() {
			case reflect.Bool,
				reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
				reflect.Float32, reflect.Float64,
				reflect.String:
				for scanner.Next() {
					value := reflect.New(base)
					if err := fillFn(value.Interface()); err != nil {
						return err
					}
				}
			case reflect.Struct:
				columns, err := scanner.Columns()
				if err != nil {
					return err
				}

				for scanner.Next() {
					value := reflect.New(base)
					if values, err := mapStructFieldsIntoSlice(value, columns); err != nil {
						return err
					} else {
						if err := scanner.Scan(values...); err != nil {
							return err
						} else {
							appendFn(value)
						}
					}
				}
			default:
				return unsupportedValueTypeError
			}

			return nil
		} else {
			return notSettableError
		}
	default:
		return unsupportedValueTypeError
	}
}

func getTaggedFieldValueMap(v reflect.Value) (map[string]interface{}, error) {
	rt := Deref(v.Type())
	size := rt.NumField()
	result := make(map[string]interface{}, size)

	for i := 0; i < size; i++ {
		key := parseTagName(rt.Field(i))
		if len(key) == 0 {
			return nil, nil
		}

		valueField := reflect.Indirect(v).Field(i)
		switch valueField.Kind() {
		case reflect.Ptr:
			if !valueField.CanInterface() {
				return nil, notReadableValueError
			}
			if valueField.IsNil() {
				baseValueType := Deref(valueField.Type())
				valueField.Set(reflect.New(baseValueType))
			}
			result[key] = valueField.Interface()
		default:
			if !valueField.CanAddr() || !valueField.Addr().CanInterface() {
				return nil, notReadableValueError
			}
			result[key] = valueField.Addr().Interface()
		}
	}

	return result, nil
}

func mapStructFieldsIntoSlice(v reflect.Value, columns []string) ([]interface{}, error) {
	indirect := reflect.Indirect(v)

	taggedMap, err := getTaggedFieldValueMap(v)
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(columns))
	if taggedMap == nil || len(taggedMap) == 0 {
		for i := 0; i < len(values); i++ {
			valueField := indirect.Field(i)
			switch valueField.Kind() {
			case reflect.Ptr:
				if !valueField.CanInterface() {
					return nil, notReadableValueError
				}
				if valueField.IsNil() {
					baseValueType := Deref(valueField.Type())
					valueField.Set(reflect.New(baseValueType))
				}
				values[i] = valueField.Interface()
			default:
				if !valueField.CanAddr() || !valueField.Addr().CanInterface() {
					return nil, notReadableValueError
				}
				values[i] = valueField.Addr().Interface()
			}
		}
	} else {
		for i, column := range columns {
			if tagged, ok := taggedMap[column]; ok {
				values[i] = tagged
			} else {
				var anonymous interface{}
				values[i] = &anonymous
			}
		}
	}

	return values, nil
}

func parseTagName(field reflect.StructField) string {
	key := field.Tag.Get(tagName)
	if len(key) == 0 {
		return ""
	} else {
		options := strings.Split(key, ",")
		return options[0]
	}
}

func Deref(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func validatePtr(v *reflect.Value) error {
	// sequence is very important, IsNil must be called after checking Kind() with reflect.Ptr,
	// panic otherwise
	if !v.IsValid() || v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("Error: not a valid pointer: %v", v)
	}

	return nil
}
