package r2

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// PresignResult holds a short-lived PUT URL and the object key for clients.
type PresignResult struct {
	UploadURL string
	Key       string
	PublicURL string
	ExpiresAt time.Time
}

// Client generates presigned upload URLs for Cloudflare R2 (S3-compatible API).
type Client struct {
	bucket        string
	publicBase    string
	presignTTL    time.Duration
	presignClient *s3.PresignClient
}

// NewClient builds an R2 client from configuration.
func NewClient(cfg *config.R2Config) (*Client, error) {
	if cfg == nil || !cfg.Enabled {
		return nil, fmt.Errorf("r2 is not configured")
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		),
		awsconfig.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("r2 aws config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(strings.TrimRight(cfg.Endpoint, "/"))
		o.UsePathStyle = true
	})

	return &Client{
		bucket:        cfg.Bucket,
		publicBase:    strings.TrimRight(cfg.PublicBaseURL, "/"),
		presignTTL:    cfg.PresignTTL,
		presignClient: s3.NewPresignClient(s3Client),
	}, nil
}

// PresignPutObject returns a presigned URL for uploading one object.
func (c *Client) PresignPutObject(ctx context.Context, key, contentType string) (*PresignResult, error) {
	if key == "" {
		return nil, fmt.Errorf("object key is required")
	}

	out, err := c.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(c.presignTTL))
	if err != nil {
		return nil, fmt.Errorf("presign put object: %w", err)
	}

	expiresAt := time.Now().Add(c.presignTTL)
	publicURL := c.publicURL(key)

	return &PresignResult{
		UploadURL: out.URL,
		Key:       key,
		PublicURL: publicURL,
		ExpiresAt: expiresAt,
	}, nil
}

func (c *Client) publicURL(key string) string {
	if c.publicBase == "" {
		return key
	}
	return c.publicBase + "/" + strings.TrimPrefix(key, "/")
}
