// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bcmdataexports

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bcmdataexports"
	"github.com/blampe/patches/mirrors/aws/v6/internal/conns"
	"github.com/blampe/patches/mirrors/aws/v6/internal/sweep"
	"github.com/blampe/patches/mirrors/aws/v6/internal/sweep/awsv2"
	"github.com/blampe/patches/mirrors/aws/v6/internal/sweep/framework"
	"github.com/blampe/patches/mirrors/aws/v6/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_bcmdataexports_export", sweepExports)
}

func sweepExports(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.BCMDataExportsClient(ctx)

	var sweepResources []sweep.Sweepable

	in := bcmdataexports.ListExportsInput{}
	pages := bcmdataexports.NewListExportsPaginator(conn, &in)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, b := range page.Exports {
			arn := aws.ToString(b.ExportArn)

			log.Printf("[INFO] Deleting BCM Data Exports Export: %s", arn)
			sweepResources = append(sweepResources, framework.NewSweepResource(newExportResource, client,
				framework.NewAttribute(names.AttrARN, arn),
			))
		}
	}

	return sweepResources, nil
}
