package gamelift

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/gamelift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/gamelift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/blampe/patches/mirrors/aws/v5/internal/conns"
	"github.com/blampe/patches/mirrors/aws/v5/internal/enum"
	"github.com/blampe/patches/mirrors/aws/v5/internal/errs"
	"github.com/blampe/patches/mirrors/aws/v5/internal/flex"
	tftags "github.com/blampe/patches/mirrors/aws/v5/internal/tags"
)

// @SDKResource("aws_gamelift_matchmaking_configuration", name="Gamelift Matchmaking Configuration")
func ResourceMatchMakingConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMatchmakingConfigurationCreate,
		ReadWithoutTimeout:   resourceMatchmakingConfigurationRead,
		UpdateWithoutTimeout: resourceMatchmakingConfigurationUpdate,
		DeleteWithoutTimeout: resourceMatchmakingConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"acceptance_required": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"acceptance_timeout_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 600),
			},
			"additional_player_count": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"backfill_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.BackfillMode](),
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_event_data": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"flex_match_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.FlexMatchMode](),
			},
			"game_property": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 16,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 32),
						},
						"value": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 96),
						},
					},
				},
			},
			"game_session_data": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 4096),
			},
			"game_session_queue_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 256),
						validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9:/-]+$`), "must contain only alphanumeric characters, colon, slash and hyphens"),
					),
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 128),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-\.]*$`), "must contain only alphanumeric characters, hyphens and periods"),
				),
			},
			"notification_target": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 300),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9:_/-]*$`), "must contain only alphanumeric characters, colons, underscores, slashes and hyphens"),
				),
			},
			"request_timeout_seconds": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 43200),
			},
			"rule_set_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rule_set_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 128),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-\.]*$`), "must contain only alphanumeric characters, hyphens and periods"),
				),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaTrulyComputed(),
		},
	}
}

func resourceMatchmakingConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig(ctx)
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	input := gamelift.CreateMatchmakingConfigurationInput{
		AcceptanceRequired:    aws.Bool(d.Get("acceptance_required").(bool)),
		Name:                  aws.String(d.Get("name").(string)),
		RequestTimeoutSeconds: aws.Int32(int32(d.Get("request_timeout_seconds").(int))),
		RuleSetName:           aws.String(d.Get("rule_set_name").(string)),
		Tags:                  Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("acceptance_timeout_seconds"); ok {
		input.AcceptanceTimeoutSeconds = aws.Int32(int32(v.(int)))
	}
	if v, ok := d.GetOk("additional_player_count"); ok {
		input.AdditionalPlayerCount = aws.Int32(int32(v.(int)))
	}
	if v, ok := d.GetOk("backfill_mode"); ok {
		input.BackfillMode = awstypes.BackfillMode(v.(string))
	}
	if v, ok := d.GetOk("custom_event_data"); ok {
		input.CustomEventData = aws.String(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("flex_match_mode"); ok {
		input.FlexMatchMode = awstypes.FlexMatchMode(v.(string))
	}
	if v, ok := d.GetOk("game_property"); ok {
		set := v.(*schema.Set)
		input.GameProperties = expandGameliftGameProperties(set.List())
	}
	if v, ok := d.GetOk("game_session_data"); ok {
		input.GameSessionData = aws.String(v.(string))
	}
	if v, ok := d.GetOk("game_session_queue_arns"); ok {
		input.GameSessionQueueArns = flex.ExpandStringValueSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("notification_target"); ok {
		input.NotificationTarget = aws.String(v.(string))
	}

	log.Printf("[INFO] Creating GameLift Matchmaking Configuration: %v", input)
	out, err := conn.CreateMatchmakingConfiguration(ctx, &input)
	if err != nil {
		return diag.Errorf("error creating GameLift Matchmaking Configuration: %s", err)
	}

	d.SetId(aws.ToString(out.Configuration.ConfigurationArn))
	return resourceMatchmakingConfigurationRead(ctx, d, meta)
}

func resourceMatchmakingConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)

	log.Printf("[INFO] Describing GameLift Matchmaking Configuration: %s", d.Id())
	out, err := conn.DescribeMatchmakingConfigurations(ctx, &gamelift.DescribeMatchmakingConfigurationsInput{
		Names: []string{d.Id()},
	})
	if err != nil {
		if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "Configuration not found") ||
			errs.IsA[*awstypes.NotFoundException](err) {
			log.Printf("[WARN] GameLift Matchmaking Configuration (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("error reading GameLift Matchmaking Configuration (%s): %s", d.Id(), err)
	}
	configurations := out.Configurations

	if len(configurations) < 1 {
		log.Printf("[WARN] GameLift Matchmaking Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if len(configurations) != 1 {
		return diag.Errorf("expected exactly 1 GameLift Matchmaking Configuration, found %d under %q",
			len(configurations), d.Id())
	}
	configuration := configurations[0]

	arn := aws.ToString(configuration.ConfigurationArn)
	d.Set("acceptance_required", configuration.AcceptanceRequired)
	d.Set("acceptance_timeout_seconds", configuration.AcceptanceTimeoutSeconds)
	d.Set("additional_player_count", configuration.AdditionalPlayerCount)
	d.Set("arn", arn)
	d.Set("backfill_mode", configuration.BackfillMode)
	d.Set("creation_time", configuration.CreationTime.Format("2006-01-02 15:04:05"))
	d.Set("custom_event_data", configuration.CustomEventData)
	d.Set("description", configuration.Description)
	d.Set("flex_match_mode", configuration.FlexMatchMode)
	d.Set("game_property", flattenGameliftGameProperties(configuration.GameProperties))
	d.Set("game_session_data", configuration.GameSessionData)
	d.Set("game_session_queue_arns", configuration.GameSessionQueueArns)
	d.Set("name", configuration.Name)
	d.Set("notification_target", configuration.NotificationTarget)
	d.Set("request_timeout_seconds", configuration.RequestTimeoutSeconds)
	d.Set("rule_set_arn", configuration.RuleSetArn)
	d.Set("rule_set_name", configuration.RuleSetName)

	tags, err := listTags(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("error listing tags for GameLift Matchmaking Configuration (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %v", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all: %v", err)
	}

	return nil
}

func resourceMatchmakingConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	log.Printf("[INFO] Updating GameLift Matchmaking Configuration: %s", d.Id())

	input := gamelift.UpdateMatchmakingConfigurationInput{
		Name:                  aws.String(d.Id()),
		AcceptanceRequired:    aws.Bool(d.Get("acceptance_required").(bool)),
		RequestTimeoutSeconds: aws.Int32(int32(d.Get("request_timeout_seconds").(int))),
		RuleSetName:           aws.String(d.Get("rule_set_name").(string)),
	}

	if v, ok := d.GetOk("acceptance_timeout_seconds"); ok {
		input.AcceptanceTimeoutSeconds = aws.Int32(int32(v.(int)))
	}
	if v, ok := d.GetOk("additional_player_count"); ok {
		input.AdditionalPlayerCount = aws.Int32(int32(v.(int)))
	}
	if v, ok := d.GetOk("backfill_mode"); ok {
		input.BackfillMode = awstypes.BackfillMode(v.(string))
	}
	if v, ok := d.GetOk("custom_event_data"); ok {
		input.CustomEventData = aws.String(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if d.HasChange("flex_match_mode") {
		if v, ok := d.GetOk("flex_match_mode"); ok {
			input.FlexMatchMode = awstypes.FlexMatchMode(v.(string))
		}
	}
	if v, ok := d.GetOk("game_property"); ok {
		set := v.(*schema.Set)
		input.GameProperties = expandGameliftGameProperties(set.List())
	}
	if v, ok := d.GetOk("game_session_data"); ok {
		input.GameSessionData = aws.String(v.(string))
	}
	if v, ok := d.GetOk("game_session_queue_arns"); ok {
		input.GameSessionQueueArns = flex.ExpandStringValueSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("notification_target"); ok {
		input.NotificationTarget = aws.String(v.(string))
	}

	_, err := conn.UpdateMatchmakingConfiguration(ctx, &input)
	if err != nil {
		return diag.Errorf("error updating Gamelift Matchmaking Configuration (%s): %s", d.Id(), err)
	}

	arn := d.Id()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := updateTags(ctx, conn, arn, o, n); err != nil {
			return diag.Errorf("error updating GameLift Matchmaking Configuration (%s) tags: %s", arn, err)
		}
	}

	return resourceMatchmakingConfigurationRead(ctx, d, meta)
}

func resourceMatchmakingConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)
	log.Printf("[INFO] Deleting GameLift Matchmaking Configuration: %s", d.Id())
	_, err := conn.DeleteMatchmakingConfiguration(ctx, &gamelift.DeleteMatchmakingConfigurationInput{
		Name: aws.String(d.Id()),
	})
	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "Configuration not found") ||
		errs.IsA[*awstypes.NotFoundException](err) {
		return nil
	}
	if err != nil {
		return diag.Errorf("error deleting GameLift Matchmaking Configuration (%s): %s", d.Id(), err)
	}

	return nil
}

func expandGameliftGameProperties(cfg []interface{}) []awstypes.GameProperty {
	properties := make([]awstypes.GameProperty, len(cfg))
	for i, property := range cfg {
		prop := property.(map[string]interface{})
		properties[i] = awstypes.GameProperty{
			Key:   aws.String(prop["key"].(string)),
			Value: aws.String(prop["value"].(string)),
		}
	}
	return properties
}

func flattenGameliftGameProperties(awsProperties []awstypes.GameProperty) []interface{} {
	properties := []interface{}{}
	for _, awsProperty := range awsProperties {
		property := map[string]string{
			"key":   *awsProperty.Key,
			"value": *awsProperty.Value,
		}
		properties = append(properties, property)
	}
	return properties
}

func expandStringList(tfList []interface{}) []*string {
	var result []*string

	for _, rawVal := range tfList {
		if v, ok := rawVal.(string); ok && v != "" {
			result = append(result, &v)
		}
	}

	return result
}
