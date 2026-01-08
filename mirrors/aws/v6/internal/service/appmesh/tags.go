// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appmesh

import (
	"context"

	tftags "github.com/blampe/patches/mirrors/aws/v6/internal/tags"
	"github.com/blampe/patches/mirrors/aws/v6/internal/types/option"
)

// setTagsOut sets KeyValueTags in Context.
func setKeyValueTagsOut(ctx context.Context, tags tftags.KeyValueTags) {
	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = option.Some(tags)
	}
}
