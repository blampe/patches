// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh

import (
	"context"

	tftags "github.com/blampe/patches/mirrors/aws/v5/internal/tags"
	"github.com/blampe/patches/mirrors/aws/v5/internal/types/option"
)

// setTagsOut sets KeyValueTags in Context.
func setKeyValueTagsOut(ctx context.Context, tags tftags.KeyValueTags) {
	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = option.Some(tags)
	}
}
