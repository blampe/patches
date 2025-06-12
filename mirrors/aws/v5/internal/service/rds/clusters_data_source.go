// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/blampe/patches/mirrors/aws/v5/internal/conns"
	"github.com/blampe/patches/mirrors/aws/v5/internal/errs/sdkdiag"
	"github.com/blampe/patches/mirrors/aws/v5/internal/namevaluesfilters"
	tfslices "github.com/blampe/patches/mirrors/aws/v5/internal/slices"
	"github.com/blampe/patches/mirrors/aws/v5/names"
)

// @SDKDataSource("aws_rds_clusters", name="Clusters")
func dataSourceClusters() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClustersRead,

		Schema: map[string]*schema.Schema{
			"cluster_arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"cluster_identifiers": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrFilter: namevaluesfilters.Schema(),
		},
	}
}

func dataSourceClustersRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	input := &rds.DescribeDBClustersInput{}

	if v, ok := d.GetOk(names.AttrFilter); ok {
		input.Filters = namevaluesfilters.New(v.(*schema.Set)).RDSFilters()
	}

	clusters, err := findDBClusters(ctx, conn, input, tfslices.PredicateTrue[*types.DBCluster]())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Clusters: %s", err)
	}

	var clusterARNs []string
	var clusterIdentifiers []string

	for _, cluster := range clusters {
		clusterARNs = append(clusterARNs, aws.ToString(cluster.DBClusterArn))
		clusterIdentifiers = append(clusterIdentifiers, aws.ToString(cluster.DBClusterIdentifier))
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))
	d.Set("cluster_arns", clusterARNs)
	d.Set("cluster_identifiers", clusterIdentifiers)

	return diags
}
