package dockerbuilder

import (
	"github.com/spf13/viper"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/extemporalgenome/slug"
	docker "github.com/fsouza/go-dockerclient"
)

type DockerBuilder struct {
	events chan transistor.Event
	Socket string
}

func init() {
	transistor.RegisterPlugin("dockerbuilder", func() transistor.Plugin {
		return &DockerBuilder{Socket: "unix:///var/run/docker.sock"}
	})
}

func (x *DockerBuilder) Description() string {
	return "Clone git repository and build a docker image"
}

func (x *DockerBuilder) SampleConfig() string {
	return ` `
}

func (x *DockerBuilder) Start(e chan transistor.Event) error {
	x.events = e
	log.Info("Started DockerBuilder")

	return nil
}

func (x *DockerBuilder) Stop() {
	log.Println("Stopping DockerBuilder")
}

func (x *DockerBuilder) Subscribe() []string {
	return []string{
		"plugins.ReleaseExtension:create:dockerbuilder",
		"plugins.ReleaseExtension:update:dockerbuilder",
		"plugins.Extension:create:dockerbuilder",
	}
}

func (x *DockerBuilder) git(env []string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)

	log.InfoWithFields("executing command", log.Fields{
		"path": cmd.Path,
		"args": strings.Join(cmd.Args, " "),
	})

	cmd.Env = env

	out, err := cmd.CombinedOutput()
	if err != nil {
		if ee, ok := err.(*exec.Error); ok {
			if ee.Err == exec.ErrNotFound {
				return nil, errors.New("Git executable not found in $PATH")
			}
		}

		return nil, errors.New(string(bytes.TrimSpace(out)))
	}

	return out, nil
}

func (x *DockerBuilder) bootstrap(repoPath string, event plugins.ReleaseExtension) error {
	var err error
	var output []byte

	// idRsaPath := fmt.Sprintf("%s/%s_id_rsa", event.Release.Git.Workdir, event.Release.Project.Repository)
	idRsaPath := fmt.Sprintf("%s/%s_id_rsa", viper.GetString("plugins.dockerbuilder.workdir"), event.Release.Project.Repository)
	repoPath = fmt.Sprintf("%s/%s", viper.GetString("plugins.dockerbuilder.workdir"), event.Release.Project.Repository)
	idRsa := fmt.Sprintf("GIT_SSH_COMMAND=ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i %s -F /dev/null", idRsaPath)

	// Git Env
	env := os.Environ()
	env = append(env, idRsa)

	_, err = exec.Command("mkdir", "-p", filepath.Dir(repoPath)).CombinedOutput()
	if err != nil {
		return err
	}

	if _, err := os.Stat(idRsaPath); os.IsNotExist(err) {
		log.InfoWithFields("creating repository id_rsa", log.Fields{
			"path": idRsaPath,
		})

		err := ioutil.WriteFile(idRsaPath, []byte(event.Release.Git.RsaPrivateKey), 0600)
		if err != nil {
			log.Debug(err)
			return err
		}
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		log.InfoWithFields("cloning repository", log.Fields{
			"path": repoPath,
		})

		output, err := x.git(env, "clone", event.Release.Git.Url, repoPath)
		if err != nil {
			log.Debug(err)
			return err
		}
		log.Info(string(output))
	}

	output, err = x.git(env, "-C", repoPath, "pull", "origin", event.Release.Git.Branch)
	if err != nil {
		log.Info(err)
		return err
	}
	log.Info(string(output))

	output, err = x.git(env, "-C", repoPath, "checkout", event.Release.Git.Branch)
	if err != nil {
		log.Info(err)
		return err
	}
	log.Info(string(output))

	return nil
}

