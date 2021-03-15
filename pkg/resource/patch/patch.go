package patch

import (
	"fmt"

	json "github.com/json-iterator/go"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var DefaultPatchMaker = NewPatchMaker(DefaultAnnotator)

type PatchMaker struct {
	annotator *Annotator
}

func NewPatchMaker(annotator *Annotator) *PatchMaker {
	return &PatchMaker{annotator: annotator}
}

func (p *PatchMaker) Calculate(currentObj, modifiedObj rtclient.Object, opts ...CalculateOption) (*PatchResult, error) {
	current, err := json.ConfigCompatibleWithStandardLibrary.Marshal(currentObj)
	if err != nil {
		return nil, fmt.Errorf("convert current object to byte failed: %v", err)
	}
	modified, err := json.ConfigCompatibleWithStandardLibrary.Marshal(modifiedObj)
	if err != nil {
		return nil, fmt.Errorf("convert modified object to byte failed: %v", err)
	}

	for _, opt := range opts {
		current, modified, err = opt(current, modified)
		if err != nil {
			return nil, fmt.Errorf("apply calculate option failed: %v", err)
		}
	}

	current, _, err = DeleteNullInJson(current)
	if err != nil {
		return nil, fmt.Errorf("delete null from current byte failed: %v", err)
	}
	modified, _, err = DeleteNullInJson(modified)
	if err != nil {
		return nil, fmt.Errorf("delete null from modified byte failed: %v", err)
	}

	original, err := p.annotator.GetOriginalConfiguration(currentObj)
	if err != nil {
		return nil, fmt.Errorf("get original configuration failed: %v", err)
	}

	var patch []byte

	switch currentObj.(type) {
	case *unstructured.Unstructured:
		patch, err = jsonmergepatch.CreateThreeWayJSONMergePatch(original, modified, current)
		if err != nil {
			return nil, fmt.Errorf("generate merge patch failed: %v", err)
		}
	default:
		lookupPatchMeta, err := strategicpatch.NewPatchMetaFromStruct(currentObj)
		if err != nil {
			return nil, fmt.Errorf("lookup obj %v patch meta failed: %v", currentObj, err)
		}
		patch, err = strategicpatch.CreateThreeWayMergePatch(original, modified, current, lookupPatchMeta, true)
		if err != nil {
			return nil, fmt.Errorf("generate strategic merge patch failed: %v", err)
		}
		// $setElementOrder can make it hard to decide whether there is an actual diff or not.
		// In cases like that trying to apply the patch locally on current will make it clear.
		if string(patch) != "{}" {
			patchCurrent, err := strategicpatch.StrategicMergePatch(current, patch, currentObj)
			if err != nil {
				return nil, fmt.Errorf("apply patch again to check for an actual diff failed: %v", err)
			}
			patch, err = strategicpatch.CreateTwoWayMergePatch(current, patchCurrent, currentObj)
			if err != nil {
				return nil, fmt.Errorf("create patch again to check for an actual diff failed: %v", err)
			}
		}
	}

	return &PatchResult{
		Patch:    patch,
		Current:  current,
		Modified: modified,
		Original: original,
	}, nil
}

type CalculateOption func([]byte, []byte) ([]byte, []byte, error)

type PatchResult struct {
	Patch    []byte
	Current  []byte
	Modified []byte
	Original []byte
}

func (p *PatchResult) IsEmpty() bool {
	return string(p.Patch) == "{}"
}
