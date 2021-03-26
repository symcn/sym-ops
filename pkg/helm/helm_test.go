package helm

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

var (
	repoName      = "stable"
	repoURL       = "http://mirror.azure.cn/kubernetes/charts/"
	chartName     = "nginx-ingress"
	chartVersion  = "1.41.3"
	downloadDir   = "./Charts"
	overrideValue = `
controller:
  name: %s
`
	charts       = []byte{}
	renderedTpls = map[string]string{}
)

func downloadCharts() error {
	cfg := &repo.Entry{
		Name: repoName,
		URL:  repoURL,
	}
	env := cli.New()

	// add repo
	rp, err := repo.NewChartRepository(cfg, getter.All(env))
	if err != nil {
		return err
	}
	_, err = rp.DownloadIndexFile()
	if err != nil {
		return err
	}

	// update repo to RepositoryConfig
	var f repo.File
	repoFile := env.RepositoryConfig
	b, err := ioutil.ReadFile(repoFile)
	if err != nil && os.IsNotExist(err) {
		return err
	}
	err = yaml.Unmarshal(b, &f)
	if err != nil {
		return err
	}
	f.Update(cfg)
	err = f.WriteFile(repoFile, 0644)
	if err != nil {
		return err
	}

	// download charts
	dl := downloader.ChartDownloader{
		Out:              os.Stdout,
		RepositoryConfig: repoFile,
		RepositoryCache:  env.RepositoryCache,
		Getters:          getter.All(env),
	}
	err = os.MkdirAll(downloadDir, 0755)
	if err != nil {
		return err
	}

	filename, _, err := dl.DownloadTo(repoName+"/"+chartName, chartVersion, downloadDir)
	if err != nil {
		return err
	}
	fmt.Println("download file:", filename)

	charts, err = os.ReadFile(filename)
	if err != nil {
		return err
	}
	return nil
}

func TestRenderTpls(t *testing.T) {
	err := downloadCharts()
	if err != nil {
		t.Error(err)
		return
	}

	// error chartPkg
	_, err = renderTpls([]byte("error charts"), chartName, "", "")
	if err == nil {
		t.Error("error charts must have error")
		return
	}
	// error overrideValue
	_, err = renderTpls(charts, chartName, "", "error overrideValue")
	if err == nil {
		t.Error("error overrideValue must have error")
		return
	}
	_, err = renderTpls(charts, chartName, "", fmt.Sprintf(overrideValue, "a b"))
	if err != nil {
		t.Error(err)
		return
	}
	// empty real name
	_, err = renderTpls(charts, "", "", "")
	if err != nil {
		t.Error(err)
		return
	}

	// normal
	resetControllerName := "reset-controller-name"
	result, err := renderTpls(charts, chartName, "", fmt.Sprintf(overrideValue, resetControllerName))
	if err != nil {
		t.Error(err)
		return
	}
	renderedTpls = result

	for _, tpl := range result {
		if strings.Contains(tpl, resetControllerName) {
			return
		}
	}
	t.Error("overrideValue is not found, check charts or overrideValue or renderTpls error")
}

func TestBuildK8sObjectWithRenderedTpls(t *testing.T) {
	// renderedTpls is nil
	k8sobjs := buildK8sObjectWithRenderedTpls(nil)
	if len(k8sobjs) > 0 {
		t.Error("nil renderedTpls must be empty k8sobj")
		return
	}

	// renderedTpls content is empty
	k8sobjs = buildK8sObjectWithRenderedTpls(map[string]string{
		"1": "",
		"2": "####",
	})
	if len(k8sobjs) > 0 {
		t.Error("empty yaml must be empty k8sobj")
		return
	}
	// error renderedTpls
	k8sobjs = buildK8sObjectWithRenderedTpls(map[string]string{
		"1": "error yaml",
	})
	if len(k8sobjs) > 0 {
		t.Error("error yaml must be empty k8sobj")
		return
	}

	k8sobjs = buildK8sObjectWithRenderedTpls(renderedTpls)
	for _, obj := range k8sobjs {
		t.Log(obj.YAML2String())
	}
}

func TestRenderTemplate(t *testing.T) {
	// error chart
	_, err := RenderTemplate([]byte("error chart"), "", "", "")
	if err == nil {
		t.Error("error char must have error")
		return
	}

	k8sobjs, err := RenderTemplate(charts, chartName, "", fmt.Sprintf(overrideValue, "RenderTemplate-controller"))
	if err != nil {
		t.Error(err)
		return
	}
	if len(k8sobjs) < 1 {
		t.Error("not render template charts")
		return
	}
}
