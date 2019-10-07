package gitsync

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/spf13/viper"
)

type GitSync struct {
	events chan transistor.Event
}

func init() {
	transistor.RegisterPlugin("gitsync", func() transistor.Plugin {
		return &GitSync{}
	}, plugins.GitSync{})
}

func (x *GitSync) Description() string {
	return "Sync Git repositories and create new features"
}

func (x *GitSync) SampleConfig() string {
	return ` `
}

func (x *GitSync) Start(e chan transistor.Event) error {
	x.events = e
	log.Info("Started GitSync")

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	// create global gitconfig file
	gitconfigPath := fmt.Sprintf("%s/.gitconfig", usr.HomeDir)
	if _, err := os.Stat(gitconfigPath); os.IsNotExist(err) {
		log.Warn("Local .gitconfig file not found! Writing default.")
		err = ioutil.WriteFile(gitconfigPath, []byte("[user]\n  name = codeamp \n  email = codeamp@codeamp.com"), 0600)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return nil
}

func (x *GitSync) Stop() {
	log.Info("Stopping GitSync")
}

func (x *GitSync) Subscribe() []string {
	return []string{
		"gitsync:create",
		"gitsync:update",
	}
}

func (x *GitSync) git(env []string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	defer cmd.Wait()
	//defer syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)

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

func (x *GitSync) toGitCommit(entry string, head bool) (plugins.GitCommit, error) {
	items := strings.Split(entry, "#@#")
	commiterDate, err := time.Parse("2006-01-02T15:04:05-07:00", items[4])

	if err != nil {
		return plugins.GitCommit{}, err
	}

	return plugins.GitCommit{
		Hash:       items[0],
		ParentHash: items[1],
		Message:    items[2],
		User:       items[3],
		Head:       head,
		Created:    commiterDate,
	}, nil
}

func (x *GitSync) commits(project plugins.Project, git plugins.Git) ([]plugins.GitCommit, error) {
	var err error
	var output []byte

	idRsaPath := fmt.Sprintf("%s/%s_id_rsa", viper.GetString("plugins.gitsync.workdir"), project.Repository)
	idRsa := fmt.Sprintf("GIT_SSH_COMMAND=ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i %s -F /dev/null", idRsaPath)
	repoPath := fmt.Sprintf("%s/%s_%s", viper.GetString("plugins.gitsync.workdir"), project.Repository, git.Branch)

	// Git Env
	env := os.Environ()
	env = append(env, idRsa)

	_, err = exec.Command("mkdir", "-p", filepath.Dir(repoPath)).CombinedOutput()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(idRsaPath); os.IsNotExist(err) {
		log.InfoWithFields("creating repository id_rsa", log.Fields{
			"path": idRsaPath,
		})

		err := ioutil.WriteFile(idRsaPath, []byte(git.RsaPrivateKey), 0600)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		log.InfoWithFields("cloning repository", log.Fields{
			"path": repoPath,
		})

		_, err := x.git(env, "clone", git.Url, repoPath)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}

	output, err = x.git(env, "-C", repoPath, "reset", "--hard", fmt.Sprintf("origin/%s", git.Branch))
	if err != nil {
		log.Error(err)
		return nil, err
	}

	output, err = x.git(env, "-C", repoPath, "clean", "-fd")
	if err != nil {
		log.Error(err)
		return nil, err
	}

	output, err = x.git(env, "-C", repoPath, "pull", "origin", git.Branch)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	output, err = x.git(env, "-C", repoPath, "checkout", git.Branch)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	output, err = x.git(env, "-C", repoPath, "log", "--first-parent", "--date=iso-strict", "-n", "50", "--pretty=format:%H#@#%P#@#%s#@#%cN#@#%cd", git.Branch)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	var commits []plugins.GitCommit
	for i, line := range strings.Split(strings.TrimSuffix(string(output), "\n"), "\n") {
		head := false
		if i == 0 {
			head = true
		}
		commit, err := x.toGitCommit(line, head)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

func (x *GitSync) Process(e transistor.Event) error {
	if e.Event() == "gitsync:create" {
		payload := e.Payload.(plugins.GitSync)
		event := e.NewEvent(transistor.GetAction("status"), transistor.GetState("running"), "Fetching resource")
		x.events <- event

		commits, err := x.commits(payload.Project, payload.Git)
		if err != nil {
			log.Error(err)

			errEvent := e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), fmt.Sprintf("%v (Action: %v)", err.Error(), "failed"))
			x.events <- errEvent

			return err
		}

		var _commits []plugins.GitCommit
		for i := range commits {
			c := commits[i]
			c.Repository = payload.Project.Repository
			c.Ref = fmt.Sprintf("refs/heads/%s", payload.Git.Branch)

			_commits = append(_commits, c)

			if c.Hash == payload.From {
				break
			}
		}

		payload.Commits = _commits

		event = e.NewEvent(transistor.GetAction("status"), transistor.GetState("complete"), "Operation Complete")
		event.SetPayload(payload)

		x.events <- event
	}

	if e.Event() == "gitsync:update" {
		log.ErrorWithFields("Event received by githubsync yet unhandled!", log.Fields{"event": e})
	}

	return nil
}
