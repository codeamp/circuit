package dockerbuilder

import(
	docker "github.com/fsouza/go-dockerclient"
)

////////////////////////////

type Dockerer interface {
	NewClient(Socket string) (DockerClienter, error)
}

type LegitimateDocker struct {
}

func (l LegitimateDocker) NewClient(socket string) (DockerClienter, error){
	client, err := docker.NewClient(socket)
	legitimateDockerClient := LegitimateDockerClient{ client }
	return legitimateDockerClient, err
}

////////////////////////////

type DockerClienter interface {
	BuildImage(docker.BuildImageOptions) error
	PushImage(docker.PushImageOptions, docker.AuthConfiguration) error
	TagImage(string, docker.TagImageOptions) error
	InspectImage(string) (DockerImager, error)
}

type LegitimateDockerClient struct {
	Client *docker.Client
}

func (l LegitimateDockerClient) BuildImage(buildOptions docker.BuildImageOptions) error {
	return l.Client.BuildImage(buildOptions)
}

func (l LegitimateDockerClient) InspectImage(name string) (DockerImager, error) {
	image, err := l.Client.InspectImage(name)
	legitimateDockerImage := LegitimateDockerImage{ image }
	return legitimateDockerImage, err
}

func (l LegitimateDockerClient) TagImage(name string, tagImageOptions docker.TagImageOptions) error {
	return l.Client.TagImage(name, tagImageOptions)
}

func (l LegitimateDockerClient) PushImage(pushImageOptions docker.PushImageOptions, authConfig docker.AuthConfiguration) error {
	return l.Client.PushImage(pushImageOptions, authConfig)
}


////////////////////////////

type DockerImager interface {
	ID() string
}

type LegitimateDockerImage struct {
	Image *docker.Image
}

func (i LegitimateDockerImage) ID() string {
	return i.Image.ID
}