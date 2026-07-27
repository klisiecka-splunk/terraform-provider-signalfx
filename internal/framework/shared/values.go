// Copyright Splunk, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwshared

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func OptionalStringValue(value string) types.String {
	if value == "" {
		return types.StringNull()
	}

	return types.StringValue(value)
}

func StringListValue(ctx context.Context, values []string) (types.List, diag.Diagnostics) {
	if len(values) == 0 {
		return types.ListNull(types.StringType), nil
	}

	return types.ListValueFrom(ctx, types.StringType, values)
}

func StringMapValue(ctx context.Context, values map[string]string) (types.Map, diag.Diagnostics) {
	if len(values) == 0 {
		return types.MapNull(types.StringType), nil
	}

	return types.MapValueFrom(ctx, types.StringType, values)
}

func StringSliceFromList(ctx context.Context, value types.List) ([]string, diag.Diagnostics) {
	var values []string
	if value.IsNull() || value.IsUnknown() {
		return values, nil
	}

	diags := value.ElementsAs(ctx, &values, false)
	return values, diags
}

func StringMapFromMap(ctx context.Context, value types.Map) (map[string]string, diag.Diagnostics) {
	values := map[string]string{}
	if value.IsNull() || value.IsUnknown() {
		return values, nil
	}

	diags := value.ElementsAs(ctx, &values, false)
	return values, diags
}
