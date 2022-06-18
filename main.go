package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

func main() {
	// Get token from the environment variable
	token := os.Getenv("TOKEN")
	if token == "" {
        panic("No token")
	}

	// Create bot from environment value.
	b, err := gotgbot.NewBot(token, &gotgbot.BotOpts{
		Client: http.Client{},
		DefaultRequestOpts: &gotgbot.RequestOpts{
			Timeout: gotgbot.DefaultTimeout,
			APIURL:  gotgbot.DefaultAPIURL,
		},
	})
	if err != nil {
		panic("failed to create new bot: " + err.Error())
	}

	// Create updater and dispatcher.
	updater := ext.NewUpdater(&ext.UpdaterOpts{
		ErrorLog: nil,
		DispatcherOpts: ext.DispatcherOpts{
			// If an error is returned by a handler, log it and continue going.
			Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
				fmt.Println("an error occurred while handling update:", err.Error())
				return ext.DispatcherActionNoop
			},
			MaxRoutines: ext.DefaultMaxRoutines,
		},
	})
	dispatcher := updater.Dispatcher

	// Add echo handler to reply to all text messages.
	dispatcher.AddHandler(handlers.Message{
		AllowChannel: true,
		Filter: func(msg *gotgbot.Message) bool {
			return msg != nil && msg.MediaGroupId != "" && msg.Chat.Type == "channel"
		},
		Response: Dowork,
	})
	dispatcher.AddHandler(handlers.NewCommand("start", Start))

	// Start receiving updates.
	err = updater.StartPolling(b, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		panic("failed to start polling: " + err.Error())
	}
	fmt.Printf("%s has been started...\n", b.User.Username)

	// Idle, to keep updates coming in, and avoid bot stopping.
	updater.Idle()
	_ = updater.Stop()
}

func Dowork(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	_, err := msg.Copy(b, msg.Chat.Id, &gotgbot.CopyMessageOpts{Caption: ""})
	if err != nil {
		return ext.EndGroups
	}
	_, _ = msg.Delete(b, nil)
	return ext.EndGroups
}

func Start(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	_, _ = msg.Reply(b, "I'm alive, just add me in a channel with delete and post message permission to test!", nil)
	return ext.EndGroups
}
