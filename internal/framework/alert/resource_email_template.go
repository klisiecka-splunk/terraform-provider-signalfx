// Copyright Splunk, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwalert

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/signalfx/signalfx-go/emailtemplate"

	fwembed "github.com/splunk-terraform/terraform-provider-signalfx/internal/framework/embed"
	"github.com/splunk-terraform/terraform-provider-signalfx/internal/framework/fwerr"
	fwshared "github.com/splunk-terraform/terraform-provider-signalfx/internal/framework/shared"
)

type ResourceEmailTemplate struct {
	fwembed.ResourceData
	fwembed.ResourceIDImporter
}

type emailTemplateModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	TriggerSubject  types.String `tfsdk:"trigger_subject"`
	TriggerBody     types.String `tfsdk:"trigger_body"`
	ResolvedSubject types.String `tfsdk:"resolved_subject"`
	ResolvedBody    types.String `tfsdk:"resolved_body"`
	To              types.List   `tfsdk:"to"`
	Cc              types.List   `tfsdk:"cc"`
	Bcc             types.List   `tfsdk:"bcc"`
	CustomHeaders   types.Map    `tfsdk:"custom_headers"`
	CreatedOnMs     types.Int64  `tfsdk:"created_on_ms"`
	CreatedBy       types.String `tfsdk:"created_by"`
	UpdatedOnMs     types.Int64  `tfsdk:"updated_on_ms"`
	UpdatedBy       types.String `tfsdk:"updated_by"`
}

var (
	_ resource.Resource                = &ResourceEmailTemplate{}
	_ resource.ResourceWithConfigure   = &ResourceEmailTemplate{}
	_ resource.ResourceWithImportState = &ResourceEmailTemplate{}
)

func NewResourceEmailTemplate() resource.Resource {
	return &ResourceEmailTemplate{}
}

func (et *ResourceEmailTemplate) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email_template"
}

func (et *ResourceEmailTemplate) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a reusable detector alert email template.",
		Attributes: map[string]schema.Attribute{
			"id": fwshared.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the email template.",
			},
			"trigger_subject": schema.StringAttribute{
				Required:    true,
				Description: "Subject used when a detector alert triggers.",
			},
			"trigger_body": schema.StringAttribute{
				Required:    true,
				Description: "Body used when a detector alert triggers.",
			},
			"resolved_subject": schema.StringAttribute{
				Required:    true,
				Description: "Subject used when a detector alert resolves.",
			},
			"resolved_body": schema.StringAttribute{
				Required:    true,
				Description: "Body used when a detector alert resolves.",
			},
			"to": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Email addresses to include as template recipients.",
			},
			"cc": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Email addresses to include as carbon copy recipients.",
			},
			"bcc": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Email addresses to include as blind carbon copy recipients.",
			},
			"custom_headers": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Custom email headers to include when notifications use this template.",
			},
			"created_on_ms": schema.Int64Attribute{
				Computed:    true,
				Description: "Timestamp in milliseconds when the email template was created.",
			},
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "User that created the email template.",
			},
			"updated_on_ms": schema.Int64Attribute{
				Computed:    true,
				Description: "Timestamp in milliseconds when the email template was last updated.",
			},
			"updated_by": schema.StringAttribute{
				Computed:    true,
				Description: "User that last updated the email template.",
			},
		},
	}
}

func (et *ResourceEmailTemplate) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model emailTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := model.toEmailTemplate(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	details, err := et.Details().Client.CreateEmailTemplate(ctx, payload)
	if resp.Diagnostics.Append(fwerr.ErrorHandler(ctx, resp.State, err)...); resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(model.updateFromEmailTemplate(ctx, details)...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	}
}

func (et *ResourceEmailTemplate) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model emailTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	details, err := et.Details().Client.GetEmailTemplate(ctx, model.ID.ValueString())
	if resp.Diagnostics.Append(fwerr.ErrorHandler(ctx, resp.State, err)...); resp.Diagnostics.HasError() || err != nil {
		return
	}

	resp.Diagnostics.Append(model.updateFromEmailTemplate(ctx, details)...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	}
}

func (et *ResourceEmailTemplate) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model emailTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var prior emailTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.ID = prior.ID

	payload, diags := model.toEmailTemplate(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	details, err := et.Details().Client.UpdateEmailTemplate(ctx, model.ID.ValueString(), payload)
	if resp.Diagnostics.Append(fwerr.ErrorHandler(ctx, resp.State, err)...); resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(model.updateFromEmailTemplate(ctx, details)...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	}
}

func (et *ResourceEmailTemplate) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model emailTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := et.Details().Client.DeleteEmailTemplate(ctx, model.ID.ValueString())
	resp.Diagnostics.Append(fwerr.ErrorHandler(ctx, resp.State, err)...)
}

func (model emailTemplateModel) toEmailTemplate(ctx context.Context) (*emailtemplate.EmailTemplate, diag.Diagnostics) {
	var diags diag.Diagnostics
	details := &emailtemplate.EmailTemplate{
		Name:            model.Name.ValueString(),
		TriggerSubject:  model.TriggerSubject.ValueString(),
		TriggerBody:     model.TriggerBody.ValueString(),
		ResolvedSubject: model.ResolvedSubject.ValueString(),
		ResolvedBody:    model.ResolvedBody.ValueString(),
	}

	var fieldDiags diag.Diagnostics
	details.To, fieldDiags = fwshared.StringSliceFromList(ctx, model.To)
	diags.Append(fieldDiags...)
	if diags.HasError() {
		return details, diags
	}
	details.Cc, fieldDiags = fwshared.StringSliceFromList(ctx, model.Cc)
	diags.Append(fieldDiags...)
	if diags.HasError() {
		return details, diags
	}
	details.Bcc, fieldDiags = fwshared.StringSliceFromList(ctx, model.Bcc)
	diags.Append(fieldDiags...)
	if diags.HasError() {
		return details, diags
	}
	details.CustomHeaders, fieldDiags = fwshared.StringMapFromMap(ctx, model.CustomHeaders)
	diags.Append(fieldDiags...)

	return details, diags
}

func (model *emailTemplateModel) updateFromEmailTemplate(ctx context.Context, details *emailtemplate.EmailTemplate) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(details.Id)
	model.Name = types.StringValue(details.Name)
	model.TriggerSubject = types.StringValue(details.TriggerSubject)
	model.TriggerBody = types.StringValue(details.TriggerBody)
	model.ResolvedSubject = types.StringValue(details.ResolvedSubject)
	model.ResolvedBody = types.StringValue(details.ResolvedBody)
	model.CreatedOnMs = types.Int64Value(details.CreatedOnMs)
	model.CreatedBy = fwshared.OptionalStringValue(details.CreatedBy)
	model.UpdatedOnMs = types.Int64Value(details.UpdatedOnMs)
	model.UpdatedBy = fwshared.OptionalStringValue(details.UpdatedBy)

	model.To, diags = fwshared.StringListValue(ctx, details.To)
	if diags.HasError() {
		return diags
	}
	model.Cc, diags = fwshared.StringListValue(ctx, details.Cc)
	if diags.HasError() {
		return diags
	}
	model.Bcc, diags = fwshared.StringListValue(ctx, details.Bcc)
	if diags.HasError() {
		return diags
	}
	model.CustomHeaders, diags = fwshared.StringMapValue(ctx, details.CustomHeaders)

	return diags
}
