// pkg/repository/dynamodb/book_repo.go
package dynamodb

import (
	"context"
	"errors"
	"log"

	"github.com/FriedGlue/BookIt/api/pkg/models"
	"github.com/FriedGlue/BookIt/api/pkg/usecase"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

type bookRepo struct {
	db        *dynamodb.DynamoDB
	tableName string
}

func NewBookRepo(sess *session.Session, table string) usecase.BookRepo {
	return &bookRepo{db: dynamodb.New(sess), tableName: table}
}

func (r *bookRepo) Load(ctx context.Context, id string) (models.Book, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"BookID": {S: aws.String(id)},
		},
	}
	res, err := r.db.GetItemWithContext(ctx, input)
	if err != nil {
		return models.Book{}, err
	}
	if res.Item == nil {
		return models.Book{}, errors.New("not found")
	}
	var b models.Book
	if err := dynamodbattribute.UnmarshalMap(res.Item, &b); err != nil {
		return b, err
	}
	return b, nil
}

func (r *bookRepo) QueryAll(ctx context.Context) ([]models.Book, error) {
	out, err := r.db.ScanWithContext(ctx, &dynamodb.ScanInput{TableName: aws.String(r.tableName)})
	if err != nil {
		return nil, err
	}
	var books []models.Book
	if err := dynamodbattribute.UnmarshalListOfMaps(out.Items, &books); err != nil {
		return nil, err
	}
	return books, nil
}

func (r *bookRepo) Save(ctx context.Context, b models.Book) error {
	av, err := dynamodbattribute.MarshalMap(b)
	if err != nil {
		return err
	}
	_, err = r.db.PutItemWithContext(ctx, &dynamodb.PutItemInput{TableName: aws.String(r.tableName), Item: av})
	return err
}

func (r *bookRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.DeleteItemWithContext(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"BookID": {S: aws.String(id)},
		},
	})
	return err
}

func (r *bookRepo) SearchByISBN(ctx context.Context, isbn string) ([]models.Book, error) {
	expr, _ := expression.NewBuilder().WithFilter(expression.Equal(expression.Name("ISBN"), expression.Value(isbn))).Build()
	out, err := r.db.ScanWithContext(ctx, &dynamodb.ScanInput{
		TableName:                 aws.String(r.tableName),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return nil, err
	}
	var books []models.Book
	if err := dynamodbattribute.UnmarshalListOfMaps(out.Items, &books); err != nil {
		return nil, err
	}
	return books, nil
}

func (r *bookRepo) SearchByTitle(ctx context.Context, q string) ([]models.Book, error) {
	log.Printf("DynamoDB SearchByTitle called with query: '%s'", q)

	// Create a more flexible search expression that checks both title fields
	// Use OR condition to match either regular title or lowercase title
	// Use begins_with for more flexible matching
	expr, err := expression.NewBuilder().WithFilter(
		expression.Or(
			expression.Contains(expression.Name("title"), q),
			expression.Contains(expression.Name("titleLowercase"), q),
			expression.BeginsWith(expression.Name("title"), q),
			expression.BeginsWith(expression.Name("titleLowercase"), q),
		),
	).Build()

	if err != nil {
		log.Printf("Error building expression: %v", err)
		return nil, err
	}

	input := &dynamodb.ScanInput{
		TableName:                 aws.String(r.tableName),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}
	log.Printf("DynamoDB scan input: %v", input)

	out, err := r.db.ScanWithContext(ctx, input)
	if err != nil {
		log.Printf("DynamoDB scan error: %v", err)
		return nil, err
	}

	log.Printf("DynamoDB scan returned %d items", len(out.Items))
	if len(out.Items) > 0 {
		log.Printf("First item raw: %v", out.Items[0])
	}

	var books []models.Book
	if err := dynamodbattribute.UnmarshalListOfMaps(out.Items, &books); err != nil {
		log.Printf("Error unmarshaling items: %v", err)
		return nil, err
	}

	log.Printf("Returning %d books from title search", len(books))
	return books, nil
}
