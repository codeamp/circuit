package dockerbuilder_mock

import (
	"github.com/codeamp/circuit/plugins/dockerbuilder"
	docker "github.com/fsouza/go-dockerclient"
)

////////////////////////////

type MockedDocker struct {
}

func (l MockedDocker) NewClient(socket string) (dockerbuilder.DockerClienter, error) {
	return MockedDockerClient{}, nil
}

////////////////////////////

type MockedDockerClient struct {
}

func (l MockedDockerClient) BuildImage(buildOptions docker.BuildImageOptions) error {
	return nil
}

func (l MockedDockerClient) InspectImage(name string) (dockerbuilder.DockerImager, error) {
	return MockedDockerImage{}, nil
}

func (l MockedDockerClient) TagImage(name string, tagImageOptions docker.TagImageOptions) error {
	return nil
}

func (l MockedDockerClient) PushImage(pushImageOptions docker.PushImageOptions, authConfig docker.AuthConfiguration) error {
	return nil
}

////////////////////////////

type MockedDockerImage struct {
}

func (i MockedDockerImage) ID() string {
	return "HELLO-THIS-IS-A-MOCKED-IDENTIFIER"
}
