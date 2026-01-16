// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/blampe/patches/mirrors/aws/v6/internal/conns"
	tftags "github.com/blampe/patches/mirrors/aws/v6/internal/tags"
	"github.com/blampe/patches/mirrors/aws/v6/internal/tfresource"
	"github.com/blampe/patches/mirrors/aws/v6/internal/types/option"
)

func findSecretTag(ctx context.Context, conn *secretsmanager.Client, identifier, key string) (*string, error) {
	listTags, err := listSecretTags(ctx, conn, identifier)

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if !listTags.KeyExists(key) {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(nil))
	}

	return listTags.KeyValue(key), nil
}

func listSecretTags(ctx context.Context, conn *secretsmanager.Client, identifier string) (tftags.KeyValueTags, error) {
	output, err := findSecretByID(ctx, conn, identifier)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return keyValueTags(ctx, output.Tags), nil
}

// ListTags lists secretsmanager service tags and set them in Context.
// It is called from outside this package.
func (p *servicePackage) ListTags(ctx context.Context, meta any, identifier string) error {
	tags, err := listSecretTags(ctx, meta.(*conns.AWSClient).SecretsManagerClient(ctx), identifier)

	if err != nil {
		return err
	}

	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = option.Some(tags)
	}

	return nil
}
