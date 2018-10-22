package dockerbuilder

import(
	docker "github.com/fsouza/go-dockerclient"
)

////////////////////////////

type Dockerer interface {
	NewClient(Socket string) (DockerClienter, error)
}

type Docker struct {
}

func (l Docker) NewClient(socket string) (DockerClienter, error){
	client, err := docker.NewClient(socket)
	DockerClient := DockerClient{ client }
	return DockerClient, err
}

////////////////////////////

type DockerClienter interface {
	BuildImage(docker.BuildImageOptions) error
	PushImage(docker.PushImageOptions, docker.AuthConfiguration) error
	TagImage(string, docker.TagImageOptions) error
	InspectImage(string) (DockerImager, error)
}

type DockerClient struct {
	Client *docker.Client
}

func (l DockerClient) BuildImage(buildOptions docker.BuildImageOptions) error {
	return l.Client.BuildImage(buildOptions)
}

func (l DockerClient) InspectImage(name string) (DockerImager, error) {
	image, err := l.Client.InspectImage(name)
	DockerImage := DockerImage{ image }
	return DockerImage, err
}

func (l DockerClient) TagImage(name string, tagImageOptions docker.TagImageOptions) error {
	return l.Client.TagImage(name, tagImageOptions)
}

func (l DockerClient) PushImage(pushImageOptions docker.PushImageOptions, authConfig docker.AuthConfiguration) error {
	return l.Client.PushImage(pushImageOptions, authConfig)
}


////////////////////////////

type DockerImager interface {
	ID() string
}

type DockerImage struct {
	Image *docker.Image
}

func (i DockerImage) ID() string {
	return i.Image.ID
}