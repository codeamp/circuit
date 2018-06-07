package dockerbuilder

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/extemporalgenome/slug"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/spf13/viper"
)

type DockerBuilder struct {
	events chan transistor.Event
	Socket string
}

func init() {
	transistor.RegisterPlugin("dockerbuilder", func() transistor.Plugin {
		return &DockerBuilder{Socket: "unix:///var/run/docker.sock"}
	}, plugins.ReleaseExtension{})
}

func (x *DockerBuilder) Description() string {
	return "Clone git repository and build a docker image"
}

func (x *DockerBuilder) SampleConfig() string {
	return ` `
}

func (x *DockerBuilder) Start(e chan transistor.Event) error {
	x.events = e

	// create global gitconfig file
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	gitconfigPath := fmt.Sprintf("%s/.gitconfig", usr.HomeDir)
	if _, err := os.Stat(gitconfigPath); os.IsNotExist(err) {
		log.Warn("Local .gitconfig file not found! Writing default.")
		err = ioutil.WriteFile(gitconfigPath, []byte("[user]\n  name = codeamp \n  email = codeamp@codeamp.com"), 0600)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	log.Info("Started DockerBuilder")
	return nil
}

func (x *DockerBuilder) Stop() {
	log.Info("Stopping DockerBuilder")
}

func (x *DockerBuilder) Subscribe() []string {
	return []string{
		"project:dockerbuilder:create",
		"project:dockerbuilder:update",
		"release:dockerbuilder:create",
	}
}

func (x *DockerBuilder) git(env []string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)

	log.DebugWithFields("executing command", log.Fields{
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

func (x *DockerBuilder) bootstrap(repoPath string, event transistor.Event) error {
	payload := event.Payload.(plugins.ReleaseExtension)

	var err error
	var output []byte

	// idRsaPath := fmt.Sprintf("%s/%s_id_rsa", payload.Release.Git.Workdir, payload.Release.Project.Repository)
	idRsaPath := fmt.Sprintf("%s/%s_id_rsa", viper.GetString("plugins.dockerbuilder.workdir"), payload.Release.Project.Repository)
	repoPath = fmt.Sprintf("%s/%s_%s", viper.GetString("plugins.dockerbuilder.workdir"), payload.Release.Project.Repository, payload.Release.Git.Branch)
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

		err := ioutil.WriteFile(idRsaPath, []byte(payload.Release.Git.RsaPrivateKey), 0600)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		log.InfoWithFields("cloning repository", log.Fields{
			"path": repoPath,
		})

		output, err := x.git(env, "clone", payload.Release.Git.Url, repoPath)
		if err != nil {
			log.Error(err)
			return err
		}
		log.Info(string(output))
	}

	output, err = x.git(env, "-C", repoPath, "reset", "--hard", fmt.Sprintf("origin/%s", payload.Release.Git.Branch))
	if err != nil {
		log.Error(err)
		return err
	}

	log.Info(string(output))

	output, err = x.git(env, "-C", repoPath, "clean", "-fd")
	if err != nil {
		log.Error(err)
		return err
	}

	output, err = x.git(env, "-C", repoPath, "checkout", payload.Release.Git.Branch)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Info(string(output))

	output, err = x.git(env, "-C", repoPath, "pull", "origin", payload.Release.Git.Branch)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Info(string(output))

	return nil
}

func (x *DockerBuilder) build(repoPath string, event transistor.Event, dockerBuildOut io.Writer) error {
	payload := event.Payload.(plugins.ReleaseExtension)

	repoPath = fmt.Sprintf("%s/%s_%s", viper.GetString("plugins.dockerbuilder.workdir"), payload.Release.Project.Repository, payload.Release.Git.Branch)
	gitArchive := exec.Command("git", "archive", payload.Release.HeadFeature.Hash)
	gitArchive.Dir = repoPath
	gitArchiveOut, err := gitArchive.StdoutPipe()
	if err != nil {
		log.Error(err)
		return err
	}

	gitArchiveErr, err := gitArchive.StderrPipe()
	if err != nil {
		log.Error(err)
		return err
	}

	err = gitArchive.Start()
	if err != nil {
		log.Fatal(err)
		return err
	}

	user, err := event.GetArtifact("user")
	if err != nil {
		return err
	}

	password, err := event.GetArtifact("password")
	if err != nil {
		return err
	}

	email, err := event.GetArtifact("email")
	if err != nil {
		return err
	}

	registryHost, err := event.GetArtifact("host")
	if err != nil {
		log.Error(err)
		return err
	}

	dockerBuildIn := bytes.NewBuffer(nil)
	go func() {
		io.Copy(os.Stderr, gitArchiveErr)
	}()

	io.Copy(dockerBuildIn, gitArchiveOut)

	err = gitArchive.Wait()
	if err != nil {
		log.Error(err)
		return err
	}

	buildArgs := []docker.BuildArg{}
	for _, secret := range payload.Release.Secrets {
		if secret.Type == plugins.GetType("build") {
			ba := docker.BuildArg{
				Name:  secret.Key,
				Value: secret.Value,
			}
			buildArgs = append(buildArgs, ba)
		}
	}
	fullImagePath := fullImagePath(event)
	authConfigs := docker.AuthConfigurations{
		Configs: map[string]docker.AuthConfiguration{
			registryHost.String(): {
				Username:      user.String(),
				Password:      password.String(),
				Email:         email.String(),
				ServerAddress: registryHost.String(),
			},
		},
	}
	buildOptions := docker.BuildImageOptions{
		Dockerfile:   fmt.Sprintf("Dockerfile"),
		Name:         fullImagePath,
		OutputStream: dockerBuildOut,
		InputStream:  dockerBuildIn,
		BuildArgs:    buildArgs,
		AuthConfigs:  authConfigs,
	}

	dockerClient, err := docker.NewClient(x.Socket)
	if err != nil {
		log.Error(err)
		return err
	}

	err = dockerClient.BuildImage(buildOptions)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (x *DockerBuilder) push(repoPath string, event transistor.Event, buildlog io.Writer) error {
	var err error

	buildlog.Write([]byte(fmt.Sprintf("Pushing %s\n", imagePathGen(event))))

	user, err := event.GetArtifact("user")
	if err != nil {
		return err
	}

	password, err := event.GetArtifact("password")
	if err != nil {
		return err
	}

	email, err := event.GetArtifact("email")
	if err != nil {
		return err
	}

	dockerClient, err := docker.NewClient(x.Socket)
	err = dockerClient.PushImage(docker.PushImageOptions{
		Name:         imagePathGen(event),
		Tag:          imageTagGen(event),
		OutputStream: buildlog,
	}, docker.AuthConfiguration{
		Username: user.String(),
		Password: password.String(),
		Email:    email.String(),
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
		Username: user.String(),
		Password: password.String(),
		Email:    email.String(),
	})
	if err != nil {
		return err
	}

	return nil
}

func (x *DockerBuilder) Process(e transistor.Event) error {
	if e.Matches("project:dockerbuilder") {
		if e.Action == transistor.GetAction("create") {
			ev := e.NewEvent(transistor.GetAction("status"), transistor.GetState("complete"), "Installation complete.")
			x.events <- ev
			return nil
		}

		if e.Action == transistor.GetAction("update") {
			ev := e.NewEvent(transistor.GetAction("status"), transistor.GetState("complete"), "Update complete.")
			x.events <- ev
			return nil
		}
	}

	if e.Matches("release:dockerbuilder") {
		x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("running"), "Fetching resources.")
		payload := e.Payload.(plugins.ReleaseExtension)

		repoPath := fmt.Sprintf("%s", payload.Release.Project.Repository)
		buildlogBuf := bytes.NewBuffer(nil)
		buildlog := io.MultiWriter(buildlogBuf, os.Stdout)

		var err error
		err = x.bootstrap(repoPath, e)
		if err != nil {
			log.Error(err)
			x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), fmt.Sprintf("%v (Action: %v, Step: bootstrap)", err.Error(), e.State))
			return err
		}

		err = x.build(repoPath, e, buildlog)
		if err != nil {
			log.Error(err)

			ev := e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), fmt.Sprintf("%v (Action: %v, Step: build)", err.Error(), e.State))
			ev.AddArtifact("build_log", buildlogBuf.String(), false)
			x.events <- ev

			return err
		}

		err = x.push(repoPath, e, buildlog)
		if err != nil {
			log.Error(err)

			ev := e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), fmt.Sprintf("%v (Action: %v, Step: push)", err.Error(), e.State))
			ev.AddArtifact("build_log", buildlogBuf.String(), false)
			x.events <- ev

			return err
		}

		user, err := e.GetArtifact("user")
		if err != nil {
			return err
		}

		password, err := e.GetArtifact("password")
		if err != nil {
			return err
		}

		email, err := e.GetArtifact("email")
		if err != nil {
			return err
		}

		registryHost, err := e.GetArtifact("host")
		if err != nil {
			log.Error(err)
		}

		ev := e.NewEvent(transistor.GetAction("status"), transistor.GetState("complete"), "Completed")
		ev.AddArtifact("user", user.String(), user.Secret)
		ev.AddArtifact("password", password.String(), password.Secret)
		ev.AddArtifact("email", email.String(), email.Secret)
		ev.AddArtifact("host", registryHost.String(), registryHost.Secret)
		ev.AddArtifact("image", fullImagePath(e), false)
		ev.AddArtifact("build_log", buildlogBuf.String(), false)
		x.events <- ev
	}

	return nil
}

// generate image tag name
func imageTagGen(event transistor.Event) string {
	payload := event.Payload.(plugins.ReleaseExtension)
	return (fmt.Sprintf("%s.%s", payload.Release.HeadFeature.Hash, payload.Release.Environment))
}

func imageTagLatest(event transistor.Event) string {
	payload := event.Payload.(plugins.ReleaseExtension)
	if payload.Release.Environment == "production" {
		return ("latest")
	}
	return (fmt.Sprintf("%s.%s", "latest", payload.Release.Environment))
}

// rengerate image path name
func imagePathGen(event transistor.Event) string {
	registryHost, err := event.GetArtifact("host")
	if err != nil {
		log.Error(err)
	}

	registryOrg, err := event.GetArtifact("org")
	if err != nil {
		log.Error(err)
	}

	payload := event.Payload.(plugins.ReleaseExtension)
	return (fmt.Sprintf("%s/%s/%s", registryHost.String(), registryOrg.String(), slug.Slug(payload.Release.Project.Repository)))
}

// return the full image path with featureHash tag
func fullImagePath(event transistor.Event) string {
	return (fmt.Sprintf("%s:%s", imagePathGen(event), imageTagGen(event)))
}
