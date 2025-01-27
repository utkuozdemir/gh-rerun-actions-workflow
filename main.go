package main

import (
	"context"
	"errors"
	"flag"
	"github.com/google/go-github/v68/github"
	"github.com/gregjones/httpcache"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	slog.SetDefault(logger)

	client := github.NewClient(
		httpcache.NewMemoryCacheTransport().Client(),
	).WithAuthToken(os.Getenv("GITHUB_TOKEN"))

	var (
		repoFlag       string
		runIDFlag      int64
		maxRerunsFlag  int
		interval       time.Duration
		apiCallTimeout time.Duration
	)

	flag.StringVar(&repoFlag, "repo", "", "owner/repository to work on")
	flag.Int64Var(&runIDFlag, "run-id", 0, "workflow run ID to check")
	flag.IntVar(&maxRerunsFlag, "max-reruns", 5, "maximum number of reruns")
	flag.DurationVar(&interval, "interval", 1*time.Minute, "interval between checks")
	flag.DurationVar(&apiCallTimeout, "api-call-timeout", 5*time.Second, "timeout for API calls")

	flag.Parse()

	repoParts := strings.Split(repoFlag, "/")
	if len(repoParts) != 2 {
		return errors.New("invalid repository format")
	}

	if runIDFlag == 0 {
		return errors.New("invalid run ID")
	}

	owner := repoParts[0]
	repo := repoParts[1]

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()

	reruns := 0

	for {
		success, failure, err := checkWorkflowRun(ctx, apiCallTimeout, client, logger, owner, repo, runIDFlag)
		if err != nil {
			return err
		}

		if success {
			logger.Info("workflow run succeeded")

			return nil
		}

		if failure {
			if maxRerunsFlag > 0 && reruns >= maxRerunsFlag {
				logger.Info("workflow run failed and reached maximum number of reruns")

				return nil
			}

			logger.Info("workflow run failed, rerunning")

			if err = rerun(ctx, apiCallTimeout, client, owner, repo, runIDFlag); err != nil {
				return err
			}

			reruns++
		}

		logger.Info("wait for the next check")

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(interval):
		}
	}
}

func rerun(ctx context.Context, timeout time.Duration, client *github.Client, owner, repo string, runID int64) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := client.Actions.RerunFailedJobsByID(ctx, owner, repo, runID)

	return err
}

func checkWorkflowRun(ctx context.Context, timeout time.Duration, client *github.Client, logger *slog.Logger, owner, repo string, runID int64) (success, failure bool, err error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	workflowRun, _, err := client.Actions.GetWorkflowRunByID(ctx, owner, repo, runID)
	if err != nil {
		return false, false, err
	}

	logger.Info("workflow run", "status", workflowRun.GetStatus(), "conclusion", workflowRun.GetConclusion())

	switch workflowRun.GetStatus() {
	case "completed":
		switch workflowRun.GetConclusion() {
		case "success":
			return true, false, nil
		case "failure", "cancelled":
			return false, true, nil
		default:
			logger.Info("workflow conclusion", "conclusion", workflowRun.GetConclusion())
		}
	}

	return false, false, nil
}
