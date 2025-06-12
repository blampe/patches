package ecr_test

import (
	"fmt"
	"regexp"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/blampe/patches/mirrors/aws/v5/internal/acctest"
)

func TestAccAWSEcrDataSource_ecrCredentials(t *testing.T) {
	ctx := acctest.Context(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEcrCredentialsDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_ecr_credentials.default", "authorization_token"),
					resource.TestCheckResourceAttrSet("data.aws_ecr_credentials.default", "expires_at"),
					resource.TestMatchResourceAttr("data.aws_ecr_credentials.default", "proxy_endpoint", regexp.MustCompile("^https://\\d+\\.dkr\\.ecr\\.[a-zA-Z]+-[a-zA-Z]+-\\d+\\.amazonaws\\.com$")),
				),
			},
		},
	})
}

var testAccCheckAwsEcrCredentialsDataSourceConfig = fmt.Sprintf(`
resource "aws_ecr_repository" "default" {
  name = "foo-repository-terraform-%d"
}

data "aws_ecr_credentials" "default" {
  registry_id = "${aws_ecr_repository.default.registry_id}"
}
`, sdkacctest.RandInt())
