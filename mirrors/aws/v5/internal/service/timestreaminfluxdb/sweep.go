// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreaminfluxdb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb"
	"github.com/blampe/patches/mirrors/aws/v5/internal/conns"
	"github.com/blampe/patches/mirrors/aws/v5/internal/sweep"
	"github.com/blampe/patches/mirrors/aws/v5/internal/sweep/awsv2"
	"github.com/blampe/patches/mirrors/aws/v5/internal/sweep/framework"
	"github.com/blampe/patches/mirrors/aws/v5/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_timestreaminfluxdb_db_instance", sweepDBInstances)
}

func sweepDBInstances(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.TimestreamInfluxDBClient(ctx)
	var input timestreaminfluxdb.ListDbInstancesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := timestreaminfluxdb.NewListDbInstancesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceDBInstance, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.Id)),
			))
		}
	}

	return sweepResources, nil
}