func (x *DockerBuilder) build(repoPath string, event plugins.ReleaseExtension, dockerBuildOut io.Writer) error {
	idRsaPath := fmt.Sprintf("%s/%s", viper.GetString("plugins.dockerbuilder.workdir"), event.Release.Project.Repository)	
	gitArchive := exec.Command("git", "archive", event.Release.HeadFeature.Hash)
	gitArchive.Dir = idRsaPath
	gitArchiveOut, err := gitArchive.StdoutPipe()
	if err != nil {
		log.Debug(err)
		return err
	}

	gitArchiveErr, err := gitArchive.StderrPipe()
	if err != nil {
		log.Debug(err)
		return err
	}

	err = gitArchive.Start()
	if err != nil {
		log.Fatal(err)
		return err
	}

	dockerBuildIn := bytes.NewBuffer(nil)
	go func() {
		io.Copy(os.Stderr, gitArchiveErr)
	}()

	io.Copy(dockerBuildIn, gitArchiveOut)

	err = gitArchive.Wait()
	if err != nil {
		log.Debug(err)
		return err
	}

	buildArgs := []docker.BuildArg{}
	for _, secret := range event.Extension.Project.Secrets {
		if(secret.Type == plugins.GetType("build-arg")){
			ba := docker.BuildArg{
				Name:  secret.Key,
				Value: secret.Value,
			}
			buildArgs = append(buildArgs, ba)
		}
	}
	fullImagePath := fullImagePath(event)
	buildOptions := docker.BuildImageOptions{
		Dockerfile:   fmt.Sprintf("Dockerfile"),
		Name:         fullImagePath,
		ContextDir:   fmt.Sprintf("%s/%s", viper.GetString("plugins.dockerbuilder.workdir"), event.Release.Project.Repository),		
		OutputStream: dockerBuildOut,
		// InputStream:  dockerBuildIn,
		BuildArgs:    buildArgs,
	}

	dockerClient, err := docker.NewClient(x.Socket)
	if err != nil {
		log.Debug(err)
		return err
	}

	err = dockerClient.BuildImage(buildOptions)
	if err != nil {
		log.Debug(err)
		return err
	}

	return nil
}

func (x *DockerBuilder) push(repoPath string, event plugins.ReleaseExtension, buildlog io.Writer) error {
	var err error

	buildlog.Write([]byte(fmt.Sprintf("Pushing %s\n", imagePathGen(event))))

	dockerClient, err := docker.NewClient(x.Socket)
	err = dockerClient.PushImage(docker.PushImageOptions{
		Name:         imagePathGen(event),
		Tag:          imageTagGen(event),
		OutputStream: buildlog,
	}, docker.AuthConfiguration{
		Username: event.Extension.Config["DOCKERBUILDER_USER"].(string),
		Password: event.Extension.Config["DOCKERBUILDER_PASSWORD"].(string),
		Email:    event.Extension.Config["DOCKERBUILDER_EMAIL"].(string),
	})
	if err != nil {
		return err
	}

	tagOptions := docker.TagImageOptions{
		Repo:  imagePathGen(event),
		Tag:   imageTagLatest(event),
		Force: true,
	}

	fullImagePath := imagePathGen(event) + ":" + imageTagGen(event)

	if err = dockerClient.TagImage(fullImagePath, tagOptions); err != nil {
		return err
	}

	err = dockerClient.PushImage(docker.PushImageOptions{
		Name:         imagePathGen(event),
		Tag:          imageTagLatest(event),
		OutputStream: buildlog,
	}, docker.AuthConfiguration{
		Username: event.Extension.Config["DOCKERBUILDER_USER"].(string),
		Password: event.Extension.Config["DOCKERBUILDER_PASSWORD"].(string),
		Email:    event.Extension.Config["DOCKERBUILDER_EMAIL"].(string),
	})
	if err != nil {
		return err
	}

	return nil
}

