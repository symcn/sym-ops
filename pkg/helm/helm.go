package helm

import (
	"bytes"
	"fmt"
	"strings"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"k8s.io/klog/v2"
)

// RenderTemplate render chart template to k8s object slice
func RenderTemplate(chartPkg []byte, rlsName, ns string, overrideValue string) ([]K8sObject, error) {
	renderedTpls, err := renderTpls(chartPkg, rlsName, ns, overrideValue)
	if err != nil {
		return nil, err
	}
	return buildK8sObjectWithRenderedTpls(renderedTpls)
}

func renderTpls(chartPkg []byte, rlsName, ns string, overrideValue string) (renderedTpls map[string]string, err error) {
	chrt, err := loader.LoadArchive(bytes.NewBuffer(chartPkg))
	if err != nil {
		return nil, fmt.Errorf("loading chart has an error: %v", err)
	}
	chrtVals, err := chartutil.ReadValues([]byte(overrideValue))
	if err != nil {
		return nil, fmt.Errorf("read overridevalue has an error: %v", err)
	}
	opts := chartutil.ReleaseOptions{
		Name:      rlsName,
		Namespace: ns,
		Revision:  1,
		IsInstall: true,
		IsUpgrade: false,
	}
	chrtValues, err := chartutil.ToRenderValues(chrt, chrtVals, opts, nil)
	if err != nil {
		return nil, fmt.Errorf("render chart values has an error: %v", err)
	}
	renderedTpls, err = engine.Render(chrt, chrtValues)
	if err != nil {
		return nil, fmt.Errorf("render error: %v", err)
	}
	return renderedTpls, nil
}

func buildK8sObjectWithRenderedTpls(renderedTpls map[string]string) ([]K8sObject, error) {
	var objects []K8sObject
	for _, tpl := range renderedTpls {
		yaml := removeNonYAMLLines(tpl)
		if yaml == "" {
			continue
		}
		o, err := ParseYAML2K8sObject([]byte(yaml))
		if err != nil {
			klog.Errorf("Failed to parse yaml %s to k8s object: %v", yaml, err)
			continue
		}
		objects = append(objects, o)
		klog.V(5).Infof("Render k8s object %s %s/%s success", o.GroupKind().Kind, o.GetNamespace(), o.GetName())
	}
	return objects, nil
}

func removeNonYAMLLines(yamlStr string) string {
	out := ""
	for _, s := range strings.Split(yamlStr, "\n") {
		if strings.HasPrefix(s, "#") {
			continue
		}
		out += s + "\n"
	}

	return strings.TrimSpace(out)
}
