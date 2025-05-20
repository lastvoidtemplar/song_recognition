package internal

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type AWSClient struct {
	ssmClient *ssm.Client
}

func NewAWSClient(logger *slog.Logger) (*AWSClient, error) {
	ctx := context.Background()

	conf, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		logger.With(slog.String("err", err.Error())).Error("Error while loading default config")
		return nil, err
	}

	return &AWSClient{
		ssmClient: ssm.NewFromConfig(conf),
	}, nil
}

func (client *AWSClient) LoadMysqlDBParameters(logger *slog.Logger) (DBMySqlConnectionOption, error) {
	ctx := context.Background()

	paramKeys := []string{
		"/rds/db_host",
		"/rds/db_port",
		"/rds/db_username",
		"/rds/db_password",
		"/rds/db_name",
	}

	params := make(map[string]string)

	decryption := true
	result, err := client.ssmClient.GetParameters(ctx, &ssm.GetParametersInput{
		Names:          paramKeys,
		WithDecryption: &decryption,
	})

	if err != nil {
		logger.With(slog.String("err", err.Error())).Error("Error while extracting db connection params")
		return DBMySqlConnectionOption{}, err
	}

	for _, param := range result.Parameters {
		params[*param.Name] = *param.Value
	}

	return DBMySqlConnectionOption{
		Host:     params["/rds/db_host"],
		Port:     params["/rds/db_port"],
		Username: params["/rds/db_username"],
		Password: params["/rds/db_password"],
		DBName:   params["/rds/db_name"],
	}, err
}
