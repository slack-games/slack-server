package commands

import "github.com/slack-games/slack-client"

// PingCommand ping back
func PingCommand() slack.ResponseMessage {
	buttons := []slack.Action{
		slack.Action{
			Name:  "Test",
			Text:  "Test",
			Type:  "button",
			Value: "test",
			Style: slack.ActionPrimary,
		},
	}

	attachments := []slack.Attachment{
		slack.Attachment{
			Text:       "Test buttons",
			Fallback:   "Buttons to select",
			CallbackID: "first_row",
			Actions:    buttons,
		},
	}

	return slack.ResponseMessage{
		Text:        "You lucky found ping page",
		Attachments: attachments,
	}
}
