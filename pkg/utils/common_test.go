package utils

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestRemoveSliceString(t *testing.T) {
	type input struct {
		slice []string
		s     string
	}

	args := []struct {
		name  string
		input input
		want  []string
	}{
		{
			name: "slice is nil",
			input: input{
				slice: nil,
				s:     "nil",
			},
			want: nil,
		},
		{
			name: "str is not exist",
			input: input{
				slice: []string{"1"},
				s:     "nil",
			},
			want: []string{"1"},
		},
		{
			name: "str is exist",
			input: input{
				slice: []string{"1"},
				s:     "1",
			},
			want: []string{},
		},
	}

	for _, ut := range args {
		t.Run(ut.name, func(t *testing.T) {
			output := RemoveSliceString(ut.input.slice, ut.input.s)
			if !reflect.DeepEqual(output, ut.want) {
				t.Errorf("input (%v, %s), expect %v but got %v", ut.input.slice, ut.input.s, ut.want, output)
			}
		})
	}
}

func TestSliceContainsString(t *testing.T) {
	type input struct {
		slice []string
		s     string
	}

	args := []struct {
		name  string
		input input
		want  bool
	}{
		{
			name: "slice is nil",
			input: input{
				slice: nil,
				s:     "nil",
			},
			want: false,
		},
		{
			name: "str is not exist",
			input: input{
				slice: []string{"1"},
				s:     "nil",
			},
			want: false,
		},
		{
			name: "str is exist",
			input: input{
				slice: []string{"1"},
				s:     "1",
			},
			want: true,
		},
		{
			name: "str is empty",
			input: input{
				slice: []string{"1"},
				s:     "",
			},
			want: false,
		},
	}

	for _, ut := range args {
		t.Run(ut.name, func(t *testing.T) {
			output := SliceContainsString(ut.input.slice, ut.input.s)
			if ut.want != output {
				t.Errorf("input (%v, %s), expect %t but got %t", ut.input.slice, ut.input.s, ut.want, output)
			}
		})
	}
}

func TestGetMapWithDefaultValue(t *testing.T) {
	type input struct {
		m  map[string]string
		s1 string
		s2 string
	}

	args := []struct {
		name  string
		input input
		want  string
	}{
		{
			name: "map is nil",
			input: input{
				m:  nil,
				s1: "nil",
				s2: "def",
			},
			want: "def",
		},
		{
			name: "map key not exist",
			input: input{
				m:  map[string]string{"key": "value"},
				s1: "key1",
				s2: "def",
			},
			want: "def",
		},
		{
			name: "map key exist",
			input: input{
				m:  map[string]string{"key": "value"},
				s1: "key",
				s2: "def",
			},
			want: "value",
		},
	}

	for _, ut := range args {
		t.Run(ut.name, func(t *testing.T) {
			output := GetMapWithDefaultValue(ut.input.m, ut.input.s1, ut.input.s2)
			if output != ut.want {
				t.Errorf("input (%v, %s, %s), expect %s but got %s", ut.input.m, ut.input.s1, ut.input.s2, ut.want, output)
			}
		})
	}
}