func (x *DockerBuilder) Process(e transistor.Event) error {
	if e.Name == "plugins.Extension:create:dockerbuilder" {
		var extensionEvent plugins.Extension
		extensionEvent = e.Payload.(plugins.Extension)
		extensionEvent.Action = plugins.GetAction("status")
		extensionEvent.State = plugins.GetState("complete")
		x.events <- e.NewEvent(extensionEvent, nil)
		return nil
	}
	
	if e.Name == "plugins.Extension:update:dockerbuilder" {
		var extensionEvent plugins.Extension
		extensionEvent = e.Payload.(plugins.Extension)
		extensionEvent.Action = plugins.GetAction("status")
		extensionEvent.State = plugins.GetState("complete")
		x.events <- e.NewEvent(extensionEvent, nil)
		return nil
	}	

	event := e.Payload.(plugins.ReleaseExtension)

	var err error

	event.Action = plugins.GetAction("status")
	event.State = plugins.GetState("fetching")
	event.StateMessage = ""
	x.events <- e.NewEvent(event, nil)

	// repoPath := fmt.Sprintf("%s/%s_%s", event.Release.Git.Workdir, event.Release.Project.Repository, event.Release.Git.Branch)
	repoPath := fmt.Sprintf("%s", event.Release.Project.Repository)

	buf := bytes.NewBuffer(nil)
	buildlog := io.MultiWriter(buf, os.Stdout)

	err = x.bootstrap(repoPath, event)
	if err != nil {
		log.Debug(err)
		event.State = plugins.GetState("failed")
		event.StateMessage = fmt.Sprintf("%v (Action: %v, Step: bootstrap)", err.Error(), event.State)
		x.events <- e.NewEvent(event, nil)
		return err
	}

	err = x.build(repoPath, event, buildlog)
	if err != nil {
		log.Debug(err)
		event.State = plugins.GetState("failed")
		event.StateMessage = fmt.Sprintf("%v (Action: %v, Step: build)", err.Error(), event.State)
		//event.BuildLog = buildlog.String()
		x.events <- e.NewEvent(event, nil)
		return err
	}

	err = x.push(repoPath, event, buildlog)
	if err != nil {
		log.Debug(err)
		event.State = plugins.GetState("failed")
		event.StateMessage = fmt.Sprintf("%v (Action: %v, Step: push)", err.Error(), event.State)
		// event.BuildLog = buildlog.String()
		x.events <- e.NewEvent(event, nil)
		return err
	}

	event.State = plugins.GetState("complete")
	event.Artifacts["IMAGE"] = fullImagePath(event)
	event.Artifacts["USER"] = event.Extension.Config["DOCKERBUILDER_USER"].(string)
	event.Artifacts["PASSWORD"] = event.Extension.Config["DOCKERBUILDER_PASSWORD"].(string)
	event.Artifacts["EMAIL"] = event.Extension.Config["DOCKERBUILDER_EMAIL"].(string)
	event.Artifacts["HOST"] = event.Extension.Config["DOCKERBUILDER_HOST"].(string)
	event.StateMessage = "Completed"
	// event.BuildLog = buildlog.String()
	x.events <- e.NewEvent(event, nil)
	return nil
}

// generate image tag name
func imageTagGen(event plugins.ReleaseExtension) string {
	return (fmt.Sprintf("%s.%s", event.Release.HeadFeature.Hash, event.Release.Environment))
}

func imageTagLatest(event plugins.ReleaseExtension) string {
	if event.Release.Environment == "production" {
		return ("latest")
	}
	return (fmt.Sprintf("%s.%s", "latest", event.Release.Environment))
}

// rengerate image path name
func imagePathGen(event plugins.ReleaseExtension) string {
	registryHost := event.Extension.Config["DOCKERBUILDER_HOST"]
	registryOrg := event.Extension.Config["DOCKERBUILDER_ORG"]
	return (fmt.Sprintf("%s/%s/%s", registryHost, registryOrg, slug.Slug(event.Release.Project.Repository)))
}

// return the full image path with featureHash tag
func fullImagePath(event plugins.ReleaseExtension) string {
	return (fmt.Sprintf("%s:%s", imagePathGen(event), imageTagGen(event)))
}
