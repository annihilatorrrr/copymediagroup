package main

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

var (
	MediaGroups = make(map[string][]gotgbot.InputMedia)
	ONHOLD      = make(map[int64]structtt)
)

type structtt struct {
	Id string
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
	updater := ext.NewUpdater(nil)
	dispatcher := updater.Dispatcher

	// Add echo handler to reply to all text messages.
	dispatcher.AddHandler(handlers.NewCommand("start", Start))
	dispatcher.AddHandler(handlers.NewMessage(message.ChatType("private"), Dowork))

	// Start receiving updates.
	err = updater.StartPolling(b, &ext.PollingOpts{
		DropPendingUpdates: false,
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
	if msg.MediaGroupId == "" {
		return ext.EndGroups
	}
	log.Println(msg.MediaGroupId)
	if msg.Photo != nil {
		MediaGroups[msg.MediaGroupId] = append(MediaGroups[msg.MediaGroupId], gotgbot.InputMediaPhoto{
			Media: msg.Photo, Caption: msg.OriginalCaptionHTML(), ParseMode: "html", CaptionEntities: msg.CaptionEntities,
		})
	}
	if msg.Document != nil {
		MediaGroups[msg.MediaGroupId] = append(MediaGroups[msg.MediaGroupId], gotgbot.InputMediaDocument{
			Media: msg.Document, Caption: msg.OriginalCaptionHTML(), ParseMode: "html",
		})
	}
	if msg.Video != nil {
		MediaGroups[msg.MediaGroupId] = append(MediaGroups[msg.MediaGroupId], gotgbot.InputMediaVideo{
			Media: msg.Video, Caption: msg.OriginalCaptionHTML(), ParseMode: "html",
		})
	}
	if msg.Audio != nil {
		MediaGroups[msg.MediaGroupId] = append(MediaGroups[msg.MediaGroupId], gotgbot.InputMediaAudio{
			Media: msg.Audio, Caption: msg.OriginalCaptionHTML(), ParseMode: "html",
		})
	}
	_, _ = msg.Delete(b, nil)
	if verm, isit := ONHOLD[msg.Chat.Id]; isit && msg.MediaGroupId == verm.Id {
		return ext.EndGroups
	}
	ONHOLD[msg.Chat.Id] = structtt{
		Id: msg.MediaGroupId,
	}
	time.Sleep(2 * time.Second)
	data, isit := MediaGroups[msg.MediaGroupId]
	if !isit {
		return ext.EndGroups
	}
	_, err = b.SendMediaGroup(msg.Chat.Id, data, nil)
	if err != nil {
		log.Println(err.Error())
	}
	return ext.EndGroups
}

func Start(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	_, _ = msg.Reply(b, "I'm alive, just send me a media group to test!", nil)
	return ext.EndGroups
}
