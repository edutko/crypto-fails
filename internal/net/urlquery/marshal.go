package urlquery

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

var (
	ErrStructExpected       = errors.New("requires a struct")
	ErrStructPtrExpected    = errors.New("requires pointer to struct")
	ErrUnsupportedFieldType = errors.New("unsupported field type")
)

func Marshal(v interface{}) ([]byte, error) {
	vv := reflect.ValueOf(v)
	if vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
	}
	if vv.Kind() != reflect.Struct {
		return nil, ErrStructExpected
	}

	fields := getFieldsByTag(vv.Type(), "query")
	values := make(url.Values)

	for k, fi := range fields {
		fv := vv.FieldByName(fi.Name)
		if fv.IsZero() && fi.Omitempty {
			continue
		}

		if fv.Kind() == reflect.Ptr {
			fv = fv.Elem()
		}

		switch fv.Kind() {
		case reflect.String:
			values.Set(k, fv.String())
		case reflect.Bool:
			values.Set(k, fmt.Sprintf("%v", fv.Bool()))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			values.Set(k, fmt.Sprintf("%v", fv.Int()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			values.Set(k, fmt.Sprintf("%v", fv.Uint()))
		case reflect.Slice:
			if fv.Type().Elem().Kind() == reflect.Uint8 {
				values.Set(k, hex.EncodeToString(fv.Bytes()))
			} else {
				return nil, ErrUnsupportedFieldType
			}
		default:
			return nil, ErrUnsupportedFieldType
		}
	}

	return []byte(values.Encode()), nil
}

func Unmarshal(data []byte, v interface{}) error {
	vv := reflect.ValueOf(v)
	if vv.Kind() != reflect.Ptr || vv.IsNil() {
		return ErrStructPtrExpected
	}
	vv = vv.Elem()

	if vv.Kind() != reflect.Struct {
		return ErrStructPtrExpected
	}

	fields := getFieldsByTag(vv.Type(), "query")

	// ignore parsing errors; ParseQuery returns all valid fields it finds
	values, _ := url.ParseQuery(string(data))

	for k, val := range values {
		if f, ok := fields[k]; ok {
			fv := vv.FieldByName(f.Name)
			if !fv.CanSet() {
				continue
			}

			if fv.Kind() == reflect.Ptr {
				fv.Set(reflect.New(fv.Type().Elem()))
				fv = fv.Elem()
			}

			switch fv.Kind() {
			case reflect.String:
				fv.SetString(val[0])
			case reflect.Bool:
				fv.SetBool(val[0] == "true")
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if i, err := strconv.ParseInt(val[0], 10, 64); err == nil {
					fv.SetInt(i)
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if i, err := strconv.ParseInt(val[0], 10, 63); err == nil {
					fv.SetUint(uint64(i))
				}
			case reflect.Slice:
				if fv.Type().Elem().Kind() == reflect.Uint8 {
					if b, err := hex.DecodeString(val[0]); err == nil {
						fv.SetBytes(b)
					}
				} else {
					return ErrUnsupportedFieldType
				}
			default:
				return ErrUnsupportedFieldType
			}
		}
	}

	return nil
}

func getFieldsByTag(t reflect.Type, tag string) map[string]fieldInfo {
	fields := make(map[string]fieldInfo)

	for i := 0; i < t.NumField(); i++ {
		parts := strings.Split(t.Field(i).Tag.Get(tag), ",")
		name := parts[0]

		if name != "-" && name != "" {
			fields[name] = fieldInfo{
				Name:      t.Field(i).Name,
				Omitempty: slices.Contains(parts[1:], "omitempty"),
			}
		}
	}

	return fields
}

type fieldInfo struct {
	Name      string
	Omitempty bool
}
