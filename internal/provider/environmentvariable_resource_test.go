package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEnvironmentVariableResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: environmentVariableResourceTemplate("one", "secret", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkly_environment_variable.variable-1", "key", "one"),
					resource.TestCheckResourceAttr("checkly_environment_variable.variable-1", "value", "secret"),
					resource.TestCheckResourceAttr("checkly_environment_variable.variable-1", "locked", "true"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "checkly_environment_variable.variable-1",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				//ImportStateVerifyIgnore: []string{"configurable_attribute", "defaulted"},
			},
			// Update and Read testing
			{
				Config: environmentVariableResourceTemplate("one", "three", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkly_environment_variable.variable-1", "locked", "false"),
					resource.TestCheckResourceAttr("checkly_environment_variable.variable-1", "value", "three"),
				),
			},
			// Update with locked==null and Read testing
			{
				Config: environmentVariableEmptyLocked("one", "four"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("checkly_environment_variable.variable-1", "locked", "false"),
					resource.TestCheckResourceAttr("checkly_environment_variable.variable-1", "value", "four"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func environmentVariableResourceTemplate(key string, value string, locked bool) string {
	return fmt.Sprintf(`
					resource "checkly_environment_variable" "variable-1" {
  						key    = "%s"
  						value  = "%s"
  						locked = %t
					}
					`, key, value, locked)
}

func environmentVariableEmptyLocked(key string, value string) string {
	return fmt.Sprintf(`
	resource "checkly_environment_variable" "variable-1" {
  		key    = "%s"
  		value  = "%s"
	}
`, key, value)
}
