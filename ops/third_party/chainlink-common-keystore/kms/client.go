package kms

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
)

// Client is an interface that defines the methods for interacting with AWS KMS. We only expose
// the methods that are needed for our use case, which is to get a public key and sign data.
//
// These methods are based on the AWS SDK for Go v2 KMS client interface.
// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/kms
type Client interface {
	GetPublicKey(ctx context.Context, input *kms.GetPublicKeyInput, opts ...func(*kms.Options)) (*kms.GetPublicKeyOutput, error)
	Sign(ctx context.Context, input *kms.SignInput, opts ...func(*kms.Options)) (*kms.SignOutput, error)
	DescribeKey(ctx context.Context, input *kms.DescribeKeyInput, opts ...func(*kms.Options)) (*kms.DescribeKeyOutput, error)
	ListKeys(ctx context.Context, input *kms.ListKeysInput, opts ...func(*kms.Options)) (*kms.ListKeysOutput, error)
}

// ClientOptions contains options for creating a KMS client.
type ClientOptions struct {
	// Profile is the AWS profile name to use (for local development).
	// If empty, uses default credential chain (IRSA, EC2 instance profiles, etc.).
	Profile string
	// Region is the AWS region. If empty, will be read from profile or environment.
	Region string
}

// NewClient constructs a new KMS client using AWS SDK v2.
// If Profile is specified, it uses profile-based authentication (for local dev).
// Otherwise, it uses the default credential chain (IRSA in production, EC2 instance profiles, etc.).
func NewClient(ctx context.Context, opts ClientOptions) (Client, error) {
	var cfg aws.Config
	var err error

	if opts.Profile != "" {
		// Use profile-based config
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithSharedConfigProfile(opts.Profile),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config with profile %s: %w", opts.Profile, err)
		}
	} else {
		// Use default credential chain
		cfg, err = config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}
	}

	// Set region if provided
	if opts.Region != "" {
		cfg.Region = opts.Region
	}

	return kms.NewFromConfig(cfg), nil
}
