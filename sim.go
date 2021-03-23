package main

import (
  "log"
  "fmt"
  "os"
	"os/signal"
	"syscall"
  "io/ioutil"
  "encoding/json"
  "database/sql"

  _ "github.com/lib/pq"
  "github.com/bwmarrin/discordgo"
  "github.com/joho/godotenv"
  //"github.com/google/uuid"
)

var attacking_league [10]string
var defending_league [10]string

func main(){
  envs, err := godotenv.Read(".env")
  CheckError(err)

  discord, err := discordgo.New("Bot " + envs["BOT_KEY"])
  CheckError(err)

  team_file_json, err := ioutil.ReadFile("teams.txt")
  CheckError(err)

  var team_file map[string]interface{}
  json.Unmarshal([]byte(team_file_json), &team_file)

  db, err := sql.Open("postgres", envs["DATABASE_URL"])
  CheckError(err)

  _, err = db.Exec(`CREATE TABLE players(
    uuid TEXT PRIMARY KEY,
    name TEXT,
    team TEXT,
    batting FLOAT8,
    pitching FLOAT8,
    defense FLOAT8,
    blaserunning FLOAT8,
    modifiers TEXT,
    blood TEXT,
    rh TEXT,
    drink TEXT,
    food TEXT,
    ritual TEXT
    )`)
  for k := range team_file {
    exec := "INSERT INTO players values ($1, $2, $3, %4, %5, %6, %7, %8, %9, %10, %11, %12)"
    db.Exec(exec, k["uuid"], k["name"], k["team"], k["batting"], k["pitching"], k["defense"], k["blaserunning"], k["modifiers"], k["blood"], k["rh"], k["drink"], k["food"], k["ritual"])

  discord.AddHandler(messageCreate)
  discord.Identify.Intents = discordgo.IntentsGuildMessages
  err = discord.Open()
  CheckError(err)

  // Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	fmt.Println("Closing bot.")
  db.Close()
	discord.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
  if m.ChannelID == "823421429601140756" {
    switch m.Content {
    case "&help":
      s.ChannelMessageSend(m.ChannelID, `**How to use the Zhanbun League Blasebot**
        **$help:** Sends this message.
        **$st:** Shows all teams.`)
    case "&st":
      // var content string = ""
    }
  }
}

func CheckError(err error) {
  if err != nil {
    log.Fatal(err)
  }
}
