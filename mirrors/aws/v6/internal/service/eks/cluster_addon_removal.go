package eks

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/blampe/patches/mirrors/aws/v6/internal/errs"

	"github.com/blampe/patches/mirrors/aws/v6/internal/flex"
	"github.com/blampe/patches/mirrors/aws/v6/internal/tfresource"
)

const (
	addonCreatedTimeout = 20 * time.Minute
	addonUpdatedTimeout = 20 * time.Minute
	addonDeletedTimeout = 40 * time.Minute
)

func removeAddons(d *schema.ResourceData, conn *eks.Client) error {
	if v, ok := d.GetOk("default_addons_to_remove"); ok && len(v.([]interface{})) > 0 {
		ctx := context.Background()
		var wg sync.WaitGroup
		var removalErrors *multierror.Error

		for _, addon := range flex.ExpandStringList(v.([]interface{})) {
			if addon == nil {
				return fmt.Errorf("addonName cannot be dereferenced")
			}
			addonName := *addon
			wg.Add(1)

			go func() {
				defer wg.Done()
				removalErrors = multierror.Append(removalErrors, removeAddon(d, conn, addonName, ctx))
			}()
		}
		wg.Wait()
		return removalErrors.ErrorOrNil()
	}
	return nil
}

func removeAddon(d *schema.ResourceData, conn *eks.Client, addonName string, ctx context.Context) error {
	log.Printf("[DEBUG] Creating EKS Add-On: %s", addonName)
	createAddonInput := &eks.CreateAddonInput{
		AddonName:          aws.String(addonName),
		ClientRequestToken: aws.String(id.UniqueId()),
		ClusterName:        aws.String(d.Id()),
		ResolveConflicts:   types.ResolveConflictsOverwrite,
	}

	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		_, err := conn.CreateAddon(ctx, createAddonInput)

		if errs.IsAErrorMessageContains[*types.InvalidParameterException](err, "CREATE_FAILED") {
			return retry.RetryableError(err)
		}

		if errs.IsAErrorMessageContains[*types.InvalidParameterException](err, "does not exist") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.CreateAddon(ctx, createAddonInput)
	}

	if err != nil {
		return fmt.Errorf("error creating EKS Add-On (%s): %w", addonName, err)
	}

	_, err = waitAddonCreatedAllowDegraded(ctx, conn, d.Id(), addonName)

	if err != nil {
		return fmt.Errorf("unexpected EKS Add-On (%s) state returned during creation: %w", addonName, err)
	}
	log.Printf("[DEBUG] Created EKS Add-On: %s", addonName)

	deleteAddonInput := &eks.DeleteAddonInput{
		AddonName:   aws.String(addonName),
		ClusterName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting EKS Add-On: %s", addonName)
	_, err = conn.DeleteAddon(ctx, deleteAddonInput)

	if err != nil {
		return fmt.Errorf("error deleting EKS Add-On (%s): %w", addonName, err)
	}

	_, err = waitAddonDeleted(ctx, conn, d.Id(), addonName, addonDeletedTimeout)

	if err != nil {
		return fmt.Errorf("error waiting for EKS Add-On (%s) to delete: %w", addonName, err)
	}
	log.Printf("[DEBUG] Deleted EKS Add-On: %s", addonName)
	return nil
}

func waitAddonCreatedAllowDegraded(ctx context.Context, conn *eks.Client, clusterName, addonName string) (*types.Addon, error) {
	// We do not care about the addons actually being created successfully here. We only want them to be adopted by
	// Terraform to be able to fully remove them afterwards again.

	stateConf := retry.StateChangeConf{
		Pending: []string{string(types.AddonStatusCreating)},
		Target:  []string{string(types.AddonStatusActive), string(types.AddonStatusDegraded)},
		Refresh: statusAddon(ctx, conn, clusterName, addonName),
		Timeout: addonCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Addon); ok {
		if status, health := output.Status, output.Health; status == types.AddonStatusCreateFailed && health != nil {
			tfresource.SetLastError(err, addonIssuesError(health.Issues))
		}

		return output, err
	}

	return nil, err
}
