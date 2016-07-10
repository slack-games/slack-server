package commands

import "github.com/slack-games/slack-client"

const helpText = `
To start a new game type _/ttt start_ or to see any existing _/ttt current_.
You play against the bot :robot_face:.
Make first move by typing _/ttt move cell-number_ - cell-number is from 1 to 9.
Example move would be _/ttt move 1_.

Good luck!
`

// HelpCommand show the possible info about available commands for user
func HelpCommand() slack.ResponseMessage {

	attachments := []slack.Attachment{
		slack.Attachment{
			Title: "/ttt start - starts a new game",
			Color: "#764FA5",
		},
		slack.Attachment{
			Title: "/ttt current - show the state of current game",
			Color: "#FF4F20",
		},
		slack.Attachment{
			Title: "/ttt move [1-9] - make move on the current board",
			Color: "#004FDD",
		},
		slack.Attachment{
			Title: "/ttt help - Shows help message",
			Color: "#76A0A0",
		},
	}

	return slack.ResponseMessage{
		Text:        helpText,
		Attachments: attachments,
	}
}
