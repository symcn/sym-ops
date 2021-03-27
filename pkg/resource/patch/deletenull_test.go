package patch

import (
	"reflect"
	"testing"
)

var (
	fc         func()
	emptymap   map[string]string
	emptyslice []string
)

func fc2() {}

func TestIsZero(t *testing.T) {
	args := []struct {
		name  string
		input interface{}
		want  bool
	}{
		{
			name:  "nil",
			input: nil,
			want:  true,
		},
		{
			name:  "empty string",
			input: "",
			want:  true,
		},
		{
			name:  "zero int64",
			input: int64(0),
			want:  false,
		},
		{
			name:  "zero float64",
			input: float64(0.0),
			want:  false,
		},
		{
			name:  "zero int32",
			input: int32(0),
			want:  true,
		},
		{
			name:  "func",
			input: fc,
			want:  true,
		},
		{
			name:  "func not error",
			input: fc2,
			want:  false,
		},
		{
			name:  "struct",
			input: struct{}{},
			want:  true,
		},
		{
			name:  "empty map",
			input: emptymap,
			want:  true,
		},
		{
			name:  "map",
			input: map[string]interface{}{},
			want:  false,
		},
		{
			name:  "empty slice",
			input: emptyslice,
			want:  true,
		},
		{
			name:  "slice",
			input: []string{},
			want:  false,
		},
		{
			name:  "array",
			input: [3]int{},
			want:  true,
		},
	}

	for _, ut := range args {
		t.Run(ut.name, func(t *testing.T) {
			out := isZero(reflect.ValueOf(ut.input))
			if out != ut.want {
				t.Errorf("input %+v want %t but got %t", ut.input, ut.want, out)
			}
		})
	}
}

func TestDeleteNullInObj(t *testing.T) {
	args := []struct {
		name  string
		input map[string]interface{}
		want  map[string]interface{}
	}{
		{
			name:  "nil",
			input: nil,
			want:  map[string]interface{}{},
		},
		{
			name: "unsupport type",
			input: map[string]interface{}{
				"int": 1,
			},
			want: map[string]interface{}{
				"int": 1,
			},
		},
		{
			name: "complex obj",
			input: map[string]interface{}{
				"slice":     []interface{}{},
				"string":    "",
				"float64":   float64(0.0),
				"bool":      false,
				"int64":     int64(1),
				"nil":       nil,
				"empty map": map[string]interface{}{},
				"empty submap": map[string]interface{}{
					"empty": nil,
				},
				"map": map[string]interface{}{
					"a": "b",
				},
				"empty array":      [0]int{},
				"value both array": [1]int{},
				"array":            [1]int{1},
				"empty struct":     struct{}{},
				"struct":           struct{ name string }{name: "name"},
			},
			want: map[string]interface{}{
				"float64":   float64(0.0),
				"bool":      false,
				"int64":     int64(1),
				"empty map": map[string]interface{}{},
				"map": map[string]interface{}{
					"a": "b",
				},
				"array":  [1]int{1},
				"struct": struct{ name string }{name: "name"},
			},
		},
	}

	for _, ut := range args {
		t.Run(ut.name, func(t *testing.T) {
			out := deleteNullInObj(ut.input)
			if !reflect.DeepEqual(out, ut.want) {
				t.Errorf("input(%+v) want got\n%+v, but got\n%+v", ut.input, ut.want, out)
			}
		})
	}
}

func TestDeleteNullInSlice(t *testing.T) {
	args := []struct {
		name  string
		input []interface{}
		want  []interface{}
	}{
		{
			name:  "nil",
			input: nil,
			want:  []interface{}{},
		},
		{
			name: "complex",
			input: []interface{}{
				"",
				int64(0),
				float64(0.0),
				map[string]interface{}{
					"a": nil,
				},
			},
			want: []interface{}{
				int64(0),
				float64(0.0),
				map[string]interface{}{},
			},
		},
	}

	for _, ut := range args {
		t.Run(ut.name, func(t *testing.T) {
			out := deleteNullInSlice(ut.input)
			if !reflect.DeepEqual(out, ut.want) {
				t.Errorf("input(%+v) want got\n%+v, but got\n%+v", ut.input, ut.want, out)
			}
		})
	}
}

func TestDeleteNullInJSON(t *testing.T) {
	args := []struct {
		name  string
		input []byte
		want1 []byte
		want2 map[string]interface{}
		isErr bool
	}{
		{
			name:  "nil json",
			input: nil,
			want1: nil,
			want2: nil,
			isErr: true,
		},
		{
			name:  "empty string",
			input: []byte(""),
			want1: nil,
			want2: nil,
			isErr: true,
		},
		{
			name:  "error json str",
			input: []byte("error_json"),
			want1: nil,
			want2: nil,
			isErr: true,
		},
		{
			name:  "empty json",
			input: []byte("{}"),
			want1: []byte("{}"),
			want2: map[string]interface{}{},
			isErr: false,
		},
		{
			name:  "json have null string struct slice",
			input: []byte(`{"a":"","b":[],"c":{}}`),
			want1: []byte(`{"c":{}}`),
			want2: map[string]interface{}{
				"c": map[string]interface{}{},
			},
			isErr: false,
		},
	}

	for _, ut := range args {
		t.Run(ut.name, func(t *testing.T) {
			out1, out2, err := DeleteNullInJson(ut.input)
			if err != nil {
				if !ut.isErr {
					t.Errorf("want no error, but got: %v", err)
				}
				return
			}
			if ut.isErr {
				t.Error("want get error, but is nil")
				return
			}

			if !reflect.DeepEqual(out1, ut.want1) || !reflect.DeepEqual(out2, ut.want2) {
				t.Errorf("input(%s) \nwant got%s, %v\nbut got%s, %v", ut.input, ut.want1, ut.want2, out1, out2)
			}
		})
	}
}
