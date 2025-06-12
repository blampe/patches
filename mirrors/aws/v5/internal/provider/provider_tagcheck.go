package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/blampe/patches/mirrors/aws/v5/names"
)

type disableTagsSchemaCheckKey struct{}

func DisableTagSchemaCheck(ctx context.Context) context.Context {
	return context.WithValue(ctx, disableTagsSchemaCheckKey{}, true)
}

func schemaMapForTagsChecking(ctx context.Context, r *schema.Resource, tagsComputed bool) map[string]*schema.Schema {
	flag := ctx.Value(disableTagsSchemaCheckKey{})
	switch flag := flag.(type) {
	case bool:
		if flag {
			//lintignore:S013
			return map[string]*schema.Schema{
				names.AttrTags: {
					Type:     schema.TypeMap,
					Computed: tagsComputed,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				names.AttrTagsAll: {
					Type:     schema.TypeMap,
					Computed: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
			}
		}
	}
	return r.SchemaMap()
}
