package generators

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"
	"text/template"

	"github.com/mitchellh/ioprogress"
	"gopkg.in/yaml.v2"
)

func GenerateReleaseJobsPackage(releaseURL string, cacheDir string) (err error) {
	filename := downloadFromURL(releaseURL, cacheDir)
	processFile(filename)
	return
}

func downloadFromURL(url string, cacheDir string) (filename string) {
	name := path.Base(url)
	filename = cacheDir + "/" + name

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Println("Could not find release in local cache. Downloading now.")
		out, _ := os.Create(filename)
		resp, _ := http.Get(url)
		defer resp.Body.Close()

		progressR := &ioprogress.Reader{
			Reader: resp.Body,
			Size:   resp.ContentLength,
		}
		io.Copy(out, progressR)
	}
	return
}

func processFile(srcFile string) {
	f, err := os.Open(srcFile)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	tarReader := getTarballReader(f)
	i := 0

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}
		name := header.Name

		switch header.Typeflag {
		case tar.TypeReg:
			if strings.HasPrefix(name, "./jobs/") {
				jobTarball := getTarballReader(tarReader)
				jobManifest := getJobManifestFromTarball(jobTarball)
				processJobManifest(jobManifest, name)
			}
		}
		i++
	}
}

func getJobManifestFromTarball(jobTarball *tar.Reader) (res *tar.Reader) {
	var jobManifestFilename = "./job.MF"

	for {
		header, _ := jobTarball.Next()
		if header.Name == jobManifestFilename {
			res = jobTarball
			break
		}
	}
	return
}

func processJobManifest(jobTarball io.Reader, tarballFilename string) {
	var elements []elementStruct
	var defaultElementType = "interface{}"
	buf := new(bytes.Buffer)
	buf.ReadFrom(jobTarball)
	manifestYaml := JobManifest{}
	yaml.Unmarshal(buf.Bytes(), &manifestYaml)

	for k, v := range manifestYaml.Properties {
		myType := defaultElementType

		if v.Default != nil {
			myType = fmt.Sprint(reflect.ValueOf(v.Default).Type())
		}
		elements = append(elements, elementStruct{
			ElementName:     parseElementName(k),
			ElementType:     myType,
			ElementYamlName: k,
		})
	}
	jobName := strings.Split(path.Base(tarballFilename), ".")[0]
	jobName = strings.ToUpper(jobName[:1]) + jobName[1:]
	job := jobStructTemplate{
		JobName:  jobName,
		Elements: elements,
	}
	tmpl, err := template.New("job").Parse(structTemplate)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, job)
	if err != nil {
		panic(err)
	}
}

func parseElementName(name string) string {
	f := strings.FieldsFunc(name, func(r rune) bool {
		return r == '_' || r == '.'
	})
	for i := range f {
		f[i] = strings.ToUpper(f[i][:1]) + f[i][1:]
	}
	return strings.Join(f, "")
}

func getTarballReader(reader io.Reader) *tar.Reader {
	gzf, err := gzip.NewReader(reader)

	if err != nil {
		fmt.Println(err)
	}
	return tar.NewReader(gzf)
}
