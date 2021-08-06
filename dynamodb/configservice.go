package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	skill "github.com/petergtz/alexa-journal"
	"github.com/petergtz/alexa-journal/util"
	"github.com/pkg/errors"
)

func CreateConfigService(tableName string, region string, errorReporter skill.ErrorReporter) *ConfigService {
	c, e := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	util.PanicOnError(errors.Wrap(e, "Unable to load SDK config"))
	return &ConfigService{
		dynamo:        dynamodb.NewFromConfig(c),
		tableName:     tableName,
		errorReporter: errorReporter,
	}
}

type ConfigService struct {
	dynamo        *dynamodb.Client
	tableName     string
	errorReporter skill.ErrorReporter
}

type record struct {
	UserID string
	skill.Config
}

func (cs *ConfigService) PersistConfig(userID string, config skill.Config) {
	r := &record{UserID: userID, Config: config}
	input, e := attributevalue.MarshalMap(r)
	util.PanicOnError(errors.Wrapf(e, "Could not marshal ConfigService record %#v", r))
	_, e = cs.dynamo.PutItem(context.TODO(), &dynamodb.PutItemInput{
		Item:      input,
		TableName: aws.String(cs.tableName),
	})
	if e != nil {
		cs.errorReporter.ReportError(errors.Wrapf(e, "Could not put item for userID \"%v\" and config \"%#v\"", userID, config))
	}
}

func (cs *ConfigService) GetConfig(userID string) skill.Config {
	key, e := attributevalue.MarshalMap(struct{ UserID string }{UserID: userID})
	util.PanicOnError(errors.Wrapf(e, "Could not marshal UserID %v", userID))

	output, e := cs.dynamo.GetItem(context.TODO(), &dynamodb.GetItemInput{
		Key:       key,
		TableName: aws.String(cs.tableName),
	})
	if e != nil {
		cs.errorReporter.ReportError(errors.Wrapf(e, "Could not get item for key \"%v\"", key))

		// degrade gracefully to defaults
		return skill.Config{}
	}

	if len(output.Item) == 0 {
		return skill.Config{
			BeSuccinct:                     false,
			ShouldExplainAboutSuccinctMode: true,
		}
	}

	var r record
	e = attributevalue.UnmarshalMap(output.Item, &r)
	util.PanicOnError(errors.Wrapf(e, "Could not unmarshal configValue.Item %#v", output.Item))

	return r.Config
}
