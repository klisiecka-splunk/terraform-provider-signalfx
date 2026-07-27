resource "signalfx_email_template" "test" {
  name = "Detector Alert Email Updated"

  trigger_subject = "Triggered: {{{detectorName}}}"

  to = [
    "primary@example.com",
    "secondary@example.com",
  ]
  cc = ["team@example.com"]

  custom_headers = {
    X-Custom-Routing-Key = "detector-alerts-updated"
  }
}
