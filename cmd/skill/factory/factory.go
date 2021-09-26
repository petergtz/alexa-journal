package factory

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"

	skill "github.com/petergtz/alexa-journal"
	"github.com/petergtz/alexa-journal/dynamodb"
	"github.com/petergtz/alexa-journal/github"
	"github.com/petergtz/alexa-journal/locale/resources"

	"github.com/petergtz/alexa-journal/drive"

	"go.uber.org/zap"
)

func CreateSkill(logger *zap.SugaredLogger) *skill.JournalSkill {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		logger.Fatal("GITHUB_TOKEN not set. Please set it to a valid token from Github.")
	}
	githubErrorReporter := github.NewGithubErrorReporter(
		"petergtz",
		"alexa-journal",
		githubToken,
		logger,
		"``fields @timestamp, @message | filter `error-id` = %v``",
		sns.New(session.Must(session.NewSession(&aws.Config{Region: aws.String("eu-west-1")}))),
		"arn:aws:sns:eu-west-1:512841817041:AlexaJournalErrors")

	return skill.NewJournalSkill(
		drive.NewDriveSheetJournalProvider(logger),
		&drive.DriveSheetErrorInterpreter{ErrorReporter: githubErrorReporter},
		logger,
		githubErrorReporter,
		CreateI18nBundle(),
		dynamodb.CreateConfigService("AlexaJournalConfig", "eu-central-1", githubErrorReporter),
	)
}

type EmptyConfigService struct{}

func (*EmptyConfigService) GetConfig(userID string) skill.Config             { return skill.Config{} }
func (*EmptyConfigService) PersistConfig(userID string, config skill.Config) {}

func CreateI18nBundle() *i18n.Bundle {
	i18nBundle := i18n.NewBundle(language.English)
	i18nBundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	i18nBundle.MustParseMessageFileBytes(resources.DeDe, "active.de.toml")
	i18nBundle.MustParseMessageFileBytes(resources.EnUs, "active.en.toml")
	return i18nBundle
}
