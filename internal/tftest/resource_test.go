// Copyright Splunk, Inc.
// SPDX-License-Identifier: MPL-2.0

package tftest

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	tfext "github.com/splunk-terraform/terraform-provider-signalfx/internal/tfextension"
)

func TestResourceOperationTestCaseCreate(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		resource *schema.Resource
		expect   diag.Diagnostics
	}{
		{
			name:     "no create methods set",
			resource: &schema.Resource{},
			expect: diag.Diagnostics{
				{Severity: diag.Error, Summary: "no create operation defined"},
			},
		},
		{
			name: "create method set",
			resource: &schema.Resource{
				Create: func(*schema.ResourceData, any) error {
					return nil
				},
			},
			expect: nil,
		},
		{
			name: "create method fails",
			resource: &schema.Resource{
				Create: func(*schema.ResourceData, any) error {
					return errors.New("failed")
				},
			},
			expect: diag.Diagnostics{
				{Severity: diag.Error, Summary: "failed"},
			},
		},
		{
			name: "create context method set",
			resource: &schema.Resource{
				CreateContext: func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
					return nil
				},
			},
			expect: nil,
		},
		{
			name: "create context method fails",
			resource: &schema.Resource{
				CreateContext: func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
					return tfext.AsWarnDiagnostics(errors.New("warn"))
				},
			},
			expect: diag.Diagnostics{
				{Severity: diag.Warning, Summary: "warn"},
			},
		},
		{
			name: "create without timeout set",
			resource: &schema.Resource{
				CreateWithoutTimeout: func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
					return nil
				},
			},
		},
		{
			name: "create without timeout fails",
			resource: &schema.Resource{
				CreateWithoutTimeout: func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
					return diag.Errorf("failed")
				},
			},
			expect: diag.Diagnostics{
				{Severity: diag.Error, Summary: "failed"},
			},
		},
	} {
		tcase := &ResourceOperationTestCase[any]{
			Name:     tc.name,
			Resource: tc.resource,
			Encoder: func(*any, *schema.ResourceData) error {
				return nil
			},
			Decoder: tfext.NopDecodeTerraform[any],
			Meta: func(testing.TB) any {
				return nil
			},
			Issues: tc.expect,
		}

		tcase.TestCreate(t)
	}
}

func TestResourceOperationTestCaseRead(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		resource *schema.Resource
		expect   diag.Diagnostics
	}{
		{
			name:     "no read methods set",
			resource: &schema.Resource{},
			expect: diag.Diagnostics{
				{Severity: diag.Error, Summary: "no read operation defined"},
			},
		},
		{
			name: "read method set",
			resource: &schema.Resource{
				Read: func(*schema.ResourceData, any) error {
					return nil
				},
			},
			expect: nil,
		},
		{
			name: "read method fails",
			resource: &schema.Resource{
				Read: func(*schema.ResourceData, any) error {
					return errors.New("failed")
				},
			},
			expect: diag.Diagnostics{
				{Severity: diag.Error, Summary: "failed"},
			},
		},
		{
			name: "read context method set",
			resource: &schema.Resource{
				ReadContext: func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
					return nil
				},
			},
			expect: nil,
		},
		{
			name: "read context method fails",
			resource: &schema.Resource{
				ReadContext: func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
					return tfext.AsWarnDiagnostics(errors.New("warn"))
				},
			},
			expect: diag.Diagnostics{
				{Severity: diag.Warning, Summary: "warn"},
			},
		},
		{
			name: "read without timeout set",
			resource: &schema.Resource{
				ReadWithoutTimeout: func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
					return nil
				},
			},
		},
		{
			name: "read without timeout fails",
			resource: &schema.Resource{
				ReadWithoutTimeout: func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
					return diag.Errorf("failed")
				},
			},
			expect: diag.Diagnostics{
				{Severity: diag.Error, Summary: "failed"},
			},
		},
	} {
		tcase := &ResourceOperationTestCase[any]{
			Name:     tc.name,
			Resource: tc.resource,
			Encoder:  tfext.NopEncodeTerraform[any],
			Decoder:  tfext.NopDecodeTerraform[any],
			Meta: func(testing.TB) any {
				return nil
			},
			Issues: tc.expect,
		}

		tcase.TestRead(t)
	}
}

