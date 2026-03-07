package slack

// severityIcon maps severity levels to Slack emoji icons.
func severityIcon(severity string) string {
	switch severity {
	case "critical":
		return ":rotating_light:"
	case "major":
		return ":warning:"
	case "minor":
		return ":large_yellow_circle:"
	case "resolved":
		return ":white_check_mark:"
	default:
		return ":information_source:"
	}
}

// SlackSendAlert sends a severity-formatted alert to a Slack webhook.
// Severity: "critical", "major", "minor", "resolved", or any other string.
// Includes a header with severity icon, divider, message body, and footer.
func SlackSendAlert(webhookURL, message, severity string) error {
	icon := severityIcon(severity)

	blocks := []any{
		SlackBlock(SlackHeader(icon + " Alert - " + severity)),
		SlackBlock(SlackDivider()),
		SlackBlock(SlackSection(message)),
		SlackBlock(SlackContext("Sent by Haira Agent")),
	}

	return SlackSendBlocks(webhookURL, blocks)
}
