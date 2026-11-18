package ctmsdestinations_test

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
)

// Summary fields from TEST_DO_NOT_COMMIT/02_SETUP/terraform.tfstate.
func testDestinationSummaryValues(name string) map[string]knownvalue.Check {
	return map[string]knownvalue.Check{
		"name":        knownvalue.StringExact(name),
		"description": knownvalue.StringExact(""),
		"platform_types": knownvalue.ListExact([]knownvalue.Check{
			knownvalue.StringExact(""),
		}),
		"application": knownvalue.Null(),
		"data_center": knownvalue.Null(),
		// Recording redacts destination URLs, so verify a populated HTTPS URL
		// for both the live response and the redacted replay.
		"url": knownvalue.StringRegexp(regexp.MustCompile(`^https://[^/[:space:]]+/.+$`)),
	}
}
