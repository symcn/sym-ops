package utils

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsIgnoreAnnotationKey(t *testing.T) {
	type input struct {
		k string
	}

	args := []struct {
		name  string
		input input
		want  bool
	}{
		{
			name: "k is empty",
			input: input{
				k: "",
			},
			want: false,
		},
		{
			name: "ignore",
			input: input{
				k: ClusterAnnotationMonitor,
			},
			want: true,
		},
		{
			name: "not ignore",
			input: input{
				k: "notignore",
			},
			want: false,
		},
	}

	for _, ut := range args {
		t.Run(ut.name, func(t *testing.T) {
			output := isIgnoreAnnotationKey(ut.input.k)
			if output != ut.want {
				t.Errorf("input (%s), expect %t but got %t", ut.input.k, ut.want, output)
			}
		})
	}
}

func TestObjectMetaEqual(t *testing.T) {
	type input struct {
		obj1 metav1.Object
		obj2 metav1.Object
	}

	args := []struct {
		name  string
		input input
		want  bool
	}{
		{
			name: "obj1 an obj2 both nil",
			input: input{
				obj1: nil,
				obj2: nil,
			},
			want: true,
		},
		{
			name: "obj1 is nil",
			input: input{
				obj1: nil,
				obj2: &appsv1.Deployment{},
			},
			want: false,
		},
		{
			name: "obj2 is nil",
			input: input{
				obj1: &appsv1.Deployment{},
				obj2: nil,
			},
			want: false,
		},
		{
			name: "different finalizer",
			input: input{
				obj1: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Finalizers: []string{"finalizers1"},
					},
				},
				obj2: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Finalizers: []string{"finalizers2"},
					},
				},
			},
			want: false,
		},
		{
			name: "different labels",
			input: input{
				obj1: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
						Finalizers: []string{"finalizers"},
					},
				},
				obj2: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
							"key3": "value3",
						},
						Finalizers: []string{"finalizers"},
					},
				},
			},
			want: false,
		},
		{
			name: "different annotation without ignore",
			input: input{
				obj1: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
						Annotations: map[string]string{
							"a1": "v1",
							"a2": "v2",
						},
						Finalizers: []string{"finalizers"},
					},
				},
				obj2: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
						Annotations: map[string]string{
							"a1": "v1",
							"a2": "v1",
						},
						Finalizers: []string{"finalizers"},
					},
				},
			},
			want: false,
		},
		{
			name: "different annotation with ignore",
			input: input{
				obj1: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
						Annotations: map[string]string{
							"a1":                     "v1",
							"a2":                     "v2",
							ClusterAnnotationMonitor: "111",
						},
						Finalizers: []string{"finalizers"},
					},
				},
				obj2: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
						Annotations: map[string]string{
							"a1":                     "v1",
							"a2":                     "v2",
							ClusterAnnotationMonitor: "123",
						},
						Finalizers: []string{"finalizers"},
					},
				},
			},
			want: true,
		},
	}

	for _, ut := range args {
		t.Run(ut.name, func(t *testing.T) {
			output := ObjecteMetaEqual(ut.input.obj1, ut.input.obj2)
			if output != ut.want {
				t.Errorf("input (%v, %v), expect %t but got %t", ut.input.obj1, ut.input.obj2, ut.want, output)
			}
		})
	}
}
