/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ecr

import (
	"context"
	"fmt"
	"regexp"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

// Client provides operations for interacting with AWS ECR
type Client struct {
	ecrClient *ecr.Client
	region    string
}

// NewClient creates a new ECR client for the specified region
func NewClient(region string) *Client {
	return &Client{
		region: region,
	}
}

// GetLatestTag retrieves the latest tag from the specified ECR repository
func (c *Client) GetLatestTag(ctx context.Context, repositoryName, tagFilter string) (string, error) {
	if c.ecrClient == nil {
		if err := c.initClient(ctx); err != nil {
			return "", fmt.Errorf("failed to initialize ECR client: %w", err)
		}
	}

	// List all image tags
	input := &ecr.DescribeImagesInput{
		RepositoryName: aws.String(repositoryName),
		ImageIds:       []types.ImageIdentifier{},
	}

	result, err := c.ecrClient.DescribeImages(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to describe images in repository %s: %w", repositoryName, err)
	}

	if len(result.ImageDetails) == 0 {
		return "", fmt.Errorf("no images found in repository %s", repositoryName)
	}

	// Extract and filter tags
	var tags []string
	var tagRegex *regexp.Regexp

	if tagFilter != "" {
		tagRegex, err = regexp.Compile(tagFilter)
		if err != nil {
			return "", fmt.Errorf("invalid tag filter regex: %w", err)
		}
	}

	for _, imageDetail := range result.ImageDetails {
		for _, tag := range imageDetail.ImageTags {
			if tag != "" {
				// Apply filter if specified
				if tagRegex != nil && !tagRegex.MatchString(tag) {
					continue
				}
				tags = append(tags, tag)
			}
		}
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found matching filter in repository %s", repositoryName)
	}

	// Sort tags to get the latest (this is a simple sort, you might want semantic versioning)
	sort.Slice(tags, func(i, j int) bool {
		return tags[i] > tags[j] // Descending order
	})

	return tags[0], nil
}

// GetImageDetails retrieves detailed information about images with the specified tag
func (c *Client) GetImageDetails(ctx context.Context, repositoryName, tag string) (*types.ImageDetail, error) {
	if c.ecrClient == nil {
		if err := c.initClient(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize ECR client: %w", err)
		}
	}

	input := &ecr.DescribeImagesInput{
		RepositoryName: aws.String(repositoryName),
		ImageIds: []types.ImageIdentifier{
			{
				ImageTag: aws.String(tag),
			},
		},
	}

	result, err := c.ecrClient.DescribeImages(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe image %s:%s: %w", repositoryName, tag, err)
	}

	if len(result.ImageDetails) == 0 {
		return nil, fmt.Errorf("image not found: %s:%s", repositoryName, tag)
	}

	return &result.ImageDetails[0], nil
}

// ListRepositories lists all ECR repositories in the region
func (c *Client) ListRepositories(ctx context.Context) ([]types.Repository, error) {
	if c.ecrClient == nil {
		if err := c.initClient(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize ECR client: %w", err)
		}
	}

	input := &ecr.DescribeRepositoriesInput{}
	result, err := c.ecrClient.DescribeRepositories(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	return result.Repositories, nil
}

// initClient initializes the ECR client with AWS configuration
func (c *Client) initClient(ctx context.Context) error {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(c.region))
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	c.ecrClient = ecr.NewFromConfig(cfg)
	return nil
}
