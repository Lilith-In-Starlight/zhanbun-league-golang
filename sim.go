package main

import (
  "fmt"
  "os"
	"os/signal"
	"syscall"

  "github.com/bwmarrin/discordgo"
)

func main(){
  fmt.Println("Initiated program")
  discord, err := discordgo.New("Bot " + "ODIzOTgyNDg3MjcyMjkyMzcy.YFovfQ.JTEBZS8UlwXHeY9gDEiJJmgm3Ks")
  if err != nil {
    fmt.Println("Error creating Discord session: ", err)
    return
  }
  discord.AddHandler(messageCreate)
  discord.Identify.Intents = discordgo.IntentsGuildMessages
  err = discord.Open()
  if err != nil {
    fmt.Println("Error opening Discord session: ", err)
  }

  // Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	fmt.Println("Closing bot.")
	discord.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
  if m.ChannelID == "823421429601140756" {
    switch m.content {
    case "&help":
      s.ChannelMessageSend(m.ChannelID, `**How to use the Zhanbun League Blasebot**
        **$help:** Sends this message.
        **$st:** Shows all teams.`)
    }
  }
}
