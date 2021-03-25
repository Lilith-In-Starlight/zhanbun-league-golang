package main

import (
    "log"
    "fmt"
    "strings"
    "os"
    "os/signal"
    "syscall"
    _ "io/ioutil"
    "database/sql"
    "math/rand"
    "time"
    "strconv"

    _ "github.com/lib/pq"
    "github.com/bwmarrin/discordgo"
    "github.com/joho/godotenv"
    "github.com/google/uuid"
)

type Player struct {
    UUID string
    Name string
    Team string
    Batting float32
    Pitching float32
    Defense float32
    Blaserunning float32
    Modifiers map[string]int
    Blood string
    Rh string
    Drink string
    Food string
    Ritual string
}
type Team struct {
    UUID string
    Name string
    Description string
    Icon string
    Lineup []string
    Rotation []string
    Modifiers map[string]int
    AvgDef float32
    CurrentPitcher int
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

var attacking_league [10]string
var defending_league [10]string

var players map[string]Player
var teams map[string]Team

var name_pool []string
var drink_pool []string
var food_pool []string
var blood_pool []string
var rh_pool []string
var ritual_pool []string

/* The modifiers, the lineup and the rotation are all stored in json format. */

var discord discordgo.Session // Variable for the bots session

// Creates a player for the given team
func NewPlayer(team string) string {
    player := new(Player)
    player.UUID = uuid.NewString()
    player.Name = name_pool[rand.Intn(len(name_pool))] + " " + name_pool[rand.Intn(len(name_pool))]
    player.Team = team
    player.Batting = 2 + rand.Float32() * 8
    player.Pitching = 2 + rand.Float32() * 8
    player.Defense = 2 + rand.Float32() * 8
    player.Blaserunning = 2 + rand.Float32() * 8
    player.Modifiers = make(map[string]int)
    player.Blood = blood_pool[rand.Intn(len(blood_pool))]
    player.Rh = rh_pool[rand.Intn(len(rh_pool))]
    player.Drink = drink_pool[rand.Intn(len(drink_pool))]
    player.Food = food_pool[rand.Intn(len(food_pool))]
    player.Ritual = ritual_pool[rand.Intn(len(ritual_pool))]
    players[player.UUID] = *player
    return player.UUID
}

// Creates a specific player with given data
func PlayerWithData(team string, uuid string, name string, batting float32, pitching float32, defense float32, blaserunning float32, modifiers map[string]int, blood string, rh string, drink string, food string, ritual string) string {
    player := new(Player)
    player.UUID = uuid
    player.Name = name
    player.Team = team
    player.Batting = batting
    player.Pitching = pitching
    player.Defense = defense
    player.Blaserunning = blaserunning
    player.Modifiers = modifiers // TODO: Parse :; to map
    player.Blood = blood
    player.Rh = rh
    player.Drink = drink
    player.Food = food
    player.Ritual = ritual
    players[player.UUID] = *player
    return player.UUID
}

// Generates a specified amount of players for a specified team
func GeneratePlayers(team string, amount int) []string {
    var ret []string
    for i := 0; i < amount; i++ {
        ret = append(ret, NewPlayer(team))
    }
    return ret
}

// Creates a team with the given name, slogan and icon
func NewTeam(name string, description string, icon string) string {
    team := new(Team)
    team.UUID = uuid.NewString()
    team.Name = name
    team.Description = description
    team.Icon = icon
    team.Lineup = GeneratePlayers(team.UUID, 10)
    team.Rotation = GeneratePlayers(team.UUID, 2)
    team.Modifiers = make(map[string]int)

    // Players have a 20% chance of being ghosts
    if rand.Float32() < 0.2 {
        team.Modifiers["intangible"] = -1
    }
    team.AvgDef = 0
    for k := range(team.Lineup) {
        team.AvgDef += players[team.Lineup[k]].Defense
    }
    team.AvgDef /= float32(len(team.Lineup))
    team.CurrentPitcher = 0
    teams[team.UUID] = *team
    return team.UUID
}

// Creates a team with the speficied data
func TeamWithData(name string, description string, icon string, uuid string, lineup []string, rotation []string, modifiers map[string]int, currentPitcher int) string {
    team := new(Team)
    team.UUID = uuid
    team.Name = name
    team.Description = description
    team.Icon = icon
    team.Lineup = lineup
    team.Rotation = rotation
    team.Modifiers = modifiers
    team.CurrentPitcher = currentPitcher
    teams[team.UUID] = *team
    return team.UUID
}

func main(){
    players = make(map[string]Player)
    teams = make(map[string]Team)
    // Make sure the RNG is random
    rand.Seed(time.Now().Unix())
    // Load the .env file, this has to be discarded for heroku releases
    envs, err := godotenv.Read(".env")
    CheckError(err)

    // Set up the bot using the bot key thats found in the environment variable
    discord, err := discordgo.New("Bot " + envs["BOT_KEY"])
    CheckError(err)

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

    // Get the pools
    rows, err := db.Query(`SELECT * FROM name_pool`)
    CheckError(err)
    for rows.Next() {
        var get string
        err = rows.Scan(&get)
        CheckError(err)
        get = strings.Replace(get, "\r", "", -1)
        name_pool = append(name_pool, get)
    }
    rows.Close()

    rows, err = db.Query(`SELECT * FROM drink_pool`)
    CheckError(err)
    for rows.Next() {
        var get string
        err = rows.Scan(&get)
        CheckError(err)
        get = strings.Replace(get, "\r", "", -1)
        drink_pool = append(drink_pool, get)
    }
    rows.Close()
    rows, err = db.Query(`SELECT * FROM food_pool`)
    CheckError(err)
    for rows.Next() {
        var get string
        err = rows.Scan(&get)
        CheckError(err)
        get = strings.Replace(get, "\r", "", -1)
        food_pool = append(food_pool, get)
    }
    rows.Close()
    rows, err = db.Query(`SELECT * FROM blood_pool`)
    CheckError(err)
    for rows.Next() {
        var get string
        err = rows.Scan(&get)
        CheckError(err)
        get = strings.Replace(get, "\r", "", -1)
        blood_pool = append(blood_pool, get)
    }
    rows.Close()
    rows, err = db.Query(`SELECT * FROM rh_pool`)
    CheckError(err)
    for rows.Next() {
        var get string
        err = rows.Scan(&get)
        CheckError(err)
        get = strings.Replace(get, "\r", "", -1)
        rh_pool = append(rh_pool, get)
    }
    rows.Close()

    rows, err = db.Query(`SELECT * FROM ritual_pool`)
    CheckError(err)
    for rows.Next() {
        var get string
        err = rows.Scan(&get)
        CheckError(err)
        get = strings.Replace(get, "\r", "", -1)
        rh_pool = append(ritual_pool, get)
    }
    rows.Close()

    // Load players from database
    rows, err = db.Query(`SELECT * FROM players`)
    CheckError(err)
    for rows.Next() {
        var uuid, name, team, modifiers, blood, rh, drink, food, ritual string
        var batting, pitching, defense, blaserunning float32
        err := rows.Scan(&uuid, &name, &team, &batting, &pitching, &defense, &blaserunning, &modifiers, &blood, &rh, &drink, &food, &ritual)
        CheckError(err)
        PlayerWithData(team, uuid, name, batting, pitching, defense, blaserunning, StringMap(modifiers), blood, rh, drink, food, ritual)
    }
    rows.Close()

    rows, err = db.Query(`SELECT * FROM teams`)
    CheckError(err)
    for rows.Next() {
        var uuid, name, description, icon, lineup, rotation, modifiers string
        var AvgDef float32
        var currentPitcher int
        err := rows.Scan(&uuid, &name, &description, &icon, &lineup, &rotation, &modifiers, &AvgDef, &CurrentPitcher)
        CheckError(err)
        TeamWithData(name, description, icon, uuid, lineup, rotation, StringMap(modifiers), AvgDef, CurrentPitcher)
    }
    rows.Close()




    // Creates the teams
    /*NewTeam("Pacificside Transcendentals", "To Infinity And Beyond", "🎭")
    NewTeam("Atlanticside Vivifiers", "Dont panic~", "😆")
    NewTeam("America Monster Trucks", "Two Trucks Playing Ball!", "🚚")
    NewTeam("Nyan City Popcats", "Nya~!", "🐱")
    NewTeam("Wisconsin Showstoppers", "Go Fetch That, W!", "🐶")
    NewTeam("Night Vales", "Welcome.", "🕵️‍♀️")
    NewTeam("Candyland Spices", "Salt is Our Sugar!", "🍭")
    NewTeam("Miletus Philosophers", "Literally What The Fuck", "🤔")
    NewTeam("Roanoke Rumble", "Literally What The Fuck", "🌟")
    NewTeam("Babel Librarians", "Shhhhhhutout!", "📚")

    NewTeam("Xnopyt Linguists", "hrrkrkrkrwpfrbrbrbrlablblblblblblwhithooap", "📜")
    NewTeam("Hoyler Erdshnekn Fun Der Yam", "Sea Slugs Give You Colorful Hugs!", "🐌")
    NewTeam("Attlerock Beans", "Powered By Digestion!", "🥫")
    NewTeam("Myspace Surfers", "Radical Play, Dude!", "🌊")
    NewTeam("New Zealand Hobbits", "Short? Stop!", "🧒")
    NewTeam("Underground Skeeltons", "Do You Wanna Have A Bad Time", "💀")
    NewTeam("Otherside Eggs", "Still Cis, Though!", "⚗️")
    NewTeam("Thaumic Paracelsii", "Rubedo Through Sports!", "🥚")
    NewTeam("Ampersandia Fallen", "Strong, United, Playin Til We Fall!", "⚰️")
    NewTeam("Lojbanistan Esperantist", "We Want A Logical Auxlang!", "📗")

    !!!ONLY USE THIS ON EMERGENCIES!!!
    _, err = db.Exec("DELETE FROM teams")
    _, err = db.Exec("DELETE FROM players")
    CheckError(err)

    for k := range teams {
        team := teams[k]
        command := `INSERT INTO teams VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
        _, err := db.Exec(command, team.UUID, strings.Replace(team.Name, "", "", -1), strings.Replace(team.Description, "", "", -1), team.Icon, SliceString(team.Lineup), SliceString(team.Rotation), MapString(team.Modifiers), team.AvgDef, team.CurrentPitcher)
        CheckError(err)
    }
    for k := range players {
        player := players[k]
        command := `INSERT INTO players VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
        _, err := db.Exec(command, player.UUID, strings.Replace(player.Name, "", "", -1), strings.Replace(player.Team, "", "", -1), player.Batting, player.Pitching, player.Defense, player.Blaserunning, MapString(player.Modifiers), player.Blood, player.Rh, player.Drink, player.Food, player.Ritual)
        CheckError(err)
    }*/

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

// Creates a string ins :; format from a slice of strings
func SliceString(slice []string) string {
    ret := ""
    for k := range slice {
        ret += slice[k] + ";;"
    }
    if len(ret) >= 2 {
        return ret[:len(ret)-2]
    } else {
        return ""
    }
}

// Creates a string ins :; format from a map
func MapString(m map[string]int) string {
    ret := ""
    for k := range m {
        ret += string(k) + "::" + strconv.Itoa(m[k]) + ";;"
    }
    if len(ret) >= 2 {
        return ret[:len(ret)-2]
    } else {
        return ""
    }
}

// Creates a map from a string in :; format
func StringMap(str string) map[string]int {
    ret := make(map[string]int)
    if str != "" {
        split := strings.Split(str, ";;")
        for k := range split {
            splot := strings.Split(split[k], "::")
            ret[splot[0]], _ = strconv.Atoi(splot[1])
        }
    }
    return ret
}
