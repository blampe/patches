// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"context"
	"fmt"
	"maps"

	ecr_sdkv1 "github.com/aws/aws-sdk-go/service/ecr"
	s3_sdkv1 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/blampe/patches/mirrors/aws/v5/internal/errs"
	"github.com/blampe/patches/mirrors/aws/v5/names"
)

func (c *AWSClient) S3Conn(ctx context.Context) *s3_sdkv1.S3 {
	return errs.Must(conn[*s3_sdkv1.S3](ctx, c, names.S3, make(map[string]any)))
}

func (client *AWSClient) S3ConnURICleaningDisabled(ctx context.Context) *s3_sdkv1.S3 {
	config := client.S3Conn(ctx).Config
	t := true
	config.DisableRestProtocolURICleaning = &t

	return s3_sdkv1.New(client.session.Copy(&config))
}

func (c *AWSClient) ECRConn(ctx context.Context) *ecr_sdkv1.ECR {
	return errs.Must(conn[*ecr_sdkv1.ECR](ctx, c, names.ECR, make(map[string]any)))
}

// conn returns the AWS SDK for Go v1 API client for the specified service.
// The default service client (`extra` is empty) is cached. In this case the AWSClient lock is held.
// This function is not a method on `AWSClient` as methods can't be parameterized (https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#no-parameterized-methods).
func conn[T any](ctx context.Context, c *AWSClient, servicePackageName string, extra map[string]any) (T, error) {
	ctx = tflog.SetField(ctx, "tf_aws.service_package", servicePackageName)

	isDefault := len(extra) == 0
	// Default service client is cached.
	if isDefault {
		c.lock.Lock()
		defer c.lock.Unlock() // Runs at function exit, NOT block.

		if raw, ok := c.conns[servicePackageName]; ok {
			if conn, ok := raw.(T); ok {
				return conn, nil
			} else {
				var zero T
				return zero, fmt.Errorf("AWS SDK v1 API client (%s): %T, want %T", servicePackageName, raw, zero)
			}
		}
	}

	sp := c.ServicePackage(ctx, servicePackageName)
	if sp == nil {
		var zero T
		return zero, fmt.Errorf("unknown service package: %s", servicePackageName)
	}

	v, ok := sp.(interface {
		NewConn(context.Context, map[string]any) (T, error)
	})
	if !ok {
		var zero T
		return zero, fmt.Errorf("no AWS SDK v1 API client factory: %s", servicePackageName)
	}

	config := c.apiClientConfig(ctx, servicePackageName)
	maps.Copy(config, extra) // Extras overwrite per-service defaults.
	conn, err := v.NewConn(ctx, config)
	if err != nil {
		var zero T
		return zero, err
	}

	if v, ok := sp.(interface {
		CustomizeConn(context.Context, T) (T, error)
	}); ok {
		conn, err = v.CustomizeConn(ctx, conn)
		if err != nil {
			var zero T
			return zero, err
		}
	}

	// Default service client is cached.
	if isDefault {
		c.conns[servicePackageName] = conn
	}

	return conn, nil
}
