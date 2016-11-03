package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aoisensi/go-discordapp/discord"
)

var channel = discord.Snowflake(os.Getenv("DOTAQR_DISCORD_CHANNEL_ID"))

var token = os.Getenv("DOTAQR_DISCORD_TOKEN")

var gw *discord.Gateway

var api *discord.Client

var (
	lastestTwitch int
)

func main() {
	api = discord.NewClient(discord.NewBotClient(token))
	gw, _ = discord.NewGateway()
	go func() {
		for {
			err := gw.Start(token)
			log.Println(err)
			log.Println("Gateway closed.")
		}
	}()

	// init
	twitches, err := getTwitches()
	if err != nil {
		log.Fatalln(err)
	}
	for _, twitch := range twitches.Streams {
		if lastestTwitch < twitch.ID {
			lastestTwitch = twitch.ID
		}
	}

	api.Channel(channel).CreateMessage(":robot: Launched. Beep boop...")

	// main loop
	for {
		time.Sleep(time.Minute) // wait a min
		loopTwitch()
		log.Println("Looped.")
	}
}

func loopTwitch() {
	twitches, err := getTwitches()
	if err != nil {
		log.Println(err)
		return
	}
	for _, twitch := range twitches.Streams {
		if lastestTwitch >= twitch.ID {
			continue
		}
		msg, err := api.Channel(channel).CreateMessage(twitch.Channel.URL)
		if err != nil {
			log.Println(err)
		}
		lastestTwitch = twitch.ID
		log.Printf("Posted message %v.", msg.ID)
	}

}

func getTwitches() (*Twitches, error) {
	req, _ := http.NewRequest("GET", "https://api.twitch.tv/kraken/streams?game=Dota+2&broadcaster_language=ja", nil)
	req.Header.Add("Client-Id", os.Getenv("DOTAQR_TWITCH_CLIENT_ID"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var twitches Twitches
	json.NewDecoder(resp.Body).Decode(&twitches)
	return &twitches, nil
}

type Twitches struct {
	Streams []struct {
		ID        int       `json:"_id"`
		CreatedAt time.Time `json:"created_at"`
		Channel   struct {
			URL string
		}
	}
}
