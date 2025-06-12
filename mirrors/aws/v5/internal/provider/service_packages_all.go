package provider

import (
	"context"

	"github.com/blampe/patches/mirrors/aws/v5/internal/conns"
	"github.com/blampe/patches/mirrors/aws/v5/internal/service/s3legacy"
)

func servicePackagesAll(ctx context.Context) []conns.ServicePackage {
	return append(servicePackages(ctx), s3legacy.ServicePackage(ctx))
}
