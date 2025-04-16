// pkg/repository/dynamodb/profile_repo.go
package dynamodb

import (
	"context"
	"fmt"

	"github.com/FriedGlue/BookIt/api/pkg/models"
	"github.com/FriedGlue/BookIt/api/pkg/usecase"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type profileRepo struct {
	db        *dynamodb.DynamoDB
	tableName string
}

func NewProfileRepo(sess *session.Session, table string) usecase.ProfileRepo {
	return &profileRepo{db: dynamodb.New(sess), tableName: table}
}

func (r *profileRepo) LoadProfile(ctx context.Context, userID string) (models.Profile, error) {
	out, err := r.db.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key:       map[string]*dynamodb.AttributeValue{"_id": {S: aws.String(userID)}},
	})
	if err != nil {
		return models.Profile{}, err
	}
	var p models.Profile
	if err := dynamodbattribute.UnmarshalMap(out.Item, &p); err != nil {
		return p, err
	}
	return p, nil
}

func (r *profileRepo) SaveProfile(ctx context.Context, p models.Profile) error {
	// Ensure the ID field is set
	if p.ID == "" {
		return fmt.Errorf("profile ID cannot be empty")
	}

	av, err := dynamodbattribute.MarshalMap(p)
	if err != nil {
		return err
	}
	_, err = r.db.PutItemWithContext(ctx, &dynamodb.PutItemInput{TableName: aws.String(r.tableName), Item: av})
	return err
}
