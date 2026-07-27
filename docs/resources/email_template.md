---
page_title: "Observability Cloud: signalfx_email_template"
description: |-
  Allows Terraform to create and manage email templates for Splunk Observability Cloud detector alerts
---
# Resource: signalfx_email_template

Email templates are reusable detector alert notification templates.
The optional `custom_headers` map sends additional email headers with messages that use the template. Use it for headers that your email infrastructure already recognizes, such as routing, classification, or downstream processing headers.

## Example

```terraform
resource "signalfx_email_template" "detector_alerts" {
  name = "Detector Alert Email"

  trigger_subject = "Triggered: {{{detectorName}}}"

  to = ["primary@example.com"]
  cc = ["team@example.com"]

  custom_headers = {
    X-Custom-Routing-Key = "detector-alerts"
  }
}
```

## Arguments

* `name` - (Required) Name of the email template.
* `trigger_subject` - (Required) Subject used when a detector alert triggers.
* `to` - (Required) Email addresses to include as template recipients. At least one address is required.
* `trigger_body` - (Optional) Body used when a detector alert triggers.
* `resolved_subject` - (Optional) Subject used when a detector alert resolves.
* `resolved_body` - (Optional) Body used when a detector alert resolves.
* `cc` - (Optional) Email addresses to include as carbon copy recipients.
* `bcc` - (Optional) Email addresses to include as blind carbon copy recipients.
* `custom_headers` - (Optional) Custom email headers to include when notifications use this template.

## Attributes

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the email template.
* `created_on_ms` - Timestamp in milliseconds when the email template was created.
* `created_by` - User that created the email template.
* `updated_on_ms` - Timestamp in milliseconds when the email template was last updated.
* `updated_by` - User that last updated the email template.
