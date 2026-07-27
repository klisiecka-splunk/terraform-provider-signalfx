// Copyright Splunk, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwshared

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionalStringValue(t *testing.T) {
	t.Parallel()

	assert.True(t, OptionalStringValue("").IsNull())
	assert.Equal(t, "value", OptionalStringValue("value").ValueString())
}

func TestStringListValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	empty, diags := StringListValue(ctx, nil)
	require.False(t, diags.HasError())
	assert.True(t, empty.IsNull())

	value, diags := StringListValue(ctx, []string{"one", "two"})
	require.False(t, diags.HasError())

	var got []string
	diags = value.ElementsAs(ctx, &got, false)
	require.False(t, diags.HasError())
	assert.Equal(t, []string{"one", "two"}, got)
}

func TestStringMapValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	empty, diags := StringMapValue(ctx, nil)
	require.False(t, diags.HasError())
	assert.True(t, empty.IsNull())

	value, diags := StringMapValue(ctx, map[string]string{"name": "value"})
	require.False(t, diags.HasError())

	var got map[string]string
	diags = value.ElementsAs(ctx, &got, false)
	require.False(t, diags.HasError())
	assert.Equal(t, map[string]string{"name": "value"}, got)
}

func TestStringSliceFromList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	got, diags := StringSliceFromList(ctx, types.ListNull(types.StringType))
	require.False(t, diags.HasError())
	assert.Empty(t, got)

	list, diags := types.ListValueFrom(ctx, types.StringType, []string{"one", "two"})
	require.False(t, diags.HasError())

	got, diags = StringSliceFromList(ctx, list)
	require.False(t, diags.HasError())
	assert.Equal(t, []string{"one", "two"}, got)
}

func TestStringMapFromMap(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	got, diags := StringMapFromMap(ctx, types.MapNull(types.StringType))
	require.False(t, diags.HasError())
	assert.Empty(t, got)

	value, diags := types.MapValueFrom(ctx, types.StringType, map[string]string{"name": "value"})
	require.False(t, diags.HasError())

	got, diags = StringMapFromMap(ctx, value)
	require.False(t, diags.HasError())
	assert.Equal(t, map[string]string{"name": "value"}, got)
}
