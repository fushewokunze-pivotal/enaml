#Enaml
## Because (EN)ough with the y(AML) already

[![wercker status](https://app.wercker.com/status/7f56eae6e591609ebab43d47a7a8a8a3/s/master "wercker status")](https://app.wercker.com/project/bykey/7f56eae6e591609ebab43d47a7a8a8a3)

### Intent
- deployment manifests as testable code
- so no one has to write another bosh deployment manifest in yaml again.

### Sample


### how to use the enaml core to develop OMG product plugins

Download the latest binary release for your OS from: https://github.com/enaml-ops/enaml/releases/latest

*create golang structs for job properties from a release*

```bash

# user @ MacBook-Pro in ~/stuff/sample [00:00:00]
$ wget https://github.com/enaml-ops/enaml/releases/download/v0.0.11/enaml


# user @ MacBook-Pro in ~/stuff/sample [00:00:00]
$ #target the bosh release you with to enamlize
$ enaml generate https://bosh.io/d/github.com/cf-platform-eng/docker-boshrelease\?v\=27
Could not find release in local cache. Downloading now.
136929594/136929594
completed generating release job structs for  https://bosh.io/d/github.com/cf-platform-eng/docker-boshrelease?v=27

# user @ MacBook-Pro in ~/stuff/sample [00:00:00]
$ ls
enaml-gen

# user @ MacBook-Pro in ~/stuff/sample [00:00:00]
$ #golang packages have been generated for all job properties in the give release version
$ ls enaml-gen/
broker-deregistrar      cf-containers-broker    docker                  swarm_agent
broker-registrar        containers              fetch-containers-images swarm_manager
```

*use the generated structs in your plugin*

```go
package main

import (
	"github.com/enaml-ops/pluginlib/pcli"
	"github.com/enaml-ops/pluginlib/product"
	"github.com/enaml-ops/stuff/sample/enaml-gen/docker"
)

func main() {
	product.Run(&MyProduct{
		DockerRef: new(docker.Docker),
	})
}

type MyProduct struct{
	DockerRef docker.Docker
}

func (s *MyProduct) GetProduct(args []string, cloudconfig []byte) []byte {
	return []byte("")
}

func (s *MyProduct) GetFlags() (flags []pcli.Flag) {
	return
}

func (s *MyProduct) GetMeta() product.Meta {
	return product.Meta{
		Name: "myfakeproduct",
	}
}
```




### maybe you've got a manifest but dont know how to maintain it (ie. key/cert/pass rotation, or automated component scaling, etc)
```go
package main

import (
    "fmt"
    "io/ioutil"
    "os"

    "github.com/enaml-ops/enaml"
)

//this will take a manifest scale its instance counts and return a new manifest
func main() {
    originalFileBytes, _ := ioutil.ReadFile(os.Args[1])
    enamlizedManifest := enaml.NewDeploymentManifest(originalFileBytes)

    for i, job := range enamlizedManifest.Jobs {
        job.Instances += 1
        enamlizedManifest.Jobs[i] = job
    }
    ymlBytes, _ := enaml.Paint(enamlizedManifest)
    fmt.Println(string(ymlBytes))
}


#then call it
$> go run main.go my-cf-manifest.yml > my-scaled-cf-manifest.yml
```


### how your deployment could look
```go
package concourse

import (
	"github.com/enaml-ops/enaml"
	"github.com/enaml-ops/enaml-concourse-sample/releasejobs"
)

var (
	DefaultName   = "concourse"
	DirectorUUID  = "asdfasdfasdf"
	StemcellAlias = "trusty"
)

func main() {
	enaml.Paint(NewDeployment())
}

type Deployment struct {
	enaml.Deployment
	Manifest *enaml.DeploymentManifest
}

func NewDeployment() (d Deployment) {
	d = Deployment{}
	d.Manifest = new(enaml.DeploymentManifest)
	d.Manifest.SetName(DefaultName)
	d.Manifest.SetDirectorUUID(DirectorUUID)
	d.Manifest.AddReleaseByName("concourse")
	d.Manifest.AddReleaseByName("garden-linux")
	d.Manifest.AddStemcellByName("ubuntu-trusty", StemcellAlias)
	web := enaml.NewInstanceGroup("web", 1, "web", StemcellAlias)
	web.AddAZ("z1")
	web.AddNetwork(enaml.Network{"name": "private"})
	atc := enaml.NewInstanceJob("atc", "concourse", releasejobs.Atc{
		ExternalUrl:        "something",
		BasicAuthUsername:  "user",
		BasicAuthPassword:  "password",
		PostgresqlDatabase: "&atc_db atc",
	})
	tsa := enaml.NewInstanceJob("tsa", "concourse", releasejobs.Tsa{})
	web.AddJob(atc)
	web.AddJob(tsa)
	db := enaml.NewInstanceGroup("db", 1, "database", StemcellAlias)
	worker := enaml.NewInstanceGroup("worker", 1, "worker", StemcellAlias)
	d.Manifest.AddInstanceGroup(web)
	d.Manifest.AddInstanceGroup(db)
	d.Manifest.AddInstanceGroup(worker)
	return
}

func (s Deployment) GetDeployment() enaml.DeploymentManifest {
	return *s.Manifest
}
```

### Development

Enaml uses [Glide](https://github.com/Masterminds/glide) to manage vendored Go
dependencies. Glide is a tool for managing the vendor directory within a Go
package. As such, Golang 1.6+ is recommended.

1. If you haven't done so already, install glide and configure your GOPATH.
2. Clone `enaml` to your GOPATH: `git clone https://github.com/enaml-ops/enaml $GOPATH/src/github.com/enaml-ops/enaml`
3. Install dependencies: `cd $GOPATH/src/github.com/enaml-ops/enaml && glide install`
4. Run the enaml tests `go test $(glide novendor)`
5. Build the `enaml` executable `go build -o $GOPATH/bin/enaml cmd/enaml/main.go`
