// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/blampe/patches/mirrors/aws/v5/internal/conns"
	"github.com/blampe/patches/mirrors/aws/v5/internal/errs/sdkdiag"
	"github.com/blampe/patches/mirrors/aws/v5/internal/tfresource"
	"github.com/blampe/patches/mirrors/aws/v5/names"
)

// @SDKDataSource("aws_dx_gateway", name="Gateway")
func dataSourceGateway() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceGatewayRead,

		Schema: map[string]*schema.Schema{
			"amazon_side_asn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrOwnerAccountID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceGatewayRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &directconnect.DescribeDirectConnectGatewaysInput{}

	gateway, err := findGateway(ctx, conn, input, func(v *awstypes.DirectConnectGateway) bool {
		return aws.ToString(v.DirectConnectGatewayName) == name
	})

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("Direct Connect Gateway", err))
	}

	d.SetId(aws.ToString(gateway.DirectConnectGatewayId))
	d.Set("amazon_side_asn", strconv.FormatInt(aws.ToInt64(gateway.AmazonSideAsn), 10))
	d.Set(names.AttrARN, gatewayARN(ctx, meta.(*conns.AWSClient), d.Id()))
	d.Set(names.AttrOwnerAccountID, gateway.OwnerAccount)

	return diags
}
