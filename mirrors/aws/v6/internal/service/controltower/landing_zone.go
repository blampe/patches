// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package controltower

import (
	"context"
	stdjson "encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/controltower"
	"github.com/aws/aws-sdk-go-v2/service/controltower/document"
	"github.com/aws/aws-sdk-go-v2/service/controltower/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/blampe/patches/mirrors/aws/v6/internal/conns"
	"github.com/blampe/patches/mirrors/aws/v6/internal/enum"
	"github.com/blampe/patches/mirrors/aws/v6/internal/errs"
	"github.com/blampe/patches/mirrors/aws/v6/internal/errs/sdkdiag"
	"github.com/blampe/patches/mirrors/aws/v6/internal/retry"
	tfsmithy "github.com/blampe/patches/mirrors/aws/v6/internal/smithy"
	tftags "github.com/blampe/patches/mirrors/aws/v6/internal/tags"
	"github.com/blampe/patches/mirrors/aws/v6/internal/tfresource"
	"github.com/blampe/patches/mirrors/aws/v6/names"
)

// @SDKResource("aws_controltower_landing_zone", name="Landing Zone")
// @Tags(identifierAttribute="arn")
func resourceLandingZone() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLandingZoneCreate,
		ReadWithoutTimeout:   resourceLandingZoneRead,
		UpdateWithoutTimeout: resourceLandingZoneUpdate,
		DeleteWithoutTimeout: resourceLandingZoneDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"drift_status": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"latest_available_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"manifest_json": {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      resourceLandingZoneManfiestSuppressEquivalentJSONDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc:             resourceLandingZoneManfiestStateFunc,
			},
			names.AttrVersion: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceLandingZoneCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ControlTowerClient(ctx)

	manifest, err := tfsmithy.DocumentFromJSONString(d.Get("manifest_json").(string), document.NewLazyDocument)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &controltower.CreateLandingZoneInput{
		Manifest: manifest,
		Tags:     getTagsIn(ctx),
		Version:  aws.String(d.Get(names.AttrVersion).(string)),
	}

	output, err := conn.CreateLandingZone(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ControlTower Landing Zone: %s", err)
	}

	id, err := landingZoneIDFromARN(aws.ToString(output.Arn))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(id)

	if _, err := waitLandingZoneOperationSucceeded(ctx, conn, aws.ToString(output.OperationIdentifier), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ControlTower Landing Zone (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceLandingZoneRead(ctx, d, meta)...)
}

func resourceLandingZoneRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ControlTowerClient(ctx)

	landingZone, err := findLandingZoneByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] ControlTower Landing Zone (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ControlTower Landing Zone (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, landingZone.Arn)
	if landingZone.DriftStatus != nil {
		if err := d.Set("drift_status", []any{flattenLandingZoneDriftStatusSummary(landingZone.DriftStatus)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting drift_status: %s", err)
		}
	} else {
		d.Set("drift_status", nil)
	}
	d.Set("latest_available_version", landingZone.LatestAvailableVersion)
	if landingZone.Manifest != nil {
		v, err := tfsmithy.DocumentToJSONString(landingZone.Manifest)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set("manifest_json", v)
	} else {
		d.Set("manifest_json", nil)
	}
	d.Set(names.AttrVersion, landingZone.Version)

	return diags
}

func resourceLandingZoneUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ControlTowerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		manifest, err := tfsmithy.DocumentFromJSONString(d.Get("manifest_json").(string), document.NewLazyDocument)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &controltower.UpdateLandingZoneInput{
			LandingZoneIdentifier: aws.String(d.Id()),
			Manifest:              manifest,
			Version:               aws.String(d.Get(names.AttrVersion).(string)),
		}

		output, err := conn.UpdateLandingZone(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ControlTower Landing Zone (%s): %s", d.Id(), err)
		}

		if _, err := waitLandingZoneOperationSucceeded(ctx, conn, aws.ToString(output.OperationIdentifier), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for ControlTower Landing Zone (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceLandingZoneRead(ctx, d, meta)...)
}

func resourceLandingZoneDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ControlTowerClient(ctx)

	log.Printf("[DEBUG] Deleting ControlTower Landing Zone: %s", d.Id())
	input := controltower.DeleteLandingZoneInput{
		LandingZoneIdentifier: aws.String(d.Id()),
	}
	output, err := conn.DeleteLandingZone(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ControlTower Landing Zone: %s", err)
	}

	if _, err := waitLandingZoneOperationSucceeded(ctx, conn, aws.ToString(output.OperationIdentifier), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ControlTower Landing Zone (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func landingZoneIDFromARN(arnString string) (string, error) {
	arn, err := arn.Parse(arnString)
	if err != nil {
		return "", err
	}

	// arn:${Partition}:controltower:${Region}:${Account}:landingzone/${LandingZoneId}
	return strings.TrimPrefix(arn.Resource, "landingzone/"), nil
}

func findLandingZoneByID(ctx context.Context, conn *controltower.Client, id string) (*types.LandingZoneDetail, error) {
	input := &controltower.GetLandingZoneInput{
		LandingZoneIdentifier: aws.String(id),
	}

	output, err := conn.GetLandingZone(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.LandingZone == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.LandingZone, nil
}

func findLandingZoneOperationByID(ctx context.Context, conn *controltower.Client, id string) (*types.LandingZoneOperationDetail, error) {
	input := &controltower.GetLandingZoneOperationInput{
		OperationIdentifier: aws.String(id),
	}

	output, err := conn.GetLandingZoneOperation(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.OperationDetails == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.OperationDetails, nil
}

func statusLandingZoneOperation(ctx context.Context, conn *controltower.Client, id string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findLandingZoneOperationByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitLandingZoneOperationSucceeded(ctx context.Context, conn *controltower.Client, id string, timeout time.Duration) (*types.LandingZoneOperationDetail, error) { //nolint:unparam
	stateConf := &sdkretry.StateChangeConf{
		Pending: enum.Slice(types.LandingZoneOperationStatusInProgress),
		Target:  enum.Slice(types.LandingZoneOperationStatusSucceeded),
		Refresh: statusLandingZoneOperation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.LandingZoneOperationDetail); ok {
		if status := output.Status; status == types.LandingZoneOperationStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func flattenLandingZoneDriftStatusSummary(apiObject *types.LandingZoneDriftStatusSummary) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrStatus: apiObject.Status,
	}

	return tfMap
}

// Normalize JSON manifest to avoid significant space and fix retentionDays cycling between string and number. This
// normalizes retentionDays to always use numbers.
func resourceLandingZoneNormalizeManifest(jsonManifest string) (string, error) {
	var data any
	if err := stdjson.Unmarshal([]byte(jsonManifest), &data); err != nil {
		return "", err
	}

	normRetentionDays := func(key string, value any) (any, bool) {
		if key != "retentionDays" {
			return nil, false
		}
		valueStr, isStr := value.(string)
		if !isStr {
			return nil, false
		}
		valueInt, err := strconv.Atoi(valueStr)
		if err != nil {
			return nil, false
		}
		return valueInt, true
	}

	var walk func(any)
	walk = func(node any) {
		switch n := node.(type) {
		case []any:
			for _, elem := range n {
				walk(elem)
			}
		case map[string]any:
			for key, elem := range n {
				if newElem, ok := normRetentionDays(key, elem); ok {
					n[key] = newElem
				}
			}
			for _, elem := range n {
				walk(elem)
			}
		default:
			{
			}
		}
	}
	walk(data)

	bytes, err := stdjson.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes[:]), nil
}

func resourceLandingZoneManfiestSuppressEquivalentJSONDiffs(k, old, new string, d *schema.ResourceData) bool {
	oldNorm, err := resourceLandingZoneNormalizeManifest(old)
	if err != nil {
		return false
	}
	newNorm, err := resourceLandingZoneNormalizeManifest(new)
	if err != nil {
		return false
	}
	return oldNorm == newNorm
}

func resourceLandingZoneManfiestStateFunc(jsonString any) string {
	s, ok := jsonString.(string)
	if !ok {
		return ""
	}
	ns, err := resourceLandingZoneNormalizeManifest(s)
	if err != nil {
		return s
	}
	return ns
}
