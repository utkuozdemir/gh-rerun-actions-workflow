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
		repoFlag      string
		runIDFlag     int64
		maxRerunsFlag int
		interval      time.Duration
	)

	flag.StringVar(&repoFlag, "repo", "", "owner/repository to work on")
	flag.Int64Var(&runIDFlag, "run-id", 0, "workflow run ID to check")
	flag.IntVar(&maxRerunsFlag, "max-reruns", 5, "maximum number of reruns")
	flag.DurationVar(&interval, "interval", 1*time.Minute, "interval between checks")

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

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	reruns := 0

	for {
		success, failure, err := checkWorkflowRun(ctx, client, logger, owner, repo, runIDFlag)
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

			if _, err = client.Actions.RerunFailedJobsByID(ctx, owner, repo, runIDFlag); err != nil {
				return err
			}

			reruns++
		}

		logger.Info("wait for the next check")

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func checkWorkflowRun(ctx context.Context, client *github.Client, logger *slog.Logger, owner, repo string, runID int64) (success, failure bool, err error) {
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
		case "failure":
			return false, true, nil
		}
	}

	return false, false, nil
}
