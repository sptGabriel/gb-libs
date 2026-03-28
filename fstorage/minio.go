package fstorage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
)

type MinIOCredentials struct {
	AccessKey string
	SecretKey string
	Endpoint  string
	Region    string
}

type MinIOFileStorage struct {
	Client *s3.Client
}

func NewMinIOFileStorage(
	ctx context.Context,
	creds MinIOCredentials,
) (*MinIOFileStorage, error) {
	staticProvider := credentials.NewStaticCredentialsProvider(
		creds.AccessKey,
		creds.SecretKey,
		"",
	)

	region := creds.Region
	if region == "" {
		region = "us-east-1" // Default region for MinIO
	}

	cfg, err := awsConfig.LoadDefaultConfig(
		ctx,
		awsConfig.WithRegion(region),
		awsConfig.WithCredentialsProvider(staticProvider),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(creds.Endpoint)
		o.UsePathStyle = true
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationUnset
	})

	return &MinIOFileStorage{Client: s3Client}, nil
}

func (m *MinIOFileStorage) UploadFile(
	ctx context.Context,
	bucket string,
	filepath string,
	content *bytes.Buffer,
	metadata map[string]string,
) error {
	tags := url.Values{}
	for k, v := range metadata {
		tags.Add(k, v)
	}

	request := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(filepath),
		Body:          content,
		ContentLength: aws.Int64(int64(content.Len())),
		Tagging:       aws.String(tags.Encode()),
	}

	_, err := m.Client.PutObject(
		ctx,
		request,
		s3.WithAPIOptions(
			v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to upload file '%s': %w", filepath, err)
	}

	return nil
}

func (m *MinIOFileStorage) GetDownloadPresignedURL(
	ctx context.Context,
	bucket string,
	filepath string,
	expiresIn time.Duration,
) (string, error) {
	presignClient := s3.NewPresignClient(m.Client)
	obj, err := presignClient.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(filepath),
		},
		s3.WithPresignExpires(expiresIn),
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate download presigned URL: %w", err)
	}

	return obj.URL, nil
}

func (m *MinIOFileStorage) GetUploadPresignedURL(
	ctx context.Context,
	bucket string,
	filepath string,
	expiresIn time.Duration,
) (string, error) {
	presignClient := s3.NewPresignClient(m.Client)
	obj, err := presignClient.PresignPutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(filepath),
		},
		s3.WithPresignExpires(expiresIn),
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate upload presigned URL: %w", err)
	}

	return obj.URL, nil
}

func (m *MinIOFileStorage) BucketExists(
	ctx context.Context,
	bucket string,
) (bool, error) {
	_, err := m.Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})

	exists := true
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NotFound:
				exists = false
				err = nil
			default:
				return false, fmt.Errorf("failed to check if bucket exists: %w", err)
			}
		}
	}

	return exists, err
}

func (m *MinIOFileStorage) CreateBucket(
	ctx context.Context,
	region string,
	bucket string,
) error {
	createBucketInput := &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	}

	// Only set location constraint if region is not us-east-1
	if region != "" && region != "us-east-1" {
		createBucketInput.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		}
	}

	_, err := m.Client.CreateBucket(ctx, createBucketInput)
	if err != nil {
		var owned *types.BucketAlreadyOwnedByYou
		var exists *types.BucketAlreadyExists

		if errors.As(err, &owned) || errors.As(err, &exists) {
			// Bucket already exists, this is not an error
			return nil
		}

		return fmt.Errorf("failed to create bucket '%s': %w", bucket, err)
	}

	return nil
}
