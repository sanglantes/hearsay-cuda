package core

import (
	"regexp"
	"strings"
)

func GetChannelFromInvite(rawMessage string) string {
	inviteSplit := strings.Split(rawMessage, ":") // TODO: SplitN in case channel-name has a colon.
	channel := inviteSplit[len(inviteSplit)-1]

	return channel
}

func GetNickFromRawMessage(rawMessage string) string {
	// :katt!kattkattkatt@172.17.0.1 PRIVMSG #boing :riiinky dinky
	// Split the first exclamation sign and slice past the colon to isolate the nickname.
	messageSplit := strings.Split(rawMessage, "!")
	nick := messageSplit[0][1:]

	return nick
}

func GetContentFromRawMessage(rawMessage string) string {
	// messageSplit will be a slice containing three strings.
	// The first colon is discarded by SplitN.
	// The hostmask and message information is sandwiched between the leading colon and the message.
	// The desired message is left as the third entry.
	messageSplit := strings.SplitN(rawMessage, ":", 3)
	//fmt.Printf("%#v\n", messageSplit)
	message := messageSplit[2]

	return message
}

func GetChannelFromRawMessage(rawMessage string) string {
	// Capture all characters between # till a space ( ) is met.
	re := regexp.MustCompile(`#\S+`)
	channel := re.FindString(rawMessage)

	return channel
}