func TestResourceOperationTestCaseUpdate(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		resource *schema.Resource
		expect   diag.Diagnostics
	}{
		{
			name:     "no update methods set",
			resource: &schema.Resource{},
			expect: diag.Diagnostics{
				{Severity: diag.Error, Summary: "no update operation defined"},
			},
		},
		{
			name: "update method set",
			resource: &schema.Resource{
				Update: func(*schema.ResourceData, any) error {
					return nil
				},
			},
			expect: nil,
		},
		{
			name: "update method fails",
			resource: &schema.Resource{
				Update: func(*schema.ResourceData, any) error {
					return errors.New("failed")
				},
			},
			expect: diag.Diagnostics{
				{Severity: diag.Error, Summary: "failed"},
			},
		},
		{
			name: "update method set",
			resource: &schema.Resource{
				UpdateContext: func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
					return nil
				},
			},
			expect: nil,
		},
		{
			name: "update context method fails",
			resource: &schema.Resource{
				UpdateContext: func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
					return tfext.AsWarnDiagnostics(errors.New("warn"))
				},
			},
			expect: diag.Diagnostics{
				{Severity: diag.Warning, Summary: "warn"},
			},
		},
		{
			name: "update without timeout set",
			resource: &schema.Resource{
				UpdateWithoutTimeout: func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
					return nil
				},
			},
		},
		{
			name: "update without timeout fails",
			resource: &schema.Resource{
				UpdateWithoutTimeout: func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
					return diag.Errorf("failed")
				},
			},
			expect: diag.Diagnostics{
				{Severity: diag.Error, Summary: "failed"},
			},
		},
	} {
		tcase := &ResourceOperationTestCase[any]{
			Name:     tc.name,
			Resource: tc.resource,
			Encoder:  tfext.NopEncodeTerraform[any],
			Decoder:  tfext.NopDecodeTerraform[any],
			Meta: func(testing.TB) any {
				return nil
			},
			Issues: tc.expect,
		}

		tcase.TestUpdate(t)
	}
}

func TestResourceOperationTestCaseDelete(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		resource *schema.Resource
		expect   diag.Diagnostics
	}{
		{
			name:     "no delete methods set",
			resource: &schema.Resource{},
			expect: diag.Diagnostics{
				{Severity: diag.Error, Summary: "no delete operation defined"},
			},
		},
		{
			name: "delete method set",
			resource: &schema.Resource{
				Delete: func(*schema.ResourceData, any) error {
					return nil
				},
			},
			expect: nil,
		},
		{
			name: "delete method fails",
			resource: &schema.Resource{
				Delete: func(*schema.ResourceData, any) error {
					return errors.New("failed")
				},
			},
			expect: diag.Diagnostics{
				{Severity: diag.Error, Summary: "failed"},
			},
		},
		{
			name: "delete method set",
			resource: &schema.Resource{
				DeleteContext: func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
					return nil
				},
			},
			expect: nil,
		},
		{
			name: "delete context method fails",
			resource: &schema.Resource{
				DeleteContext: func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
					return tfext.AsWarnDiagnostics(errors.New("warn"))
				},
			},
			expect: diag.Diagnostics{
				{Severity: diag.Warning, Summary: "warn"},
			},
		},
		{
			name: "delete without timeout set",
			resource: &schema.Resource{
				DeleteWithoutTimeout: func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
					return nil
				},
			},
		},
		{
			name: "delete without timeout fails",
			resource: &schema.Resource{
				DeleteWithoutTimeout: func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
					return diag.Errorf("failed")
				},
			},
			expect: diag.Diagnostics{
				{Severity: diag.Error, Summary: "failed"},
			},
		},
	} {
		tcase := &ResourceOperationTestCase[any]{
			Name:     tc.name,
			Resource: tc.resource,
			Encoder:  tfext.NopEncodeTerraform[any],
			Decoder:  tfext.NopDecodeTerraform[any],
			Meta: func(testing.TB) any {
				return nil
			},
			Issues: tc.expect,
		}

		tcase.TestDelete(t)
	}
}
