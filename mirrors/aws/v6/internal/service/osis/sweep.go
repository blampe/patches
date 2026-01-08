// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package osis

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/osis"
	"github.com/blampe/patches/mirrors/aws/v6/internal/conns"
	"github.com/blampe/patches/mirrors/aws/v6/internal/sweep"
	"github.com/blampe/patches/mirrors/aws/v6/internal/sweep/awsv2"
	"github.com/blampe/patches/mirrors/aws/v6/internal/sweep/framework"
	"github.com/blampe/patches/mirrors/aws/v6/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_osis_pipeline", sweepPipelines)
}

func sweepPipelines(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.OpenSearchIngestionClient(ctx)
	var input osis.ListPipelinesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := osis.NewListPipelinesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Pipelines {
			name := aws.ToString(v.PipelineName)

			sweepResources = append(sweepResources, framework.NewSweepResource(newPipelineResource, client,
				framework.NewAttribute(names.AttrID, name), framework.NewAttribute("pipeline_name", name)))
		}
	}

	return sweepResources, nil
}
