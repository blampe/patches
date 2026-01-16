// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securitylake

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securitylake"
	"github.com/blampe/patches/mirrors/aws/v6/internal/conns"
	"github.com/blampe/patches/mirrors/aws/v6/internal/sweep"
	"github.com/blampe/patches/mirrors/aws/v6/internal/sweep/awsv2"
	"github.com/blampe/patches/mirrors/aws/v6/internal/sweep/framework"
	"github.com/blampe/patches/mirrors/aws/v6/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_securitylake_data_lake", sweepDataLakes)
}

func sweepDataLakes(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SecurityLakeClient(ctx)
	var input securitylake.ListDataLakesInput
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.ListDataLakes(ctx, &input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.DataLakes {
		sweepResources = append(sweepResources, framework.NewSweepResource(newDataLakeResource, client,
			framework.NewAttribute(names.AttrID, aws.ToString(v.DataLakeArn))))
	}

	return sweepResources, nil
}
