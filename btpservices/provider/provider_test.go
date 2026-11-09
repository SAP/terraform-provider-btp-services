// btpservices/provider/provider_test.go

package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	tftest "github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

func providerFactoryDev() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"btpservice": providerserver.NewProtocol6WithError(
			New()(),
		),
	}
}

func providerFactoryWithClients() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"btpservice": providerserver.NewProtocol6WithError(
			NewWithClients(&shared.ProviderClients{}),
		),
	}
}

func TestProvider_Schema(t *testing.T) {
	tftest.UnitTest(
		t,
		tftest.TestCase{
			ProtoV6ProviderFactories: providerFactoryDev(),
			Steps: []tftest.TestStep{
				{
					Config: `provider "btpservice" {}`,
				},
			},
		},
	)
}

func TestProvider_CicdBlock(t *testing.T) {
	tftest.UnitTest(
		t,
		tftest.TestCase{
			ProtoV6ProviderFactories: providerFactoryDev(),
			Steps: []tftest.TestStep{
				{
					Config: `
provider "btpservice" {
  cicd {
    endpoint      = "https://cicd-service.cfapps.eu12.hana.ondemand.com"
    token_url     = "https://example.authentication.eu12.hana.ondemand.com/oauth/token"
    client_id     = "test-client-id"
    client_secret = "test-client-secret"
    timeout       = 60
  }
}`,
				},
			},
		},
	)
}

func TestProvider_CicdBlock_CustomTimeout(t *testing.T) {
	tftest.UnitTest(
		t,
		tftest.TestCase{
			ProtoV6ProviderFactories: providerFactoryDev(),
			Steps: []tftest.TestStep{
				{
					Config: `
provider "btpservice" {
  cicd {
    endpoint      = "https://cicd-service.cfapps.eu12.hana.ondemand.com"
    token_url     = "https://example.authentication.eu12.hana.ondemand.com/oauth/token"
    client_id     = "test-client-id"
    client_secret = "test-client-secret"
    timeout       = 120
  }
}`,
				},
			},
		},
	)
}

func TestProvider_EnvVarFallback(t *testing.T) {
	t.Setenv("BTP_CICD_ENDPOINT", "https://cicd-service.cfapps.eu12.hana.ondemand.com")
	t.Setenv("BTP_CICD_TOKEN_URL", "https://example.authentication.eu12.hana.ondemand.com/oauth/token")
	t.Setenv("BTP_CICD_CLIENT_ID", "test-client-id")
	t.Setenv("BTP_CICD_CLIENT_SECRET", "test-client-secret")

	tftest.UnitTest(
		t,
		tftest.TestCase{
			ProtoV6ProviderFactories: providerFactoryDev(),
			Steps: []tftest.TestStep{
				{
					Config: `
provider "btpservice" {
  cicd {}
}`,
				},
			},
		},
	)
}

func TestProvider_NewWithClients(t *testing.T) {
	tftest.UnitTest(
		t,
		tftest.TestCase{
			ProtoV6ProviderFactories: providerFactoryWithClients(),
			Steps: []tftest.TestStep{
				{
					Config: `provider "btpservice" {}`,
				},
			},
		},
	)
}

// ---------------------------------------------------------------------------
// Resources() / DataSources() registration tests
// ---------------------------------------------------------------------------

// expectedResourceTypes lists every resource type name the provider must expose.
var expectedResourceTypes = []string{
	"btpservice_ctms_node",
	"btpservice_ctms_route",
	"btpservice_cicd_credential_basic_auth",
	"btpservice_cicd_credential_cloud_connector",
	"btpservice_cicd_credential_webhook_secret",
	"btpservice_cicd_credential_container_registry",
	"btpservice_cicd_credential_kubernetes_config",
	"btpservice_cicd_credential_cert_based_auth_custom_idp",
	"btpservice_cicd_credential_service_key",
	"btpservice_cicd_credential_secret_text",
	"btpservice_cicd_credential_basic_auth_custom_idp",
	"btpservice_cicd_repository",
	"btpservice_cicd_job",
	"btpservice_cicd_trigger",
	"btpservice_cicd_allowed_spaces",
}

// expectedDataSourceTypes lists every data source type name the provider must expose.
var expectedDataSourceTypes = []string{
	"btpservice_ctms_node",
	"btpservice_ctms_nodes",
	"btpservice_ctms_route",
	"btpservice_ctms_routes",
	"btpservice_ctms_destination",
	"btpservice_ctms_destinations",
	"btpservice_ctms_ui_url",
	"btpservice_ctms_transport_request_logs_url",
	"btpservice_cicd_credential",
	"btpservice_cicd_credentials",
	"btpservice_cicd_credential_usage",
	"btpservice_cicd_job_credentials",
	"btpservice_cicd_repositories",
	"btpservice_cicd_repository",
	"btpservice_cicd_repository_event_receiver",
	"btpservice_cicd_repository_jobs",
	"btpservice_cicd_repository_webhook_config",
	"btpservice_cicd_trigger",
	"btpservice_cicd_triggers",
	"btpservice_cicd_builds",
	"btpservice_cicd_allowed_spaces",
}

