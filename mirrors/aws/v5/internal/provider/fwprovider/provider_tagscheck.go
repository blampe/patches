package fwprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/blampe/patches/mirrors/aws/v5/names"
)

type disableTagsSchemaCheckKey struct{}

func DisableTagSchemaCheck(ctx context.Context) context.Context {
	return context.WithValue(ctx, disableTagsSchemaCheckKey{}, true)
}

func schemaResponseForTagsChecking(
	ctx context.Context,
	r resource.ResourceWithConfigure,
) *resource.SchemaResponse {
	flag := ctx.Value(disableTagsSchemaCheckKey{})
	switch flag := flag.(type) {
	case bool:
		if flag {
			return &resource.SchemaResponse{
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						names.AttrTags: schema.MapAttribute{
							Computed: true,
						},
						names.AttrTagsAll: schema.MapAttribute{
							Computed: false,
						},
					},
				},
			}
		}
	}
	var resp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &resp)
	return &resp
}
