// btpservices/provider/cicd/settings/datasource_allowed_spaces.go
package cicdsettings

import (
	cicdmodels "github.com/SAP/terraform-provider-btp-services/internal/cicd/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// allowedSpacesModel is the Terraform state model.
type allowedSpacesModel struct {
	AllowedSpaces []allowedSpaceEntry `tfsdk:"allowed_spaces"`
}

type allowedSpaceEntry struct {
	SpaceGUID types.String `tfsdk:"space_guid"`
	Comment   types.String `tfsdk:"comment"`
}

func allowedSpacesValueFrom(v cicdmodels.AllowedSpacesResponse) allowedSpacesModel {
	entries := make([]allowedSpaceEntry, 0, len(v.AllowedSpaces))
	for _, s := range v.AllowedSpaces {
		entries = append(entries, allowedSpaceEntry{
			SpaceGUID: types.StringValue(s.SpaceGUID),
			Comment:   types.StringValue(s.Comment),
		})
	}
	return allowedSpacesModel{
		AllowedSpaces: entries,
	}
}

func (m allowedSpacesModel) toRequest() cicdmodels.AllowedSpaceListDTO {
	spaces := make([]cicdmodels.AllowedSpace, 0, len(m.AllowedSpaces))
	for _, e := range m.AllowedSpaces {
		spaces = append(spaces, cicdmodels.AllowedSpace{
			SpaceGUID: e.SpaceGUID.ValueString(),
			Comment:   e.Comment.ValueString(),
		})
	}
	return cicdmodels.AllowedSpaceListDTO{AllowedSpaces: spaces}
}