// expectedListResourceTypes lists every list resource type name the provider must expose.
var expectedListResourceTypes = []string{
	"btpservice_ctms_node",
	"btpservice_ctms_route",
	"btpservice_cicd_job",
	"btpservice_cicd_trigger",
	"btpservice_cicd_repository",
}

func TestProvider_Resources_TypeNames(t *testing.T) {
	p := New()()
	ctx := context.Background()

	got := make(map[string]bool)
	for _, factory := range p.Resources(ctx) {
		r := factory()
		var meta resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "btpservice"}, &meta)
		got[meta.TypeName] = true
	}

	for _, name := range expectedResourceTypes {
		if !got[name] {
			t.Errorf("missing resource type %q", name)
		}
	}
}

func TestProvider_DataSources_TypeNames(t *testing.T) {
	p := New()()
	ctx := context.Background()

	got := make(map[string]bool)
	for _, factory := range p.DataSources(ctx) {
		ds := factory()
		var meta datasource.MetadataResponse
		ds.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "btpservice"}, &meta)
		got[meta.TypeName] = true
	}

	for _, name := range expectedDataSourceTypes {
		if !got[name] {
			t.Errorf("missing data source type %q", name)
		}
	}
}

func TestProvider_ListResources_TypeNames(t *testing.T) {
	p, ok := New()().(interface {
		ListResources(context.Context) []func() list.ListResource
	})
	if !ok {
		t.Fatal("provider does not implement ProviderWithListResources")
	}
	ctx := context.Background()

	got := make(map[string]bool)
	for _, factory := range p.ListResources(ctx) {
		lr := factory()
		var meta resource.MetadataResponse
		lr.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "btpservice"}, &meta)
		got[meta.TypeName] = true
	}

	for _, name := range expectedListResourceTypes {
		if !got[name] {
			t.Errorf("missing list resource type %q", name)
		}
	}
}

// expectedActionTypes lists every action type name the provider must expose.
var expectedActionTypes = []string{
	"btpservice_cicd_run_build",
	"btpservice_cicd_abort_build",
	"btpservice_cicd_delete_build",
}

func TestProvider_Actions_TypeNames(t *testing.T) {
	p, ok := New()().(interface {
		Actions(context.Context) []func() action.Action
	})
	if !ok {
		t.Fatal("provider does not implement ProviderWithActions")
	}
	ctx := context.Background()

	got := make(map[string]bool)
	for _, factory := range p.Actions(ctx) {
		a := factory()
		var meta action.MetadataResponse
		a.Metadata(ctx, action.MetadataRequest{ProviderTypeName: "btpservice"}, &meta)
		got[meta.TypeName] = true
	}

	for _, name := range expectedActionTypes {
		if !got[name] {
			t.Errorf("missing action type %q", name)
		}
	}
}

