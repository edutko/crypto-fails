package auth

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

type Session struct {
	IsAdmin  bool   `query:"adm"`
	Expires  int    `query:"exp"`
	RealName string `query:"name,omitempty"`
	Username string `query:"uid"`
}

func GetCurrentSession(ctx context.Context) *Session {
	if s := ctx.Value(sessionCtxKey); s != nil {
		return s.(*Session)
	}
	return nil
}

func ContextWithSession(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, sessionCtxKey, s)
}

func (s *Session) QueryString() string {
	type fieldInfo struct {
		name      string
		omitempty bool
	}
	t := reflect.TypeOf(*s)
	fieldMap := make(map[string]fieldInfo)
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("query")
		tag = strings.Split(tag, ",")[0]
		if tag != "-" && tag != "" {
			fieldMap[tag] = fieldInfo{
				name:      t.Field(i).Name,
				omitempty: strings.Contains(t.Field(i).Tag.Get("query"), ",omitempty"),
			}
		}
	}

	values := make(url.Values)
	for k, fi := range fieldMap {
		fv := reflect.ValueOf(s).Elem().FieldByName(fi.name)
		if fv.IsZero() && fi.omitempty {
			continue
		}
		switch fv.Kind() {
		case reflect.String:
			values.Set(k, fv.String())
		case reflect.Bool:
			values.Set(k, fmt.Sprintf("%v", fv.Bool()))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			values.Set(k, fmt.Sprintf("%v", fv.Int()))
		default:
			panic("unexpected field type")
		}
	}

	return values.Encode()
}

func ParseSession(qs string) *Session {
	var s Session
	t := reflect.TypeOf(s)
	fieldMap := make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("query")
		tag = strings.Split(tag, ",")[0]
		if tag != "-" && tag != "" {
			fieldMap[tag] = t.Field(i).Name
		}
	}

	// ignore parsing errors
	values, _ := url.ParseQuery(qs)

	for k, v := range values {
		if name, ok := fieldMap[k]; ok {
			fv := reflect.ValueOf(&s).Elem().FieldByName(name)
			if !fv.CanSet() {
				continue
			}
			switch fv.Kind() {
			case reflect.String:
				fv.SetString(v[0])
			case reflect.Bool:
				fv.SetBool(v[0] == "true")
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if i, err := strconv.ParseInt(v[0], 10, 64); err == nil {
					fv.SetInt(i)
				}
			default:
				panic("unexpected field type")
			}
		}
	}

	return &s
}

const sessionCtxKey = "session"
