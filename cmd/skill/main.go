package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/petergtz/go-alexa/lambda"

	skill "github.com/petergtz/alexa-journal"
	"github.com/petergtz/alexa-journal/github"

	"github.com/petergtz/alexa-journal/drive"

	"go.uber.org/zap"
)

func main() {
	logger := createLoggerWith(zap.NewAtomicLevelAt(zap.DebugLevel))
	defer logger.Sync()

	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		logger.Fatal("GITHUB_TOKEN not set. Please set it to a valid token from Github.")
	}

	lambda.StartLambdaSkill(skill.NewJournalSkill(
		&drive.DriveSheetJournalProvider{Log: logger},
		&drive.DriveSheetErrorInterpreter{},
		logger,
		github.NewGithubErrorReporter(
			"petergtz",
			"alexa-journal",
			githubToken,
			logger,
			"`fields @timestamp, @message | filter `error-id` = %v`"),
	), logger)
}

func createLoggerWith(logLevel zap.AtomicLevel) *zap.SugaredLogger {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = logLevel
	loggerConfig.DisableStacktrace = true
	logger, e := loggerConfig.Build()
	if e != nil {
		log.Panic(e)
	}
	rand.Seed(time.Now().UnixNano())
	return logger.Sugar().With("function-instance-id", rand.Int63())
}
