package grafana

import (
	"errors"
	"fmt"
	"net/url"
	"testing"

	gapi "github.com/grafana/grafana-api-golang-client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	testAccAnnotationInitialText string = "basic text"
	testAccAnnotationUpdatedText string = "basic text updated"
)

func TestAccAnnotation_basic(t *testing.T) {
	CheckOSSTestsEnabled(t)
	var annotation gapi.Annotation

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccAnnotationCheckDestroy(&annotation),
		Steps: []resource.TestStep{
			{
				// Test resource creation.
				Config: testAnnotationConfig(testAccAnnotationInitialText),
				Check: resource.ComposeTestCheckFunc(
					testAccAnnotationCheckExists("grafana_annotation.test", &annotation),
					resource.TestCheckResourceAttr("grafana_annotation.test", "text", testAccAnnotationInitialText),
				),
			},
			{
				// Updates text.
				Config: testAnnotationConfig(testAccAnnotationUpdatedText),
				Check: resource.ComposeTestCheckFunc(
					testAccAnnotationCheckExists("grafana_annotation.test", &annotation),
					resource.TestCheckResourceAttr("grafana_annotation.test", "text", testAccAnnotationUpdatedText),
				),
			},
			{
				// Importing matches the state of the previous step.
				ResourceName:      "grafana_annotation.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAnnotationCheckExists(rn string, annotation *gapi.Annotation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id not set")
		}

		client := testAccProvider.Meta().(*client).gapi
		annotations, err := client.Annotations(url.Values{})
		if err != nil {
			return fmt.Errorf("error getting annotation: %s", err)
		}

		if len(annotations) < 1 {
			return errors.New("Grafana API returned no annotations")
		}

		*annotation = annotations[0]

		return nil
	}
}

func testAccAnnotationCheckDestroy(annotation *gapi.Annotation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*client).gapi
		annotations, err := client.Annotations(url.Values{})
		if err != nil {
			return err
		}

		if len(annotations) < 1 {
			return nil
		}

		for _, annotation := range annotations {
			if annotation.Text == testAccAnnotationInitialText || annotation.Text == testAccAnnotationUpdatedText {
				return errors.New("annotation still exists")
			}
		}

		return nil
	}
}

func testAnnotationConfig(text string) string {
	return fmt.Sprintf(`
resource "grafana_annotation" "test" {
    text = "%s"
}
`, text)
}
