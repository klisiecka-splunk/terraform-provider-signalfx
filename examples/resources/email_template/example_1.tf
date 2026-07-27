resource "signalfx_email_template" "detector_alerts" {
  name = "Detector Alert Email"

  trigger_subject = "Triggered: {{{detectorName}}}"

  to = ["primary@example.com"]
  cc = ["team@example.com"]

  custom_headers = {
    X-Custom-Routing-Key = "detector-alerts"
  }
}
