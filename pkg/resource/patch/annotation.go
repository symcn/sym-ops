package patch

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime/debug"

	json "github.com/json-iterator/go"
	"github.com/symcn/sym-ops/pkg/types"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var DefaultAnnotator = NewAnnotator(types.LastAppliedConfig)

type Annotator struct {
	metadataAccessor meta.MetadataAccessor
	key              string
}

func NewAnnotator(key string) *Annotator {
	return &Annotator{
		metadataAccessor: meta.NewAccessor(),
		key:              key,
	}
}

func (a *Annotator) SetLastAppliedAnnotation(obj rtclient.Object) error {
	modified, err := a.GetModifiedConfiguration(obj, false)
	if err != nil {
		return err
	}
	// Remove nulls from json
	modifiedWithoutNulls, _, err := DeleteNullInJson(modified)
	if err != nil {
		return err
	}
	return a.SetOriginalConfiguration(obj, modifiedWithoutNulls)
}

// GetModifiedConfiguration retrieves the modified configuration of the object.
// If annotate if true, it embeds the result as an annotation in the modified
// configuration. If an object was read from the command input, it will use that
// version of the object. Otherwise, it will use the version from the server.
func (a *Annotator) GetModifiedConfiguration(obj rtclient.Object, annotate bool) ([]byte, error) {
	// First serialize the object without the annotation to prevent recursion,
	// then add that serialization to it as the annotation and serialize it again.
	var modified []byte

	// Otherwise, use the server side version of the object.
	// Get the current annotations from the object.
	annots, err := a.metadataAccessor.Annotations(obj)
	if err != nil {
		return nil, err
	}

	if annots == nil {
		annots = map[string]string{}
	}

	original := annots[a.key]
	delete(annots, a.key)
	if err = a.metadataAccessor.SetAnnotations(obj, annots); err != nil {
		return nil, err
	}

	// Do not include an empty annotation map
	if len(annots) == 0 {
		a.metadataAccessor.SetAnnotations(obj, nil)
	}
	modified, err = json.ConfigCompatibleWithStandardLibrary.Marshal(obj)
	if err != nil {
		return nil, err
	}

	if annotate {
		annots[a.key], err = zipAndBase64EncodeAnnotation(modified)
		if err != nil {
			return nil, err
		}
		if err = a.metadataAccessor.SetAnnotations(obj, annots); err != nil {
			return nil, err
		}

		modified, err = json.ConfigCompatibleWithStandardLibrary.Marshal(obj)
		if err != nil {
			return nil, err
		}
	}

	// Restore the object to its original condition.
	annots[a.key] = original
	if err = a.metadataAccessor.SetAnnotations(obj, annots); err != nil {
		return nil, err
	}
	return modified, nil
}

func zipAndBase64EncodeAnnotation(original []byte) (string, error) {
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new zip archive.
	w := zip.NewWriter(buf)

	f, err := w.Create("original")
	if err != nil {
		return "", err
	}
	_, err = f.Write(original)
	if err != nil {
		return "", err
	}

	// Make sure to check the error on Close.
	err = w.Close()
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func (a *Annotator) SetOriginalConfiguration(obj runtime.Object, original []byte) error {
	if len(original) < 1 {
		return nil
	}

	annots, err := a.metadataAccessor.Annotations(obj)
	if err != nil {
		return err
	}

	if annots == nil {
		annots = map[string]string{}
	}

	annots[a.key], err = zipAndBase64EncodeAnnotation(original)
	if err != nil {
		return err
	}
	return a.metadataAccessor.SetAnnotations(obj, annots)
}

func (a *Annotator) GetOriginalConfiguration(obj rtclient.Object) ([]byte, error) {
	annots, err := a.metadataAccessor.Annotations(obj)
	if err != nil {
		return nil, err
	}
	if annots == nil {
		return nil, nil
	}

	original, ok := annots[a.key]
	if !ok {
		return nil, nil
	}

	if decoded, err := base64.StdEncoding.DecodeString(original); err == nil {
		if http.DetectContentType(decoded) == "application/zip" {
			return unZipAnnotation(decoded)
		}
	}
	return []byte(original), nil
}

func unZipAnnotation(original []byte) ([]byte, error) {
	annotation, err := ioutil.ReadAll(bytes.NewReader(original))
	if err != nil {
		return nil, err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(annotation), int64(len(annotation)))
	if err != nil {
		return nil, err
	}

	zipFile := zipReader.File[0]
	unzippedFileBytes, err := readZipFile(zipFile)
	if err != nil {
		return nil, err
	}
	return unzippedFileBytes, nil
}

func readZipFile(zf *zip.File) (data []byte, err error) {
	if zf == nil {
		return nil, errors.New("zip file is nil")
	}
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("readZipFile error:\n%v\n%s", e, debug.Stack())
		}
	}()

	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ioutil.ReadAll(f)
}
