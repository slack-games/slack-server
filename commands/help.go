package commands

import "github.com/riston/slack-server/slack"

const helpText = `
To start a new game type */game start*. Now you'r opponent is a bot.
Make first move by typing */game move cell-number* - cell-number is from 1 to 9.
Example move would be */game move 1*.
`

// HelpCommand show the possible info about available commands for user
func HelpCommand() slack.ResponseMessage {
	return slack.ResponseMessage{
		Text: helpText,
		Attachments: []slack.Attachment{
			slack.Attachment{
				"/game start", "Starts new game", "", "", "#764FA5",
			},
			slack.Attachment{
				"/game current", "Show the state of current game", "", "", "#FF4F20",
			},
			slack.Attachment{
				"/game move [1-9]", "Make move on the current board", "", "", "#004FDD",
			},
			slack.Attachment{
				"/game help", "Shows the current help message", "", "", "#76A0A0",
			},
		},
	}
}