func TestMergeMap(t *testing.T) {
	type input struct {
		m1 map[string]string
		m2 map[string]string
	}

	args := []struct {
		name  string
		input input
		want  map[string]string
	}{
		{
			name: "map is nil",
			input: input{
				m1: nil,
				m2: nil,
			},
			want: map[string]string{},
		},
		{
			name: "map is empty",
			input: input{
				m1: map[string]string{},
				m2: map[string]string{},
			},
			want: map[string]string{},
		},
		{
			name: "map1 is nil",
			input: input{
				m1: nil,
				m2: map[string]string{"key2": "value2"},
			},
			want: map[string]string{"key2": "value2"},
		},
		{
			name: "map2 is nil",
			input: input{
				m1: map[string]string{"key1": "value1"},
				m2: nil,
			},
			want: map[string]string{"key1": "value1"},
		},
		{
			name: "map2 merge map1",
			input: input{
				m1: map[string]string{"key1": "value1"},
				m2: map[string]string{"key2": "value2"},
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "map2 merge and override map1",
			input: input{
				m1: map[string]string{
					"key1": "value1",
					"key2": "value1",
				},
				m2: map[string]string{"key2": "value2"},
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	for _, ut := range args {
		t.Run(ut.name, func(t *testing.T) {
			output := MergeMap(ut.input.m1, ut.input.m2)
			if !reflect.DeepEqual(output, ut.want) {
				t.Errorf("input (%+v, %+v), expect %+v but got %+v", ut.input.m1, ut.input.m2, ut.want, output)
			}
		})
	}
}

func TestTransformK8sName(t *testing.T) {
	type input struct {
		s string
	}

	args := []struct {
		name  string
		input input
		want  string
	}{
		{
			name: "empty",
			input: input{
				s: "",
			},
			want: "",
		},
		{
			name: "string upper",
			input: input{
				s: "ABC",
			},
			want: "abc",
		},
		{
			name: "have '-' and '_' str",
			input: input{
				s: "_ABC-aBc_xx-",
			},
			want: "-abc-abc-xx-",
		},
		{
			name: "string upper",
			input: input{
				s: "ABC",
			},
			want: "abc",
		},
		{
			name: "string upper",
			input: input{
				s: "ABC",
			},
			want: "abc",
		},
	}

	for _, ut := range args {
		t.Run(ut.name, func(t *testing.T) {
			output := TransformK8sName(ut.input.s)
			if output != ut.want {
				t.Errorf("input (%s), expect %s but got %s", ut.input.s, ut.want, output)
			}
		})
	}
}

func TestGetPodContainerImageVersion(t *testing.T) {
	type input struct {
		containerName string
		podSpec       *corev1.PodSpec
	}
	args := []struct {
		name  string
		input input
		want  string
	}{
		{
			name: "podSpec is nil",
			input: input{
				containerName: "containerName",
				podSpec:       nil,
			},
			want: "",
		},
		{
			name: "multi containers but not exist specify container",
			input: input{
				containerName: "containerName",
				podSpec: &corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "containerName1",
							Image: "image:v1",
						},
						{
							Name:  "containerName2",
							Image: "image:v2",
						},
					},
				},
			},
			want: "",
		},
		{
			name: "multi containers exist specify container",
			input: input{
				containerName: "containerName1",
				podSpec: &corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "containerName1",
							Image: "image:v1",
						},
						{
							Name:  "containerName2",
							Image: "image:v2",
						},
					},
				},
			},
			want: "v1",
		},
		{
			name: "multi containers have error image content",
			input: input{
				containerName: "containerName1",
				podSpec: &corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "containerName1",
							Image: "imageerror",
						},
						{
							Name:  "containerName2",
							Image: "image:v2",
						},
					},
				},
			},
			want: "",
		},
	}

	for _, ut := range args {
		t.Run(ut.name, func(t *testing.T) {
			output := GetPodContainerImageVersion(ut.input.containerName, ut.input.podSpec)
			if output != ut.want {
				t.Errorf("input (%s, %v), expect %s but got %s", ut.input.containerName, ut.input.podSpec, ut.want, output)
			}
		})
	}
}

func TestTransInt32PtrInt32(t *testing.T) {
	type input struct {
		i   *int32
		def int32
	}

	var normal int32 = 10

	args := []struct {
		name  string
		input input
		want  int32
	}{
		{
			name: "i is nil",
			input: input{
				i:   nil,
				def: 1,
			},
			want: 1,
		},
		{
			name: "normal",
			input: input{
				i:   &normal,
				def: 1,
			},
			want: normal,
		},
	}

	for _, ut := range args {
		t.Run(ut.name, func(t *testing.T) {
			output := TransInt32Ptr2Int32(ut.input.i, ut.input.def)
			if output != ut.want {
				t.Errorf("input (%v, %d), expect %d but got %d", ut.input.i, ut.input.def, ut.want, output)
			}
		})
	}
}
