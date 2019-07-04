package github

import (
	"context"
	"fmt"
	"math/rand"

	"go.uber.org/zap"

	"github.com/google/go-github/github"
	gh "github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type GithubErrorReporter struct {
	ghClient *gh.Client
	logger   *zap.SugaredLogger
	ctx      context.Context
	owner    string
	repo     string
}

func NewGithubErrorReporter(owner, repo, token string, logger *zap.SugaredLogger) *GithubErrorReporter {
	ctx := context.TODO()
	return &GithubErrorReporter{
		ghClient: gh.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))),
		ctx:      ctx,
		logger:   logger,
		repo:     repo,
		owner:    owner,
	}
}

func (r *GithubErrorReporter) ReportError(message string, e error) {
	errorID := rand.Int63()

	issue, _, ghErr := r.ghClient.Issues.Create(r.ctx, r.owner, r.repo, &github.IssueRequest{
		Title:    github.String(message),
		Body:     github.String(fmt.Sprintf("An error occurred and it can be found by grepping the logs for `%v`", errorID)),
		Assignee: github.String("petergtz"),
	})
	if ghErr != nil {
		r.logger.Errorw("Error while trying to report error with message: "+message, "original-error", e, "github-error", ghErr)
		return
	}
	r.logger.Errorw(message, "github-issue-id", issue.Number, "github-issue-url", issue.GetHTMLURL(), "error-id", errorID, "error", fmt.Sprintf("%+v", e))
}
