// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry"
	"github.com/blampe/patches/mirrors/aws/v5/internal/conns"
	"github.com/blampe/patches/mirrors/aws/v5/internal/sweep"
	"github.com/blampe/patches/mirrors/aws/v5/internal/sweep/awsv2"
	"github.com/blampe/patches/mirrors/aws/v5/internal/sweep/framework"
	"github.com/blampe/patches/mirrors/aws/v5/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_servicecatalogappregistry_application", sweepScraper)
}

func sweepScraper(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ServiceCatalogAppRegistryClient(ctx)

	var sweepResources []sweep.Sweepable

	pages := servicecatalogappregistry.NewListApplicationsPaginator(conn, &servicecatalogappregistry.ListApplicationsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, application := range page.Applications {
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceApplication, client,
				framework.NewAttribute(names.AttrID, aws.ToString(application.Id)),
			))
		}
	}

	return sweepResources, nil
}
