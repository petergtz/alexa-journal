package github

import (
	"context"
	"fmt"
	"math/rand"
	"runtime/debug"

	"go.uber.org/zap"

	"github.com/google/go-github/github"
	gh "github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type GithubErrorReporter struct {
	ghClient *gh.Client
	logger   *zap.SugaredLogger
	ctx      context.Context
	owner    string
	repo     string
	logsURL  string
}

func NewGithubErrorReporter(owner, repo, token string, logger *zap.SugaredLogger, logsURL string) *GithubErrorReporter {
	ctx := context.TODO()
	return &GithubErrorReporter{
		ghClient: gh.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))),
		ctx:      ctx,
		logger:   logger,
		repo:     repo,
		owner:    owner,
		logsURL:  logsURL,
	}
}

func (r *GithubErrorReporter) ReportPanic(e interface{}) {
	errorID := rand.Int63()

	issue, _, ghErr := r.ghClient.Issues.Create(r.ctx, r.owner, r.repo, &github.IssueRequest{
		Title:    github.String(fmt.Sprintf("Internal Server Error (ErrID: %v)", errorID)),
		Body:     github.String(fmt.Sprintf("An error occurred and it can be found using %v", fmt.Sprintf(r.logsURL, errorID))),
		Assignee: github.String("petergtz"),
	})
	if ghErr != nil {
		r.logger.Errorw("Error while trying to report error", "original-error", e, "github-error", ghErr)
		return
	}
	var errorString string
	if _, hasStackTrace := e.(interface{ StackTrace() errors.StackTrace }); hasStackTrace {
		errorString = fmt.Sprintf("%+v", e)
	} else {
		errorString = fmt.Sprintf("%v\n%s", e, debug.Stack())
	}
	r.logger.Errorw("Internal Server Error",
		"github-issue-id", issue.Number,
		"github-issue-url", issue.GetHTMLURL(),
		"error-id", errorID,
		"error", errorString)

}
