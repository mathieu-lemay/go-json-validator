package validation

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/mathieu-lemay/go-json-validator/internal/logging"

	"github.com/go-playground/validator/v10"
	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
)

var (
	_    = logging.ConfigureLogger(logging.WithLevel(zerolog.DebugLevel))
	fake = faker.New()
)

type JsonMap map[string]interface{}

type T struct {
	I int    `json:"i" validate:"required"`
	S string `json:"s" validate:"required"`
	B bool   `json:"b" validate:"required"`
}

type PT struct {
	I *int
	S *string
	B *bool
}

func TestParse(t *testing.T) {
	testCases := []struct {
		msg      string
		data     string
		isValid  bool
		expected JsonMap
	}{
		{
			"missing values",
			`{}`,
			false,
			JsonMap{
				"i": nil,
				"s": nil,
				"b": nil,
			},
		},
		{
			"null values",
			`{
				"i": null,
				"s": null,
				"b": null
			}`,
			false,
			JsonMap{
				"i": nil,
				"s": nil,
				"b": nil,
			},
		},
		{
			"zero values",
			`{
				"i": 0,
				"s": "",
				"b": false
			}`,
			true,
			JsonMap{
				"i": float64(0),
				"s": "",
				"b": false,
			},
		},
		{
			"non-zero values",
			`{
				"i": 5,
				"s": "string",
				"b": true
			}`,
			true,
			JsonMap{
				"i": float64(5),
				"s": "string",
				"b": true,
			},
		},
	}

	for _, tc := range testCases {
		var st T
		res, err := parse([]byte(tc.data), &st)
		assert.Equal(t, tc.expected, must(toMap(res)), tc.msg)
		assert.NoError(t, err, tc.msg)
	}
}

func TestCopyStruct(t *testing.T) {
	src := PT{
		I: ptr(42),
		S: ptr("str"),
		B: ptr(true),
	}

	var dst T

	err := copyStruct(&src, &dst)

	expected := T{
		I: 42,
		S: "str",
		B: true,
	}

	assert.Equal(t, expected, dst)
	assert.NoError(t, err)
}

func TestCopyStructPointers(t *testing.T) {
	i := fake.Int()
	s := fake.RandomStringWithLength(8)
	b := fake.Bool()

	src := PT{
		I: &i,
		S: &s,
		B: &b,
	}

	var dst PT

	err := copyStruct(&src, &dst)
	require.NoError(t, err)

	assert.Equal(t, src, dst)
	assert.Equal(t, i, *dst.I)
	assert.Equal(t, s, *dst.S)
	assert.Equal(t, b, *dst.B)
}

func TestCopyStruct_Errors(t *testing.T) {
	src := PT{
		I: nil,
		S: ptr("str"),
		B: ptr(true),
	}

	var dst T

	err := copyStruct(&src, &dst)

	assert.ErrorContains(t, err, "can't set nil value to non ptr field: I")
}

func TestParseAndValidateOptionalValues(t *testing.T) {
	testCases := []struct {
		msg      string
		data     string
		isValid  bool
		expected *T
	}{
		{
			"missing values",
			`{}`,
			false,
			nil,
		},
		{
			"null values",
			`{
				"i": null,
				"s": null,
				"b": null
			}`,
			false,
			nil,
		},
		{
			"zero values",
			`{
				"i": 0,
				"s": "",
				"b": false
			}`,
			true,
			&T{I: 0, S: "", B: false},
		},
		{
			"non-zero values",
			`{
				"i": 5,
				"s": "string",
				"b": true
			}`,
			true,
			&T{I: 5, S: "string", B: true},
		},
	}

	for _, tc := range testCases {
		var st T
		err := parseAndValidate([]byte(tc.data), &st)
		if tc.isValid {
			assert.NoError(t, err, tc.msg)
			assert.Equal(t, *tc.expected, st, tc.msg)
		} else {
			assert.Error(t, err, tc.msg)
			assert.ErrorAs(t, err, &validator.ValidationErrors{}, tc.msg)
		}
	}
}

func BenchmarkReference(b *testing.B) {
	_ = logging.ConfigureLogger(logging.WithLevel(zerolog.ErrorLevel))
	data := []byte(`{"i": 45, "s": "some-str", "b": true}`)

	for i := 0; i < b.N; i++ {
		var st T
		err := json.Unmarshal(data, &st)
		assert.NoError(b, err)

		err = validate.Struct(&st)
		assert.NoError(b, err)
	}
}

func BenchmarkParseAndValidate(b *testing.B) {
	_ = logging.ConfigureLogger(logging.WithLevel(zerolog.ErrorLevel))
	data := []byte(`{"i": 45, "s": "some-str", "b": true}`)

	for i := 0; i < b.N; i++ {
		var st T
		err := parseAndValidate(data, &st)
		assert.NoError(b, err)
	}
}

func BenchmarkParseAndValidateNoCache(b *testing.B) {
	_ = logging.ConfigureLogger(logging.WithLevel(zerolog.ErrorLevel))
	data := []byte(`{"i": 45, "s": "some-str", "b": true}`)

	for i := 0; i < b.N; i++ {
		structsCache = make(map[reflect.Type]reflect.Type)

		var st T
		err := parseAndValidate(data, &st)
		assert.NoError(b, err)
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}

func ptr[T any](v T) *T {
	return &v
}

func toMap(t interface{}) (JsonMap, error) {
	a, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	var res JsonMap
	json.Unmarshal(a, &res)

	return res, nil
}
