package shim

import (
	"context"

	pfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/blampe/patches/mirrors/aws/v6/internal/provider/framework"
	"github.com/blampe/patches/mirrors/aws/v6/internal/provider/sdkv2"
	"github.com/blampe/patches/mirrors/aws/v6/internal/tags"
)

type UpstreamProvider struct {
	SDKV2Provider           *schema.Provider
	PluginFrameworkProvider pfprovider.Provider
}

func NewUpstreamProvider(ctx context.Context) (UpstreamProvider, error) {
	primary, err := sdkv2.NewProvider(ctx)
	if err != nil {
		return UpstreamProvider{}, err
	}
	pf, err := framework.NewProvider(ctx, primary)
	if err != nil {
		//lintignore:R009
		panic(err)
	}
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
