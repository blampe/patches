package shim

import (
	"context"
	"fmt"

	pfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/blampe/patches/mirrors/aws/v5/internal/provider"
	"github.com/blampe/patches/mirrors/aws/v5/internal/provider/fwprovider"
	"github.com/blampe/patches/mirrors/aws/v5/internal/tags"
)

type UpstreamProvider struct {
	SDKV2Provider           *schema.Provider
	PluginFrameworkProvider pfprovider.Provider
}

func NewUpstreamProvider(ctx context.Context) (UpstreamProvider, error) {
	ctx = fwprovider.DisableTagSchemaCheck(ctx)
	primary, err := provider.New(provider.DisableTagSchemaCheck(ctx))
	if err != nil {
		return UpstreamProvider{}, err
	}
	if primary != nil {
		markTagsAllNotComputedForResources(primary)
	}
	pf := fwprovider.New(primary)
	return UpstreamProvider{
		SDKV2Provider:           primary,
		PluginFrameworkProvider: pf,
	}, nil
}

type TagConfig = tags.DefaultConfig

type TagIgnoreConfig = tags.IgnoreConfig

func NewTagConfig(ctx context.Context, i interface{}) TagConfig {
	return TagConfig{Tags: tags.New(ctx, i)}
}

// For resources with tags_all attribute, ensures that the schema of tags_all matches the schema of
// tags. In particular, this makes sure tags_all is not computed and is ForceNew if necessary. The
// rationale for this is that Pulumi copies tags to tags_all before it hits the TF layer, so these
// attributes must match in schema.
func markTagsAllNotComputedForResources(sdkV2Provider *schema.Provider) {
	updatedResourcesMap := map[string]*schema.Resource{}
	for rn, r := range sdkV2Provider.ResourcesMap {
		updatedResourcesMap[rn] = markTagsAllNotComputedForResource(rn, r)
	}
	sdkV2Provider.ResourcesMap = updatedResourcesMap
}

func markTagsAllNotComputedForResource(rn string, r *schema.Resource) *schema.Resource {
	u := *r
	if r.SchemaFunc != nil {
		old := r.SchemaFunc
		u.SchemaFunc = func() map[string]*schema.Schema {
			return markTagsAllNotComputedForSchema(rn, old())
		}
	} else {
		u.Schema = markTagsAllNotComputedForSchema(rn, r.Schema)
	}
	return &u
}

func markTagsAllNotComputedForSchema(rn string, s map[string]*schema.Schema) map[string]*schema.Schema {
	if _, ok := s["tags_all"]; !ok {
		return s
	}
	updatedSchema := map[string]*schema.Schema{}
	for k, v := range s {
		if k == "tags_all" {
			if tagsSchema, ok := s["tags"]; ok {
				tagsAll := *tagsSchema
				updatedSchema[k] = &tagsAll
			} else {
				panic(fmt.Sprintf("Unable to edit tagsAll schema for %q", rn))
			}
		} else {
			updatedSchema[k] = v
		}
	}
	return updatedSchema
}
