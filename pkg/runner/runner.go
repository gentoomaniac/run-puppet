package runner

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"

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
	ctx           context.Context
	tracer        trace.Tracer
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

func WithContext(ctx context.Context) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.ctx = ctx
	}
}
func WithTracer(tracer trace.Tracer) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.tracer = tracer
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
	log.Debug().Msg("runner")

	ctx, span := r.options.tracer.Start(r.options.ctx, "runner.Run()")
	defer span.End()

	//TODO: add random delay

	if r.options.clone {
		err := cloneRepo(ctx, r.options.tracer, r.options.localRepoPath, r.options.remoteRepoUrl)
		if err != nil {
			return err
		}
	}

	return nil
}

func cloneRepo(ctx context.Context, tracer trace.Tracer, localPath string, remoteUrl string) error {
	_, span := tracer.Start(ctx, "cloneRepo()")
	defer span.End()

	if err := os.RemoveAll(localPath); err != nil {
		return err
	}

	_, err := git.PlainClone(localPath, false, &git.CloneOptions{
		URL:      remoteUrl,
		Progress: os.Stdout,
	})

	if err != nil {
		return err
	}
	return nil
}
