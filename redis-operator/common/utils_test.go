package common

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestMergeEnvs(t *testing.T) {
	type args struct {
		from []corev1.EnvVar
		to   []corev1.EnvVar
	}
	tests := []struct {
		name string
		args args
		want []corev1.EnvVar
	}{
		{
			name: "Merge with intersections, len(from) < len(to) ",
			args: args{
				from: []corev1.EnvVar{{Name: "foo1", Value: "bar1"}},
				to:   []corev1.EnvVar{{Name: "foo1", Value: "bar1"}, {Name: "foo2", Value: "bar2"}},
			},
			want: []corev1.EnvVar{{Name: "foo1", Value: "bar1"}, {Name: "foo2", Value: "bar2"}},
		},
		{
			name: "Merge with intersections, len(from) > len(to) ",
			args: args{
				from: []corev1.EnvVar{{Name: "foo1", Value: "bar1"}, {Name: "foo2", Value: "bar2"}},
				to:   []corev1.EnvVar{{Name: "foo1", Value: "bar1"}},
			},
			want: []corev1.EnvVar{{Name: "foo1", Value: "bar1"}, {Name: "foo2", Value: "bar2"}},
		},
		{
			name: "Merge with no intersections",
			args: args{
				from: []corev1.EnvVar{{Name: "foo1", Value: "bar1"}, {Name: "foo2", Value: "bar2"}},
				to:   []corev1.EnvVar{{Name: "foo3", Value: "bar3"}},
			},
			want: []corev1.EnvVar{{Name: "foo3", Value: "bar3"}, {Name: "foo1", Value: "bar1"}, {Name: "foo2", Value: "bar2"}},
		},
		{
			name: "Merge empty first",
			args: args{
				from: []corev1.EnvVar{},
				to:   []corev1.EnvVar{{Name: "foo", Value: "bar"}},
			},
			want: []corev1.EnvVar{{Name: "foo", Value: "bar"}},
		},
		{
			name: "Merge empty second",
			args: args{
				from: []corev1.EnvVar{{Name: "foo", Value: "bar"}},
				to:   []corev1.EnvVar{},
			},
			want: []corev1.EnvVar{{Name: "foo", Value: "bar"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeEnvs(tt.args.from, tt.args.to); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeEnvs() = %v, want %v", got, tt.want)
			}
		})
	}
}
