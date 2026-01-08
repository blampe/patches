// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/blampe/patches/mirrors/aws/v6/internal/conns"
	"github.com/blampe/patches/mirrors/aws/v6/internal/errs/sdkdiag"
	"github.com/blampe/patches/mirrors/aws/v6/internal/verify"
	"github.com/blampe/patches/mirrors/aws/v6/names"
)

// @SDKDataSource("aws_networkfirewall_resource_policy", name="Resource Policy")
func dataSourceResourcePolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceResourcePolicyRead,

		Schema: map[string]*schema.Schema{
			names.AttrPolicy: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResourceARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func dataSourceResourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	resourceARN := d.Get(names.AttrResourceARN).(string)
	policy, err := findResourcePolicyByARN(ctx, conn, resourceARN)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading NetworkFirewall Resource Policy (%s): %s", resourceARN, err)
	}

	d.SetId(resourceARN)
	d.Set(names.AttrPolicy, policy)
	d.Set(names.AttrResourceARN, resourceARN)

	return diags
}
