package urlquery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	testCases := []struct {
		name        string
		val         interface{}
		expected    []byte
		expectedErr error
	}{
		{"basic struct", testStruct1{"foo"}, []byte("str=foo"), nil},
		{"pointer to struct", &testStruct1{"bar"}, []byte("str=bar"), nil},
		{"complex struct", testStruct2{
			Str:       "foo",
			Bool:      false,
			Int:       1,
			SecondInt: 0,
			Uint:      2,
			Float:     3.14,
			Bytes:     []byte("bar"),
		}, []byte("bool=false&bytes=626172&int=1&str=foo&uint=2"), nil},
		{"pointer field", testStruct3{"foo", ptr("bar")}, []byte("pstr=bar&str=foo"), nil},

		{"string", "hi", nil, ErrStructExpected},
		{"int", 123, nil, ErrStructExpected},
		{"bool", true, nil, ErrStructExpected},
		{"float", 1.23, nil, ErrStructExpected},
		{"slice", []byte{0, 1, 2}, nil, ErrStructExpected},
		{"map", map[string]int{"0": 0, "1": 1}, nil, ErrStructExpected},

		{"unsupported field: slice", testStruct4{"foo", []int{1, 2, 3}}, nil, ErrUnsupportedFieldType},
		{"unsupported field: struct", testStruct5{"foo", testStruct1{"bar"}}, nil, ErrUnsupportedFieldType},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := Marshal(tc.val)

			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expected, b)
		})
	}
}

func TestUnmarshal(t *testing.T) {
	testCases := []struct {
		name        string
		qs          []byte
		val         interface{}
		expectedVal interface{}
		expectedErr error
	}{
		{"basic struct", []byte("str=foo"), &testStruct1{}, &testStruct1{"foo"}, nil},
		{"complex struct", []byte("bool=false&bytes=626172&int=1&str=foo&uint=2"), &testStruct2{}, &testStruct2{
			Str:       "foo",
			Bool:      false,
			Int:       1,
			SecondInt: 0,
			Uint:      2,
			Bytes:     []byte("bar"),
		}, nil},
		{"pointer field", []byte("pstr=bar&str=foo"), &testStruct3{}, &testStruct3{"foo", ptr("bar")}, nil},

		{"nil", []byte("str=foo"), nil, nil, ErrStructPtrExpected},
		{"string", []byte("str=foo"), "hi", "hi", ErrStructPtrExpected},
		{"string pointer", []byte("str=foo"), ptr("hi"), ptr("hi"), ErrStructPtrExpected},

		{"unsupported field: slice", []byte("ints=foo"), &testStruct4{}, &testStruct4{}, ErrUnsupportedFieldType},
		{"unsupported field: struct", []byte("struct=foo"), &testStruct5{}, &testStruct5{}, ErrUnsupportedFieldType},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Unmarshal(tc.qs, tc.val)

			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedVal, tc.val)
		})
	}
}

type testStruct1 struct {
	Str string `query:"str"`
}

type testStruct2 struct {
	Str       string  `query:"str"`
	Bool      bool    `query:"bool"`
	Int       int     `query:"int"`
	SecondInt int     `query:"second_int,omitempty"`
	Uint      uint    `query:"uint"`
	Float     float64 `query:"-"`
	Bytes     []byte  `query:"bytes"`
}

type testStruct3 struct {
	Str string  `query:"str"`
	Ptr *string `query:"pstr"`
}

type testStruct4 struct {
	Str  string `query:"str"`
	Ints []int  `query:"ints"`
}

type testStruct5 struct {
	Str    string      `query:"str"`
	Struct testStruct1 `query:"struct"`
}

func ptr[T any](v T) *T {
	return &v
}
