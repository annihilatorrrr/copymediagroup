package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

var MediaGroups = make(map[string][]Mdata)

type Pics struct {
	Media   []gotgbot.PhotoSize
	Caption string
	ParseM  string
}

type Mdata struct {
	photo Pics
	video Video
	aud   Audio
	doc   Docs
}

type Video struct {
	Media   *gotgbot.Video
	Caption string
	ParseM  string
}

type Docs struct {
	Media   *gotgbot.Document
	Caption string
	ParseM  string
}

type Audio struct {
	Media   *gotgbot.Audio
	Caption string
	ParseM  string
}

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
	if msg.MediaGroupId != "" {
		return ext.EndGroups
	}
	_, isit := MediaGroups[msg.MediaGroupId]
	log.Println(msg.MediaGroupId)
	if !isit {
		if msg.Photo != nil {
			MediaGroups[msg.MediaGroupId] = append(MediaGroups[msg.MediaGroupId], Mdata{
				photo: Pics{Media: msg.Photo, Caption: msg.OriginalCaptionHTML(), ParseM: "html"},
			})
		}
		if msg.Document != nil {
			MediaGroups[msg.MediaGroupId] = append(MediaGroups[msg.MediaGroupId], Mdata{
				doc: Docs{Media: msg.Document, Caption: msg.OriginalCaptionHTML(), ParseM: "html"},
			})
		}
		if msg.Video != nil {
			MediaGroups[msg.MediaGroupId] = append(MediaGroups[msg.MediaGroupId], Mdata{
				video: Video{Media: msg.Video, Caption: msg.OriginalCaptionHTML(), ParseM: "html"},
			})
		}
		if msg.Audio != nil {
			MediaGroups[msg.MediaGroupId] = append(MediaGroups[msg.MediaGroupId], Mdata{
				aud: Audio{Media: msg.Audio, Caption: msg.OriginalCaptionHTML(), ParseM: "html"},
			})
		}
	}
	_, _ = msg.Delete(b, nil)
	return ext.EndGroups
}

func Start(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	_, _ = msg.Reply(b, "I'm alive, just add me in a channel with delete and post message permission to test!", nil)
	/* args := ctx.Args()[1:]
	_, isit := MediaGroups[args[0]]
	if isit {
		_, _ = b.SendMediaGroup(b, data, nil)
	} */
	return ext.EndGroups
}
