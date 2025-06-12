// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resiliencehub

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehub"
	"github.com/blampe/patches/mirrors/aws/v5/internal/conns"
	"github.com/blampe/patches/mirrors/aws/v5/internal/sweep"
	"github.com/blampe/patches/mirrors/aws/v5/internal/sweep/awsv2"
	"github.com/blampe/patches/mirrors/aws/v5/internal/sweep/framework"
	"github.com/blampe/patches/mirrors/aws/v5/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_resiliencehub_resiliency_policy", sweepResiliencyPolicy)
}

func sweepResiliencyPolicy(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResilienceHubClient(ctx)

	var sweepResources []sweep.Sweepable

	pages := resiliencehub.NewListResiliencyPoliciesPaginator(conn, &resiliencehub.ListResiliencyPoliciesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, policies := range page.ResiliencyPolicies {
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceResiliencyPolicy, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(policies.PolicyArn)),
			))
		}
	}

	return sweepResources, nil
}
