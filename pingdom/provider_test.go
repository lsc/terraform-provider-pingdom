package pingdom

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	// testAccProviderFactories is the ProviderFactories form of the provider
	// map. TestCase.Providers is deprecated; a factory is called per test so
	// each one gets a provider that has not been configured by another.
	testAccProviderFactories map[string]func() (*schema.Provider, error)
	testAccProvider          *schema.Provider
)

func init() {
	testAccProvider = Provider()
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"pingdom": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderConfigure(t *testing.T) {
	var expectedToken string

	var isAccTestEnabled bool
	if v := os.Getenv("TF_ACC"); v != "" {
		isAccTestEnabled = true
	}
	if v := os.Getenv("PINGDOM_API_TOKEN"); v != "" {
		expectedToken = v
	} else {
		expectedToken = "foo"
	}

	raw := map[string]any{
		"api_token": expectedToken,
	}

	// Previously there is only one client, which is the Pingdom client. It does not require obtaining any kind of
	// token during its initialization process, thus it will not verify whether the token provided is valid or not.
	// However, the case is different for the Solarwinds client because it will not initialize successfully unless
	// there are real user credentials provided.	In this case, we need to skip the init process to avoid any test
	// errors if the credentials are not provided.
	if isAccTestEnabled {
		rp := Provider()
		err := rp.Configure(context.Background(), terraform.NewResourceConfigRaw(raw))
		if err != nil {
			t.Fatal(err)
		}

		pingdomClient := rp.Meta().(*Clients).Pingdom

		if pingdomClient.APIToken != expectedToken {
			t.Fatalf("bad: %#v", pingdomClient)
		}
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("PINGDOM_API_TOKEN"); v == "" {
		t.Fatal("PINGDOM_API_TOKEN environment variable must be set for acceptance tests")
	}
	if v := os.Getenv("SOLARWINDS_USER"); v == "" {
		t.Fatal("SOLARWINDS_USER environment variable must be set for acceptance tests")
	}
	if v := os.Getenv("SOLARWINDS_PASSWD"); v == "" {
		t.Fatal("SOLARWINDS_PASSWD environment variable must be set for acceptance tests")
	}
	if v := os.Getenv("SOLARWINDS_ORG_ID"); v == "" {
		t.Fatal("SOLARWINDS_ORG_ID environment variable must be set for acceptance tests")
	}
}

func testAccCheckPingdomResourceID(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Can't find: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID not set: %s", name)
		}
		return nil
	}
}
