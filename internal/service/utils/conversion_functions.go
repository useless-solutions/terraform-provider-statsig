package utils

import "github.com/hashicorp/terraform-plugin-framework/types"

func ConvertStringList(frameworkStrings []types.String) []string {
	var strings []string
	for _, s := range frameworkStrings {
		strings = append(strings, s.ValueString())
	}
	return strings
}
