# 5. FusionReactor Alert Email Branding

**Problem:** Alert notification emails show Grafana branding (logo, company name, links). For the FusionReactor product, these need to be Intergral/FusionReactor branded.

**Solution:** Modify the alert notification HTML email template:
- Replace the Grafana logo with the FusionReactor logo
- Replace the fire emoji with an alert emoji
- Replace the Grafana Labs copyright footer with Intergral GmbH details
- Link to FusionReactor Cloud instead of Grafana

## Where to implement

- The HTML email template for alert notifications (typically `public/emails/ng_alert_notification.html` or equivalent)