// CTMS provider configuration tests make no HTTP requests. Landscape behavior
// is covered separately by the live-recorded acceptance tests.
func TestProvider_CTMSConfiguration(t *testing.T) {
	valid := func() ctmsProviderModel {
		return ctmsProviderModel{
			Endpoint:     types.StringValue("https://ctms.example/"),
			TokenURL:     types.StringValue("https://auth.example/oauth/token"),
			ClientID:     types.StringValue("test-client-id"),
			ClientSecret: types.StringValue("test-client-secret"),
			SubaccountID: types.StringValue("test-subaccount"),
			NamedUser:    types.StringNull(),
			Timeout:      types.Int64Null(),
		}
	}
	cases := []struct {
		name      string
		edit      func(*ctmsProviderModel)
		version   string
		wantError bool
	}{
		{name: "optional timeout omitted"},
		{
			name: "custom timeout",
			edit: func(c *ctmsProviderModel) {
				c.Timeout = types.Int64Value(120)
			},
		},
		{
			name: "zero timeout",
			edit: func(c *ctmsProviderModel) {
				c.Timeout = types.Int64Value(0)
			},
			wantError: true,
		},
		{
			name: "negative timeout",
			edit: func(c *ctmsProviderModel) {
				c.Timeout = types.Int64Value(-1)
			},
			wantError: true,
		},
		{
			name: "overflow timeout",
			edit: func(c *ctmsProviderModel) {
				c.Timeout = types.Int64Value(1<<63 - 1)
			},
			wantError: true,
		},
		{
			name: "unknown timeout",
			edit: func(c *ctmsProviderModel) {
				c.Timeout = types.Int64Unknown()
			},
			wantError: true,
		},
		{
			name: "endpoint includes API version",
			edit: func(c *ctmsProviderModel) {
				c.Endpoint = types.StringValue("https://ctms.example/v2")
			},
			wantError: true,
		},
		{
			name: "invalid URL",
			edit: func(c *ctmsProviderModel) {
				c.Endpoint = types.StringValue("not-a-url")
			},
			wantError: true,
		},
		{
			name: "unknown secret",
			edit: func(c *ctmsProviderModel) {
				c.ClientSecret = types.StringUnknown()
			},
			wantError: true,
		},
		{
			name: "empty secret",
			edit: func(c *ctmsProviderModel) {
				c.ClientSecret = types.StringValue("")
			},
			wantError: true,
		},
		{
			name: "empty scope",
			edit: func(c *ctmsProviderModel) {
				c.SubaccountID = types.StringValue("")
			},
			wantError: true,
		},
		{
			name: "named user",
			edit: func(c *ctmsProviderModel) {
				c.NamedUser = types.StringValue("terraform-test")
			},
		},
		{
			name: "header injection",
			edit: func(c *ctmsProviderModel) {
				c.NamedUser = types.StringValue("actor\r\ninjected: value")
			},
			wantError: true,
		},
		{
			name: "long named user",
			edit: func(c *ctmsProviderModel) {
				c.NamedUser = types.StringValue(strings.Repeat("a", 61))
			},
			wantError: true,
		},
		{
			name:      "older Terraform",
			version:   "1.13.0",
			wantError: true,
		},
	}
	for _, tt := range cases {
		t.Run(
			tt.name,
			func(t *testing.T) {
				data := valid()
				if tt.edit != nil {
					tt.edit(&data)
				}
				tfVersion := tt.version
				if tfVersion == "" {
					tfVersion = "1.14.0"
				}
				clients := &shared.ProviderClients{}
				var diagnostics diag.Diagnostics
				configureCTMS(context.Background(), &data, tfVersion, clients, &diagnostics)
				if diagnostics.HasError() != tt.wantError {
					t.Fatalf("unexpected diagnostics: %v", diagnostics)
				}
				if tt.wantError {
					if clients.Ctms != nil {
						t.Fatal("invalid configuration created a client")
					}
					return
				}
				if clients.Ctms == nil || clients.CtmsScope.Endpoint != "https://ctms.example" || clients.CtmsScope.SubaccountID != "test-subaccount" {
					t.Fatalf("unexpected CTMS client or scope: %#v", clients.CtmsScope)
				}
				if clients.Cicd != nil {
					t.Fatal("CTMS configuration unexpectedly configured CI/CD")
				}
			},
		)
	}
}

func TestProvider_CTMSEnvironmentFallback(t *testing.T) {
	for key, value := range map[string]string{
		"ENDPOINT":      "https://ctms.example/",
		"TOKEN_URL":     "https://auth.example/oauth/token",
		"CLIENT_ID":     "test-client-id",
		"CLIENT_SECRET": "test-client-secret",
		"SUBACCOUNT_ID": "env-subaccount",
		"NAMED_USER":    "env-user",
	} {
		t.Setenv("BTP_CTMS_"+key, value)
	}
	data := ctmsProviderModel{
		Endpoint:     types.StringNull(),
		TokenURL:     types.StringNull(),
		ClientID:     types.StringNull(),
		ClientSecret: types.StringNull(),
		SubaccountID: types.StringNull(),
		NamedUser:    types.StringNull(),
		Timeout:      types.Int64Null(),
	}
	first := &shared.ProviderClients{}
	var diagnostics diag.Diagnostics
	configureCTMS(context.Background(), &data, "1.14.0", first, &diagnostics)
	if diagnostics.HasError() || first.Ctms == nil || first.CtmsScope.SubaccountID != "env-subaccount" {
		t.Fatalf("environment fallback failed: %v", diagnostics)
	}
	data.SubaccountID = types.StringValue("explicit-subaccount")
	data.NamedUser = types.StringValue("") // Explicit empty disables the environment attribution.
	second := &shared.ProviderClients{}
	configureCTMS(context.Background(), &data, "1.14.0", second, &diagnostics)
	if diagnostics.HasError() || second.Ctms == nil || second.Ctms == first.Ctms || second.CtmsScope.SubaccountID != "explicit-subaccount" || first.CtmsScope.SubaccountID != "env-subaccount" {
		t.Fatalf("explicit configuration or client isolation failed: %v", diagnostics)
	}
	data.ClientSecret = types.StringUnknown() // Unknown must not silently fall back to the environment.
	configureCTMS(context.Background(), &data, "1.14.0", &shared.ProviderClients{}, &diagnostics)
	if !diagnostics.HasError() {
		t.Fatal("unknown explicit credentials fell back to environment")
	}
}
