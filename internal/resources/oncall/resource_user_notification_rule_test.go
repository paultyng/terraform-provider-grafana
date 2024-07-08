package oncall_test

import (
	"fmt"
	"testing"

	onCallAPI "github.com/grafana/amixr-api-go-client"
	"github.com/grafana/terraform-provider-grafana/v3/internal/common"
	"github.com/grafana/terraform-provider-grafana/v3/internal/testutils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccUserNotificationRule_basic(t *testing.T) {
	testutils.CheckCloudInstanceTestsEnabled(t)

	var (
		// We need an actual user to test the resource
		// This is a user created from my personal email, but it can be replaced by any existing user
		userID       = "joeyorlando"
		resourceName = "grafana_oncall_user_notification_rule.test-acc-user_notification_rule"

		testSteps []resource.TestStep

		ruleTypes = []string{
			"wait",
			"notify_by_slack",
			"notify_by_msteams",
			"notify_by_sms",
			"notify_by_phone_call",
			"notify_by_telegram",
			"notify_by_email",
			"notify_by_mobile_app",
			"notify_by_mobile_app_critical",
		}
	)

	for _, ruleType := range ruleTypes {
		for _, important := range []bool{false, true} {
			var (
				config                 string
				testCheckFuncFunctions = []resource.TestCheckFunc{
					testAccCheckOnCallUserNotificationRuleResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "user_id", userID),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
					resource.TestCheckResourceAttr(resourceName, "type", ruleType),
					resource.TestCheckResourceAttr(resourceName, "important", fmt.Sprintf("%t", important)),
				}
			)

			if ruleType == "wait" {
				config = testAccOnCallUserNotificationRuleWait(userID, important)
				testCheckFuncFunctions = append(testCheckFuncFunctions, resource.TestCheckResourceAttr(resourceName, "duration", "300"))
			} else {
				config = testAccOnCallUserNotificationRuleNotificationStep(ruleType, userID, important)
			}

			testSteps = append(testSteps, resource.TestStep{
				Config: config,
				Check:  resource.ComposeTestCheckFunc(testCheckFuncFunctions...),
			})
		}
	}

	testSteps = append(testSteps, resource.TestStep{
		ResourceName:      resourceName,
		ImportState:       true,
		ImportStateVerify: true,
	})

	resource.ParallelTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testutils.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOnCallUserNotificationRuleResourceDestroy,
		Steps:                    testSteps,
	})
}

func testAccCheckOnCallUserNotificationRuleResourceDestroy(s *terraform.State) error {
	client := testutils.Provider.Meta().(*common.Client).OnCallClient
	for _, r := range s.RootModule().Resources {
		if r.Type != "grafana_oncall_user_notification_rule" {
			continue
		}

		if _, _, err := client.UserNotificationRules.GetUserNotificationRule(r.Primary.ID, &onCallAPI.GetUserNotificationRuleOptions{}); err == nil {
			return fmt.Errorf("UserNotificationRule still exists")
		}
	}
	return nil
}

func testAccOnCallUserNotificationRuleWait(userID string, important bool) string {
	return fmt.Sprintf(`
data "grafana_oncall_user" "user" {
	name = "%s"
}

resource "grafana_oncall_user_notification_rule" "test-acc-user_notification_rule" {
  user_id   = data.grafana_oncall_user.user.id
  type      = "wait"
  position  = 1
  duration  = 300
  important = %t
}
`, userID, important)
}

func testAccOnCallUserNotificationRuleNotificationStep(ruleType, userID string, important bool) string {
	return fmt.Sprintf(`
data "grafana_oncall_user" "user" {
	name = "%s"
}

resource "grafana_oncall_user_notification_rule" "test-acc-user_notification_rule" {
  user_id   = data.grafana_oncall_user.user.id
  type      = "%s"
  position  = 1
  important = %t
}
`, userID, ruleType, important)
}

func testAccCheckOnCallUserNotificationRuleResourceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No UserNotificationRule ID is set")
		}

		client := testutils.Provider.Meta().(*common.Client).OnCallClient

		found, _, err := client.UserNotificationRules.GetUserNotificationRule(rs.Primary.ID, &onCallAPI.GetUserNotificationRuleOptions{})
		if err != nil {
			return err
		}
		if found.ID != rs.Primary.ID {
			return fmt.Errorf("UserNotificationRule policy not found: %v - %v", rs.Primary.ID, found)
		}
		return nil
	}
}
