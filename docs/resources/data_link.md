---
page_title: "Splunk Observability Cloud: signalfx_data_link"
description: |-
  Allows Terraform to create and manage data links in Splunk Observability Cloud
---

# Resource: signalfx_data_link

Manage Splunk Observability Cloud [Data Links](https://docs.signalfx.com/en/latest/managing/data-links.html).

## Example

```terraform
# A global link to Splunk Observability Cloud dashboard.
resource "signalfx_data_link" "my_data_link" {
  property_name  = "pname"
  property_value = "pvalue"

  target_signalfx_dashboard {
    is_default         = true
    name               = "sfx_dash"
    dashboard_group_id = signalfx_dashboard_group.mydashboardgroup0.id
    dashboard_id       = signalfx_dashboard.mydashboard0.id
  }
}

# A dashboard-specific link to an external URL
resource "signalfx_data_link" "my_data_link_dash" {
  context_dashboard_id = signalfx_dashboard.mydashboard0.id
  property_name        = "pname2"
  property_value       = "pvalue"

  target_external_url {
    name        = "ex_url"
    time_format = "ISO8601"
    url         = "https://www.example.com"
    property_key_mapping = {
      foo = "bar"
    }
  }
}

# A link to an AppDynamics Service
resource "signalfx_data_link" "my_data_link_appd" {
  property_name        = "pname3"
  property_value       = "pvalue"

  target_appd_url {
    name        = "appd_url"
    url         = "https://www.example.saas.appdynamics.com/#/application=1234&component=5678"
  }
}
```

## Arguments

The following arguments are supported in the resource block:

* `property_name` - (Optional) Name (key) of the metadata that's the trigger of a data link. If you specify `property_value`, you must specify `property_name`.
* `property_value` - (Optional) Value of the metadata that's the trigger of a data link. If you specify this property, you must also specify `property_name`.
* `context_dashboard_id` - (Optional) If provided, scopes this data link to the supplied dashboard id. If omitted then the link will be global.
* `target_external_url` - (Optional) Link to an external URL
  * `name` (Required) User-assigned target name. Use this value to differentiate between the link targets for a data link object.
  * `url`- (Required) URL string for a Splunk instance or external system data link target. [See the supported template variables](https://dev.splunk.com/observability/docs/administration/datalinks/).
  * `time_format` - (Optional) [Designates the format](https://dev.splunk.com/observability/docs/administration/datalinks/) of `minimum_time_window` in the same data link target object. Must be one of `"ISO8601"`, `"EpochSeconds"` or `"Epoch"` (which is milliseconds). Defaults to `"ISO8601"`.
  * `minimum_time_window` - (Optional) The [minimum time window](https://dev.splunk.com/observability/docs/administration/datalinks/) for a search sent to an external site. Defaults to `6000`
  * `property_key_mapping` - Describes the relationship between Splunk Observability Cloud metadata keys and external system properties when the key names are different.
* `target_signalfx_dashboard` - (Optional) Link to a Splunk Observability Cloud dashboard
  * `name` (Required) User-assigned target name. Use this value to differentiate between the link targets for a data link object.
  * `is_default` - (Optional) Flag that designates a target as the default for a data link object. `true` by default
  * `dashboard_id` - (Required) SignalFx-assigned ID of the dashboard link target
  * `dashboard_group_id` - (Required) SignalFx-assigned ID of the dashboard link target's dashboard group
* `target_splunk` - (Optional) Link to an external URL
  * `name` (Required) User-assigned target name. Use this value to differentiate between the link targets for a data link object.
  * `property_key_mapping` - Describes the relationship between Splunk Observability Cloud metadata keys and external system properties when the key names are different.
* `target_appd_url` - (Optional) Link to an AppDynamics URL
  * `name` (Required) User-assigned target name. Use this value to differentiate between the link targets for a data link object.
  * `url`- (Required) URL string for an AppDynamics instance.

## Attributes

In a addition to all arguments above, the following attributes are exported:

* `id` - The ID of the link.
