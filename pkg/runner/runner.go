package runner

import (
	"os"

	"github.com/rs/zerolog/log"

	git "github.com/go-git/go-git/v5"
)

type RunPuppet struct {
	options runPuppetOptions
}

type runPuppetOptions struct {
	binPath       string
	localRepoPath string
	remoteRepoUrl string
	puppetBranch  string
	clone         bool
	now           bool
	noop          bool
}

type RunPuppetOption func(o *runPuppetOptions)

func WithBinPath(path string) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.binPath = path
	}
}

func WithLocalRepoPath(path string) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.localRepoPath = path
	}
}

func WithRemoteRepoUrl(url string) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.remoteRepoUrl = url
	}
}

func WithPuppetBranch(branch string) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.puppetBranch = branch
	}
}

func WithClone(clone bool) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.clone = clone
	}
}

func WithNow(now bool) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.now = now
	}
}
func WithNoop(noop bool) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.noop = noop
	}
}

func New(opts ...RunPuppetOption) RunPuppet {
	opt := &runPuppetOptions{
		puppetBranch:  "main",
		localRepoPath: "/var/lib/puppet-repo",
	}

	for _, o := range opts {
		o(opt)
	}

	return RunPuppet{
		options: *opt,
	}
}

func (r *RunPuppet) Run() error {
	log.Debug().Msg("run puppet")

	if r.options.clone {
		if err := os.RemoveAll(r.options.localRepoPath); err != nil {
			return err
		}

		_, err := git.PlainClone(r.options.localRepoPath, false, &git.CloneOptions{
			URL:      r.options.remoteRepoUrl,
			Progress: os.Stdout,
		})

		if err != nil {
			return err
		}
	}

	return nil
}
