package patch

import (
	"archive/zip"
	"bytes"
	"reflect"
	"testing"
)

func TestReadZipFile(t *testing.T) {
	emptyReader, err := zip.NewReader(bytes.NewReader([]byte("")), 0)
	if err != nil {
		t.Error(err)
		return
	}
	args := []struct {
		name    string
		input   *zip.File
		want    []byte
		isError bool
	}{
		{
			name:    "input is nil",
			input:   nil,
			want:    nil,
			isError: true,
		},
		{
			name:    "input is empty",
			input:   &zip.File{},
			want:    nil,
			isError: true,
		},
		{
			name:    "input is empty",
			input:   emptyReader.File[0],
			want:    nil,
			isError: true,
		},
	}

	for _, ut := range args {
		t.Run(ut.name, func(t *testing.T) {
			out, err := readZipFile(ut.input)
			if err != nil {
				if !ut.isError {
					t.Errorf("want not have error, but got: %v", err)
				}
				return
			}
			if ut.isError {
				t.Error("want have error, but got nil")
				return
			}
			if !reflect.DeepEqual(ut.want, out) {
				t.Errorf("input (%v) want %v, but got:%v", ut.input, ut.want, out)
			}
		})
	}
}
