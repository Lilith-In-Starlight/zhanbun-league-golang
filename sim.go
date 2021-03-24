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

/* The modifiers, the lineup and the rotation are all stored in json format. */

var discord discordgo.Session // Variable for the bot's session

type Player struct {
    UUID string
    Name string
    Team string
    Batting float
    Pitching float
    Defense float
    Modifiers map
    Blood string
    Rh string
}
type Team struct {
    UUID string
    Name string
    Description string
    Icon float
    Lineup float
    Rotation float
    Modifiers map
    AvgDef string
    CurrentPitcher string
}
type Game struct {
    Home string
    Away string
    BatterHome int
    BatterAway int
    Weather string
    Inning int
    Top bool
    RunsHome int
    RunsAway int
    Bases [3]string
}

func main(){
    // Load the .env file, this has to be discarded for heroku releases
    envs, err := godotenv.Read(".env")
    CheckError(err)

    // Set up the bot using the bot key that's found in the environment variable
    discord, err := discordgo.New("Bot " + envs["BOT_KEY"])
    CheckError(err)


    players_file_json, err := ioutil.ReadFile("players.txt")
    CheckError(err)

    var players_file map[string]interface{}
    json.Unmarshal([]byte(players_file_json), &players_file)

    db, err := sql.Open("postgres", envs["DATABASE_URL"])
    CheckError(err)

    // Set up the tables for the players, the teams and the standings
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS players(
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

    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS teams(
        uuid TEXT PRIMARY KEY,
        name TEXT,
        description TEXT,
        icon TEXT,
        lineup TEXT,
        rotation TEXT,
        modifiers TEXT,
        avg_def FLOAT8,
        current_pitcher INTEGER
    )`)

    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS fans(
        id TEXT PRIMARY KEY,
        favorite_team TEXT,
        coins INTEGER,
        votes INTEGER,
        snoil INTEGER,
        idol TEXT,
        strikeout_amulets INTEGER,
        run_amulet INTEGER,
        home_run_amulet INTEGER,
    )`)

    // Setup the pools
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS name_pool(
        name TEXT UNIQUE
    )`)
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS food_pool(
        food TEXT UNIQUE
    )`)
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS drink_pool(
        drink TEXT UNIQUE
    )`)
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS blood_pool(
        blood TEXT UNIQUE
    )`)
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS rh_pool(
        rh TEXT UNIQUE
    )`)
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS ritual_pool(
        ritual TEXT UNIQUE
    )`)


    // Add handler for when people send messages and open the bot
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

func createTeam(name *string, description *string, icon *string) {

}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
    }
    // If the channel the message was sent is the commands channel
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

// Crash the program if it finds an error
func CheckError(err error) {
    if err != nil {
        log.Fatal(err)
    }
}
