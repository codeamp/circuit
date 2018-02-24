package gitsync

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	})
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

	return nil
}

func (x *GitSync) Stop() {
	log.Println("Stopping GitSync")
}

func (x *GitSync) Subscribe() []string {
	return []string{
		"plugins.GitPing",
		"plugins.GitSync:update",
	}
}

func (x *GitSync) git(env []string, args ...string) ([]byte, error) {
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

func (x *GitSync) toGitCommit(entry string) (plugins.GitCommit, error) {
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
		Created:    commiterDate,
	}, nil
}

func (x *GitSync) branches(project plugins.Project, git plugins.Git) ([]plugins.GitBranch, error) {
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
			log.Debug(err)
			return nil, err
		}
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		log.InfoWithFields("cloning repository", log.Fields{
			"path": repoPath,
		})

		output, err := x.git(env, "clone", git.Url, repoPath)
		if err != nil {
			log.Debug(err)
			return nil, err
		}
		log.Info(string(output))
	}

	_, err = x.git(env, "-C", repoPath, "fetch", "--all")
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	output, err = x.git(env, "-C", repoPath, "ls-remote", "--heads")
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	var branches []plugins.GitBranch

	for _, line := range strings.Split(string(output), "\n") {
		for idx, branch := range strings.Split(line, "refs/heads/") {
			if idx%2 == 1 {
				branches = append(branches, plugins.GitBranch{
					Repository: project.Repository,
					Name:       branch,
				})
			}
		}
	}

	return branches, nil
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
			log.Debug(err)
			return nil, err
		}
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		log.InfoWithFields("cloning repository", log.Fields{
			"path": repoPath,
		})

		output, err := x.git(env, "clone", git.Url, repoPath)
		if err != nil {
			log.Debug(err)
			return nil, err
		}
		log.Info(string(output))
	}

	// output, err = x.git(env, "-C", repoPath, "reset", "", git.Branch)
	// if err != nil {
	// 	log.Debug(err)
	// 	return nil, err
	// }
	// log.Info(string(output))

	output, err = x.git(env, "-C", repoPath, "pull", "origin", git.Branch)
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	log.Info(string(output))

	output, err = x.git(env, "-C", repoPath, "checkout", git.Branch)
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	log.Info(string(output))

	output, err = x.git(env, "-C", repoPath, "log", "--first-parent", "--date=iso-strict", "-n", "50", "--pretty=format:%H#@#%P#@#%s#@#%cN#@#%cd", git.Branch)

	if err != nil {
		log.Debug(err)
		return nil, err
	}

	var commits []plugins.GitCommit

	for _, line := range strings.Split(strings.TrimSuffix(string(output), "\n"), "\n") {
		commit, err := x.toGitCommit(line)
		if err != nil {
			log.Debug(err)
			return nil, err
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

func (x *GitSync) Process(e transistor.Event) error {
	log.InfoWithFields("Process GitSync event", log.Fields{
		"event": e.Name,
	})

	var err error

	gitSyncEvent := e.Payload.(plugins.GitSync)
	gitSyncEvent.Action = plugins.GetAction("status")
	gitSyncEvent.State = plugins.GetState("fetching")
	gitSyncEvent.StateMessage = ""
	x.events <- e.NewEvent(gitSyncEvent, nil)

	commits, err := x.commits(gitSyncEvent.Project, gitSyncEvent.Git)
	if err != nil {
		gitSyncEvent.State = plugins.GetState("failed")
		gitSyncEvent.StateMessage = fmt.Sprintf("%v (Action: %v)", err.Error(), gitSyncEvent.State)
		event := e.NewEvent(gitSyncEvent, err)
		x.events <- event
		return err
	}

	//branches, err := x.branches(gitSyncEvent.Project, gitSyncEvent.Git)
	//if err != nil {
	//	gitSyncEvent.State = plugins.GetState("failed")
	//	gitSyncEvent.StateMessage = fmt.Sprintf("%v (Action: %v)", err.Error(), gitSyncEvent.State)
	//	event := e.NewEvent(gitSyncEvent, err)
	//	x.events <- event
	//	return err
	//}

	//for i := range branches {
	//	branchEvent := e.NewEvent(branches[i], nil)
	//	x.events <- branchEvent
	//}

	for i := range commits {
		c := commits[i]
		c.Repository = gitSyncEvent.Project.Repository
		c.Ref = fmt.Sprintf("refs/heads/%s", gitSyncEvent.Git.Branch)

		if c.Hash == gitSyncEvent.From {
			break
		}

		x.events <- e.NewEvent(c, nil)
	}

	gitSyncEvent.State = plugins.GetState("complete")
	gitSyncEvent.StateMessage = ""
	x.events <- e.NewEvent(gitSyncEvent, nil)

	return nil
}
