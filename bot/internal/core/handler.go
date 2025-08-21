package core

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"hearsay/internal/commands"
	config "hearsay/internal/config"
	storage "hearsay/internal/storage"

	irc "github.com/fluffle/goirc/client"
	"golang.org/x/net/context"
)

var messagePool []storage.Message

func HearsayConnect(Server string, Channel string, ctx context.Context, db *sql.DB) {
	botNick := "hearsay" // TODO: Move to configure.go
	botUser := "hearsay"
	botMe := "hearsay"

	cfg := irc.NewConfig(botNick, botUser, botMe)

	// https://github.com/fluffle/goirc/blob/v1.3.1/client/connection.go#L144
	cfg.Version = "Bot"
	cfg.SSL = true
	cfg.SSLConfig = &tls.Config{InsecureSkipVerify: true}
	cfg.Server = Server

	c := irc.Client(cfg)

	quit := make(chan struct{})

	// These are handlers and WILL DO STUFF.
	c.HandleFunc(irc.CONNECTED,
		func(c *irc.Conn, l *irc.Line) {
			c.Join(Channel)
			c.Mode(botNick, config.BotMode)
			c.Away(config.CommandPrefix + "help for command list.")
			log.Printf("Joined %s\n", Channel)

			log.Println("Loading deletion scheduler...")
			go commands.DeletionWrapper(db, c, ctx)
		})

	c.HandleFunc(irc.PRIVMSG,
		func(c *irc.Conn, l *irc.Line) {
			incomingMessageAuthor := GetNickFromRawMessage(l.Raw)
			incomingMessageContent := GetContentFromRawMessage(l.Raw)
			incomingMessageChannel := GetChannelFromRawMessage(l.Raw)
			messageFinal := storage.Message{
				Nick:      incomingMessageAuthor,
				Content:   incomingMessageContent,
				Channel:   incomingMessageChannel,
				Timestamp: l.Time,
			}

			if strings.HasPrefix(incomingMessageContent, config.CommandPrefix) {
				// Case: The incoming message is preceded by our command prefix.
				commandAndArgs := strings.Split(incomingMessageContent, " ")
				receivedCommand := strings.Split(commandAndArgs[0], config.CommandPrefix)[1]
				receivedArgs := commandAndArgs[1:]
				log.Printf("Received command %s by %s.\n", receivedCommand, incomingMessageAuthor)

				go func(rCmd string, rArgs []string, rAuthor string, rChannel string) {
					if cmd, ok := commands.Commands[rCmd]; ok {
						result := cmd.Handler(rArgs, rAuthor, db)
						if result != "" {
							c.Privmsg(rChannel, result)
						}
					} else {
						c.Privmsgf(rChannel, "No such command: %s", rCmd)
					}
				}(receivedCommand, receivedArgs, incomingMessageAuthor, incomingMessageChannel)
			} else if storage.IsOptedIn(incomingMessageAuthor) {
				// Case: The incoming message is not preceded by our command prefix and the nick is not opted out.
				messagePool = append(messagePool, messageFinal)
				if len(messagePool) >= config.MaxMessagePool {
					tempPool := make([]storage.Message, len(messagePool))
					copy(tempPool, messagePool)
					messagePool = nil

					go func(pool []storage.Message) {
						err := storage.SubmitMessages(pool, db)
						if err != nil {
							log.Printf("Failed to submit messages: %v\n", err)
						} else {
							log.Printf("Wrote %d/%d messages to database.\n", len(pool), config.MaxMessagePool)
						}
					}(tempPool)
				}
			}
		})

	c.HandleFunc(irc.INVITE,
		func(c *irc.Conn, l *irc.Line) {
			channelToJoin := GetChannelFromInvite(l.Raw)
			c.Join(channelToJoin)
			log.Printf("Joined channel %s\n", channelToJoin)
		})

	c.HandleFunc(irc.DISCONNECTED,
		func(c *irc.Conn, l *irc.Line) {
			close(quit)
		})

	if err := c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())
		return
	}

	select {
	case <-ctx.Done():
		log.Println("Conext canceled. Sending QUIT...")
		c.Quit("Signing off.")
		c.Close()

	case <-quit:
		log.Println("Received server-side disconnect (such as /kill or unavailability)")
	}
}
