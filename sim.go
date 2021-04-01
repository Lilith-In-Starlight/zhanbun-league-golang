package main

import (
    "log"
    "fmt"
    "strings"
    "os"
    "os/signal"
    "syscall"
    _ "encoding/json"
    _ "io/ioutil"
    "database/sql"
    "math/rand"
    "math"
    "time"
    "strconv"

    _ "github.com/lib/pq"
    "github.com/bwmarrin/discordgo"
    _ "github.com/joho/godotenv"
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
    MaxInnings int
    Top bool
    RunsHome int
    RunsAway int
    Outs int
    Strikes int
    Balls int
    Bases [3]string
    MessageId string
    InningState string
    Announcements []string
    AnnouncementStates []Game
    ChangeBatter bool
    BattingTeam Team
    PitchingTeam Team
    Batter Player
    Pitcher Player
    LastMessage string
    betsHome map[string]int
    betsAway map[string]int
    gameEnded bool
}
type Fan struct {
    Id string
    Team string
    Coins int
    Votes int
    Shop map[string]int
    Stan string
}

const seasonNumber = 1

var day int = 1
var tape int = 1

const CommandChannelId = "822210011882061863"
const GamesChannelId = "822210201574703114"

var CurrentGamesId string
var CurrentGamesId2 string
var CurrentGamesId3 string

var cool_league [10]string
var fun_league [10]string

var players map[string]*Player
var teams map[string]*Team
var fans map[string]*Fan

var name_pool []string
var drink_pool []string
var food_pool []string
var blood_pool []string
var rh_pool []string
var ritual_pool []string

var upcoming []*Game
var games []*Game

var wins map[string]int
var losses map[string]int

var validuuids []string

var election map[string]int
var bless1 map[string]int
var bless2 map[string]int
var bless3 map[string]int


var mods []string = []string{"ash_twin", "ember_twin", "still_alive", "haunted"}

var modNames map[string]string = map[string]string {
    "ash_twin" : "Ash Twin",
    "ember_twin" : "Ember Twin",
    "still_alive" : "Still Alive",
    "haunted" : "Haunted",
    "quantum" : "Quantum",
}
var modDescs map[string]string = map[string]string {
    "ash_twin" : "This player is an Ash Twin.",
    "ember_twin" : "This player might be paired with an Ash Twin.",
    "still_alive" : "When this player's dying they'll be Still Alive.",
    "haunted" : "This player sees players who aren't there.",
    "quantum" : "This player can be in multiple states at the same time.",
}
var modIcons map[string]string = map[string]string{
    "ash_twin" : "🌫️",
    "ember_twin" : "💫",
    "still_alive" : "🐱",
    "haunted" : "👥",
    "quantum" : "⚛️",
}

var weathers []string = []string{"ash", "ember", "feedback"}

var weatherNames map[string]string = map[string]string {
    "ash" : "Ashes",
    "ember" : "Embers",
    "feedback" : "Feedback",
}
var weatherDescs map[string]string = map[string]string {
    "ash" : "Pitchers may be haunted, and players may be Paired",
    "ember" : "Batters may become ember twins.",
    "feedback" : "Players may receive feedback.",
}
var weatherIcons map[string]string = map[string]string{
    "ash" : "🌫️",
    "ember" : "💫",
    "feedback" : "🎙️",
}

var field []string

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
    players[player.UUID] = player
    return player.UUID
}

// Creates a specific player with given data, returns the UUID
func PlayerWithData(team string, uuid string, name string, batting float32, pitching float32, defense float32, blaserunning float32, modifiers map[string]int, blood string, rh string, drink string, food string, ritual string) string {
    player := new(Player)
    player.UUID = uuid
    player.Name = name
    player.Team = team
    player.Batting = batting
    player.Pitching = pitching
    player.Defense = defense
    player.Blaserunning = blaserunning
    player.Modifiers = modifiers
    player.Blood = blood
    player.Rh = rh
    player.Drink = drink
    player.Food = food
    player.Ritual = ritual
    players[player.UUID] = player
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

// Creates a team with the given name, slogan and icon, returns UUID
func NewTeam(name string, description string, icon string) string {
    team := new(Team)
    team.UUID = uuid.NewString()
    team.Name = name
    team.Description = description
    team.Icon = icon
    team.Lineup = GeneratePlayers(team.UUID, 10)
    team.Rotation = GeneratePlayers(team.UUID, 2)
    team.Modifiers = make(map[string]int)
    team.AvgDef = 0
    for k := range(team.Lineup) {
        team.AvgDef += players[team.Lineup[k]].Defense
    }
    team.AvgDef /= float32(len(team.Lineup))
    team.CurrentPitcher = 0
    teams[team.UUID] = team
    return team.UUID
}

// Creates a team with the speficied data, returns the UUID
func TeamWithData(name string, description string, icon string, uuid string, lineup []string, rotation []string, modifiers map[string]int, avgDef float32, currentPitcher int) string {
    team := new(Team)
    team.UUID = uuid
    team.Name = name
    team.Description = description
    team.Icon = icon
    team.Lineup = lineup
    team.Rotation = rotation
    team.Modifiers = modifiers
    team.CurrentPitcher = currentPitcher
    team.AvgDef = avgDef
    teams[team.UUID] = team
    return team.UUID
}

// Creates a new game and puts it in upcoming
func NewGame(home string, away string, innings int) {
    game := new(Game)
    game.Home = home
    game.Away = away
    game.BatterHome, game.BatterAway = 0, 0
    game.Weather = weathers[rand.Intn(len(weathers))]
    game.Inning = 1
    game.MaxInnings = innings
    game.Top = true
    game.RunsAway, game.RunsHome = 0, 0
    game.MaxInnings = innings
    game.InningState = "starting"
    game.Announcements = []string{}
    game.AnnouncementStates = []Game{}
    game.ChangeBatter = true
    game.betsAway = make(map[string]int)
    game.betsHome = make(map[string]int)
    game.gameEnded = false
    upcoming = append(upcoming, game)
}

func NewFan (id string, team string, coins int, votes int, shop map[string]int, stan string) string {
    fan := new(Fan)
    fan.Id = id
    fan.Team = team
    fan.Coins = coins
    fan.Votes = votes
    fan.Shop = shop
    fan.Stan = stan
    fans[fan.Id] = fan
    return fan.Id
}

func main(){
    players = make(map[string]*Player)
    teams = make(map[string]*Team)
    fans = make(map[string]*Fan)
    wins = make(map[string]int)
    losses = make(map[string]int)
    election = make(map[string]int)
    bless1 = make(map[string]int)
    bless2 = make(map[string]int)
    bless3 = make(map[string]int)
    // Make sure the RNG is random
    rand.Seed(time.Now().Unix())
    // Load the .env file, this has to be discarded for heroku releases
    // envs, err := godotenv.Read(".env")
    // CheckError(err)

    // Set up the bot using the bot key thats found in the environment variable
    // discord, err := discordgo.New("Bot " + envs["BOT_KEY"])
    discord, err := discordgo.New("Bot " + os.Getenv("BOT_KEY"))
    CheckError(err)

    // db, err := sql.Open("postgres", envs["DATABASE_URL"])
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
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
        shop TEXT,
        stan TEXT
    )`)
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS seasons(
        number INTEGER PRIMARY KEY,
        day INTEGER,
        tape INTEGER,
        election TEXT,
        wins TEXT,
        losses TEXT,
        b1 TEXT,
        b2 TEXT,
        b3 TEXT
    )`)

    // Setup the pools
    //db.Exec("DROP TABLE leagues")
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS leagues(
        cool TEXT UNIQUE,
        fun TEXT UNIQUE
    )`)
    CheckError(err)
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
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS dead_people(
        uuid TEXT
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
        ritual_pool = append(ritual_pool, get)
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
        for strings.Contains(name, "''") {
            name = strings.Replace(name, "''", "'", -1)
        }
        modifieds := StringMap(modifiers)
        if name == "Normal Cowboy" {
            modifieds["still_alive"] = -1
        }
        PlayerWithData(team, uuid, name, batting, pitching, defense, blaserunning, modifieds, blood, rh, drink, food, ritual)
    }
    rows.Close()

    rows, err = db.Query(`SELECT * FROM teams`)
    CheckError(err)
    for rows.Next() {
        var uuid, name, description, icon, lineup, rotation, modifiers string
        var AvgDef float32
        var currentPitcher int
        err := rows.Scan(&uuid, &name, &description, &icon, &lineup, &rotation, &modifiers, &AvgDef, &currentPitcher)
        modi := StringMap(modifiers)
        CheckError(err)
        if name == "Otherside Eggs" {
            icon = "🥚"
        } else if name == "Thaumic Paracelsii" {
            icon = "⚗️"
        }
        TeamWithData(name, description, icon, uuid, StringSlice(lineup), StringSlice(rotation), modi, AvgDef, currentPitcher)
    }

    rows.Close()

    rows, err = db.Query(`SELECT * FROM dead_people`)
    CheckError(err)
    for rows.Next() {
        var get string
        err = rows.Scan(&get)
        CheckError(err)
        get = strings.Replace(get, "\r", "", -1)
        field = append(field, get)
    }
    rows.Close()

    rows, err = db.Query(`SELECT * FROM seasons`)
    CheckError(err)
    for rows.Next() {
        var s, d, t int
        var e, l, w, b1, b2, b3 string
        err = rows.Scan(&s, &d, &t, &e, &w, &l, &b1, &b2, &b3)
        CheckError(err)
        day, tape, election, losses, wins, bless1, bless2, bless3 = d, t, StringMap(e), StringMap(l), StringMap(w), StringMap(b1), StringMap(b2), StringMap(b3)
    }
    rows.Close()

    // This was used to generate the leagues. Only use it if the database is reset
    /*i := 0
    j := 0
    for k := range teams {
        if i % 2 == 0 {
            fun_league[j] = string(k)
        } else {
            cool_league[j] = string(k)
            fmt.Println(fun_league[j], cool_league[j])
            _, err = db.Exec(`INSERT INTO leagues (cool, fun) VALUES ($1, $2)`, cool_league[j], fun_league[j])
            CheckError(err)
            j += 1
        }
        i += 1
    }*/


    rows, err = db.Query(`SELECT * FROM leagues`)
    CheckError(err)
    i := 0
    for rows.Next() {
        var coole, fune string
        err := rows.Scan(&coole, &fune)
        CheckError(err)
        cool_league[i] = coole
        fun_league[i] = fune
        i += 1
    }

    rows.Close()


    /*text, err := json.Marshal(players)
    err = ioutil.WriteFile("players.json", text, 0644)

    text, err = json.Marshal(teams)
    err = ioutil.WriteFile("teams.json", text, 0644)

    text, err = json.Marshal(fans)
    err = ioutil.WriteFile("fans.json", text, 0644)

    text, err = json.Marshal(cool_league)
    err = ioutil.WriteFile("cool.json", text, 0644)

    text, err = json.Marshal(fun_league)
    err = ioutil.WriteFile("fun.json", text, 0644)*/


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


    // Open the goroutine that handles upcoming and current games
    go HandleGames(discord, db)
    time.Sleep(0)

    // Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session and the database.
	fmt.Println("Closing bot.")
    db.Close()
    discord.Close()
}

func HandleGames(session *discordgo.Session, db *sql.DB) {
    for true {
        // If there's upcoming games, but no games are currently being played
        if len(upcoming) > 0 && len(games) == 0 {
            // The three messages for the games
            msg, _ := session.ChannelMessageSend(GamesChannelId, "**Preparing Games**")
            CurrentGamesId = msg.ID
            msg, _ = session.ChannelMessageSend(GamesChannelId, "**Preparing Games**")
            CurrentGamesId2 = msg.ID
            msg, _ = session.ChannelMessageSend(GamesChannelId, "**Preparing Games**")
            CurrentGamesId3 = msg.ID

            // Send the upcoming games to the currently played games
            for k := range upcoming {
                games = append(games, upcoming[k])
            }
            upcoming = make([]*Game, 0)

        }
        if len(upcoming) == 0 { // If there are no upcoming games
            if tape == 0 || tape >= 10 {
                tape = 1
            }
            var touched = [10]int{-1,-1,-1,-1,-1,-1,-1,-1,-1,-1}
            for i := 0; i<10;i++ {
                j := i + tape
                if j >= 10 {
                    j -= 10
                }
                if touched[i] == -1 && touched[j] == -1 {
                    NewGame(fun_league[i], cool_league[j], 9)
                }
            }
            time.Sleep(10 * time.Second)
        }

        /* The games have to be divided in three messages
        To avoid running up against the character limit */
        if len(games) > 0 {
            if len(games) > 6 {
                HandlePlays(session, CurrentGamesId, 0, 3)
                HandlePlays(session, CurrentGamesId2, 3, 6)
                HandlePlays(session, CurrentGamesId3, 6, len(games))
            } else if len(games) > 3 {
                HandlePlays(session, CurrentGamesId, 0, 3)
                HandlePlays(session, CurrentGamesId2, 3, len(games))
            } else if len(games) > 0 {
                HandlePlays(session, CurrentGamesId, 0, len(games))
            }

            allEnded := true
            for _, j := range games {
                if !j.gameEnded {
                    allEnded = false
                }
            }
            if allEnded {
                games = make([]*Game, 0)
                day += 1
                if day < 91 && day % 2 == 0 {
                    tape += 1
                } else if day >= 91 {
                    tape += 1
                }
                updateDatabases(db)
            }

            time.Sleep(1 * time.Second + 500 * time.Millisecond)
        }

        time.Sleep(0)
    }
}

func HandlePlays (session *discordgo.Session, message string, start int, end int) {
    output := ""
    Loop:
    for k := start; k < end; k++ {
        game := games[k]
        /* Nothing can go on until the announcements are finished
        This is done to prevent the game from announcing the half-inning
        After its player steps up to bat (error from the previous iteration of the sim)
        Among other things*/
        if len(game.Announcements) == 0 {
            if game.gameEnded == false {
                switch game.InningState {
                case "inning", "starting":
                    // If it's the top of the inning, the batter is the home team
                    // And the pitcher is the away team
                    game.BattingTeam = *teams[game.Home]
                    game.PitchingTeam = *teams[game.Away]
                    game.Batter = *players[game.BattingTeam.Lineup[game.BatterHome]]
                    game.Pitcher = *players[game.PitchingTeam.Rotation[game.PitchingTeam.CurrentPitcher]]
                    batterNumber := game.BatterHome
                    if !game.Top {
                        game.BattingTeam  = *teams[game.Away]
                        game.PitchingTeam = *teams[game.Home]
                        game.Batter = *players[game.BattingTeam.Lineup[game.BatterAway]]
                        game.Pitcher = *players[game.PitchingTeam.Rotation[game.PitchingTeam.CurrentPitcher]]
                        batterNumber = game.BatterAway
                    }

                    if game.InningState == "starting" {
                        // Reset everything when the half-inning is starting and announce the half-inning
                        if game.Top {
                            game.Announcements = append(game.Announcements, "Top of " + strconv.Itoa(game.Inning))
                            game.AnnouncementStates = append(game.AnnouncementStates, *game)
                        } else {
                            game.Announcements = append(game.Announcements, "Bottom of " + strconv.Itoa(game.Inning))
                            game.AnnouncementStates = append(game.AnnouncementStates, *game)
                        }
                        game.Bases = [3]string{"", "", ""}
                        game.Outs = 0
                        game.Strikes = 0
                        game.Balls = 0

                        game.InningState = "inning"
                    }
                    // If in the last tick it was determined that the batter should change
                    if game.ChangeBatter {
                        game.Strikes = 0
                        game.Balls = 0
                        if !game.Top {
                            game.BatterAway = (game.BatterAway + 1) % len(game.BattingTeam.Lineup)
                            game.Batter = *players[game.BattingTeam.Lineup[game.BatterAway]]
                            batterNumber = game.BatterAway
                        } else {
                            game.BatterHome = (game.BatterHome + 1) % len(game.BattingTeam.Lineup)
                            game.Batter = *players[game.BattingTeam.Lineup[game.BatterHome]]
                            batterNumber = game.BatterHome
                        }
                        game.Announcements = append(game.Announcements, game.Batter.Name + " batting for the " + game.BattingTeam.Name)
                        game.AnnouncementStates = append(game.AnnouncementStates, *game)
                        game.ChangeBatter = false
                    }
                    appended := DoWeather(teams[game.BattingTeam.UUID], teams[game.PitchingTeam.UUID], game.Weather, batterNumber)
                    // Update them one last time in case the weather did something

                    game.BattingTeam = *teams[game.Home]
                    game.PitchingTeam = *teams[game.Away]
                    game.Pitcher = *players[game.PitchingTeam.Rotation[game.PitchingTeam.CurrentPitcher]]
                    game.Batter = *players[game.BattingTeam.Lineup[game.BatterHome]]
                    if !game.Top {
                        game.BattingTeam = *teams[game.Away]
                        game.PitchingTeam = *teams[game.Home]
                        game.Batter = *players[game.BattingTeam.Lineup[game.BatterAway]]
                    }

                    game.Announcements = append(game.Announcements, appended...)
                    for _ = range appended {
                        game.AnnouncementStates = append(game.AnnouncementStates, *game)
                    }
                    if game.InningState == "inning" {
                        probability := game.Batter.Batting + game.Pitcher.Pitching //The range of probabilities
                        happen := rand.Float32() * probability // What actually happens
                        // The game.Batter manages to bat
                        if happen < game.Batter.Batting {
                            var runsScored int
                            if happen < game.Batter.Batting * 0.01 { // Homer
                                if game.Batter.UUID == game.Bases[0] && game.Bases[0] == game.Bases[1] && game.Bases[1] == game.Bases[2] {
                                    game.Announcements = append(game.Announcements, game.Batter.Name + " hits a solo grand slam!?!?!?")
                                } else {
                                    if game.Bases[0] != "" && game.Bases[1] != "" && game.Bases[2] != "" {
                                        game.Announcements = append(game.Announcements, game.Batter.Name + " hits a grand slam!!!!!")
                                    } else {

                                        game.Announcements = append(game.Announcements, game.Batter.Name + " hits a home run!!!")
                                    }
                                }
                                runsScored += Advance(&game.Bases, game.Batter.UUID, -1)
                                runsScored += Advance(&game.Bases, "", 0)
                                runsScored += Advance(&game.Bases, "", 1)
                                runsScored += Advance(&game.Bases, "", 2)

                                game.Strikes = 0
                                game.Balls = 0
                            } else if happen < game.Batter.Batting * 0.025 { // Triplet
                                game.Announcements = append(game.Announcements, game.Batter.Name + " hits a triple!!!")
                                runsScored += Advance(&game.Bases, game.Batter.UUID, -1)
                                runsScored += Advance(&game.Bases, "", 0)
                                runsScored += Advance(&game.Bases, "", 1)
                                game.Strikes = 0
                                game.Balls = 0
                            } else if happen < game.Batter.Batting * 0.1 { // Twin
                                game.Announcements = append(game.Announcements, game.Batter.Name + " hits a double!!")
                                runsScored += Advance(&game.Bases, game.Batter.UUID, -1)
                                runsScored += Advance(&game.Bases, "", 0)
                                game.Strikes = 0
                                game.Balls = 0
                            } else if happen < game.Batter.Batting * 0.4 { // Singlet
                                game.Announcements = append(game.Announcements, game.Batter.Name + " hits a single!")
                                runsScored += Advance(&game.Bases, game.Batter.UUID, -1)
                                game.Strikes = 0
                                game.Balls = 0

                            } else if happen < game.Batter.Batting * 0.999 {
                                game.Announcements = append(game.Announcements, game.Batter.Name + " hits a flyout!")
                                game.Outs += 1
                                game.ChangeBatter = true
                            } else {
                                game.Announcements = append(game.Announcements, game.Batter.Name + " hits the ball so hard it goes to another game!")
                            }
                            game.AnnouncementStates = append(game.AnnouncementStates, *game)
                            if runsScored != 0 {
                                if game.Top {
                                    game.RunsHome += runsScored
                                } else {
                                    game.RunsAway += runsScored
                                }
                                if runsScored == 1 {
                                    game.Announcements = append(game.Announcements, strconv.Itoa(runsScored) + " run scored.")
                                } else {
                                    game.Announcements = append(game.Announcements, strconv.Itoa(runsScored) + " runs scored.")
                                }
                                game.AnnouncementStates = append(game.AnnouncementStates, *game)
                            }
                            if game.Batter.Modifiers["quantum"] == 0 || rand.Float64() < 0.5 {
                                game.ChangeBatter = true
                            }

                        // The game.Batter fails to bat
                        } else {
                            if happen < game.Batter.Batting + game.Pitcher.Pitching * 0.1 {
                                game.Announcements = append(game.Announcements, "Ball.")
                            } else if happen < game.Batter.Batting + game.Pitcher.Pitching * 0.25 {
                                game.Announcements = append(game.Announcements, "Strike, swinging.")
                                game.Strikes += 1
                            } else if happen < game.Batter.Batting + game.Pitcher.Pitching * 0.65 {
                                game.Announcements = append(game.Announcements, "Strike, looking.")
                                game.Strikes += 1
                            } else if happen < game.Batter.Batting + game.Pitcher.Pitching * 0.99999 {
                                game.Announcements = append(game.Announcements, "Strike, flinching.")
                                game.Strikes += 1
                            } else {
                                game.Announcements = append(game.Announcements, "Strike, knows too much.")
                                game.Strikes += 1
                            }
                            game.AnnouncementStates = append(game.AnnouncementStates, *game)
                        }

                        if game.Balls >= 4 {
                            game.Strikes = 0
                            game.Balls = 0
                            game.Announcements = append(game.Announcements, game.Batter.Name + " gets a walk.")
                            game.AnnouncementStates = append(game.AnnouncementStates, *game)
                            Advance(&game.Bases, game.Batter.UUID, -1)
                            game.ChangeBatter = true
                        }

                        if game.Strikes >= 3 {
                            game.Strikes = 0
                            game.Balls = 0
                            game.Outs += 1
                            game.Announcements = append(game.Announcements, game.Batter.Name + " strikes out.")
                            game.AnnouncementStates = append(game.AnnouncementStates, *game)
                            game.ChangeBatter = true
                        }
                        if game.Outs >= 3 {
                            if !game.Top {
                                game.Inning += 1
                                // Game is Over
                                if game.RunsAway != game.RunsHome && game.Inning > 9 {
                                    // games = append(games[:k], games[k+1:]...)
                                    game.gameEnded = true
                                    // TODO give money to the players that bet
                                    // TODO send "game ended" message
                                    if game.RunsAway > game.RunsHome {
                                        wins[game.Away] = wins[game.Away] + 1
                                        losses[game.Home] = losses[game.Home] + 1
                                        for i, j := range game.betsAway {
                                            fans[i].Coins += j * 2 + rand.Intn(7) - 1
                                        }
                                    } else {
                                        wins[game.Home] = wins[game.Home] + 1
                                        losses[game.Away] = losses[game.Away] + 1
                                        for i, j := range game.betsHome {
                                            fans[i].Coins += j * 2 + rand.Intn(7) - 1
                                        }
                                    }
                                    fmt.Println("Finished")
                                    break Loop
                                }
                                game.Announcements = append(game.Announcements, "Inning " + strconv.Itoa(game.Inning-1) + " is now an outing.")
                                game.AnnouncementStates = append(game.AnnouncementStates, *game)
                            }
                            game.Top = !game.Top
                            game.InningState = "starting"
                            game.ChangeBatter = true
                        }
                    }
                }
                output += game.LastMessage
                output += "\n----------------\n"
            } else {
                this_output := ""
                this_output += "**" + teams[game.Home].Name + " at " + teams[game.Away].Name + "**" + "\n"
                this_output += teams[game.Home].Icon + " **" + teams[game.Home].Name + "** (" + strconv.Itoa(game.RunsHome) + ")\n"
                this_output += "🩸" + strconv.Itoa(game.Inning-1) + "\n\n"
                this_output += "Game Finished! "
                if game.RunsHome > game.RunsAway {
                    this_output += teams[game.Home].Name + " wins!\n\n"
                } else {
                    this_output += teams[game.Away].Name + " wins!\n\n"
                }
                output += this_output
                output += "\n------------\n"
            }
        } else {
            // fmt.Println(game.Announcements[0])
            this_output := ""
            this_output += "**" + teams[game.Home].Name + " at " + teams[game.Away].Name + "**" + "\n"
            this_output += teams[game.Home].Icon + " **" + teams[game.Home].Name + "** (" + strconv.Itoa(game.AnnouncementStates[0].RunsHome) + ")\n"
            if game.AnnouncementStates[0].Top {
                this_output += "🔺"
            } else {
                this_output += "🔻"
            }
            this_output += strconv.Itoa(game.AnnouncementStates[0].Inning) + " [" + weatherIcons[game.AnnouncementStates[0].Weather] + " " + weatherNames[game.AnnouncementStates[0].Weather] + "]" + "\n"
            this_output += teams[game.Away].Icon + " **" + teams[game.Away].Name + "** (" + strconv.Itoa(game.AnnouncementStates[0].RunsAway) + ")\n\n"
            this_output += "🏏 " + game.AnnouncementStates[0].Batter.Name + "\n"
            this_output += "⚾ " + game.AnnouncementStates[0].Pitcher.Name + "\n\n"
            this_output += "**Outs:** " + ShowInCircles(game.AnnouncementStates[0].Outs, 3) + "\n"
            this_output += "**Strikes:** " + ShowInCircles(game.AnnouncementStates[0].Strikes, 3) + "\n"
            this_output += "**Balls:** " + ShowInCircles(game.AnnouncementStates[0].Balls, 4) + "\n\n"
            // The data for the fields must not be empty
            if game.AnnouncementStates[0].Bases[0] != "" {
                this_output += "1️⃣ " + players[game.AnnouncementStates[0].Bases[0]].Name + "\n"
            } else {
                this_output += "1️⃣ <Empty>" + "\n"
            }
            if game.AnnouncementStates[0].Bases[1] != "" {
                this_output += "2️⃣ " + players[game.AnnouncementStates[0].Bases[1]].Name + "\n"
            } else {
                this_output += "2️⃣ <Empty>" + "\n"
            }
            if game.AnnouncementStates[0].Bases[2] != "" {
                this_output += "3️⃣ " + players[game.AnnouncementStates[0].Bases[2]].Name + "\n\n"
            } else {
                this_output += "3️⃣ <Empty>" + "\n"
            }
            this_output += game.Announcements[0] + "\n\n"
            game.LastMessage = this_output
            output += this_output
            output += "\n------------\n"
            game.Announcements = append(game.Announcements[:0], game.Announcements[1:]...)
            game.AnnouncementStates = append(game.AnnouncementStates[:0], game.AnnouncementStates[1:]...)
        }
    }
    if output != "" {
        _, err := session.ChannelMessageEdit(GamesChannelId, message, output)
        CheckError(err)
        // fmt.Println(output)
    }
}

func DoWeather(bat *Team, pitch *Team, w string, batter int) []string {
    var output []string
    switch w {
    case "ash":
        if rand.Float64() < 0.009 {
            if players[pitch.Rotation[pitch.CurrentPitcher]].Modifiers["ash_twin"] == 0 {
                players[pitch.Rotation[pitch.CurrentPitcher]].Modifiers["ash_twin"] = 2
            }
            output = append(output, players[pitch.Rotation[pitch.CurrentPitcher]].Name + " is covered in ashes. They are an Ash Twin!")
        }
        if rand.Float64() < 0.009 {
            if players[bat.Lineup[batter]].Modifiers["ember_twin"] > 0 {
                for i, j := range pitch.Lineup {
                    if _, ok := players[j].Modifiers["ash_twin"]; ok && rand.Float64() < 0.7 {
                        players[bat.Lineup[batter]].Modifiers["ember_twin"] = 0
                        players[bat.Lineup[batter]].Modifiers["ash_twin"] = -1
                        players[pitch.Lineup[i]].Modifiers["ash_twin"] = 0
                        output = append(output, players[bat.Lineup[batter]].Name + " is caught up in the ashes. They pair with " + players[pitch.Lineup[i]].Name + "!")
                        output = append(output, players[pitch.Lineup[i]].Name + " steps up to bat.")
                        FeedbackPlayers(&bat.Lineup[batter], &pitch.Lineup[i])
                        break
                    }
                }
            }
        }
    case "ember":
        if rand.Float64() < 0.01 {
            if players[bat.Lineup[batter]].Modifiers["ember"] == 0 {
                players[bat.Lineup[batter]].Modifiers["ember"] = 3
            }
            output = append(output, players[bat.Lineup[batter]].Name + " is caught up in the embers. They are an Ember Twin!")
        } else if rand.Float64() < 0.0001 {
            if rand.Float64() < 0.5 {
                output = append(output, players[bat.Lineup[batter]].Name + " is caught up in the embers. They are incinerated!")
                Incinerate(&bat.Lineup[batter])
                output = append(output, "An umpire throws a body on home base.")
                output = append(output, players[bat.Lineup[batter]].Name + " steps up to bat!")
            } else {
                output = append(output, players[pitch.Rotation[pitch.CurrentPitcher]].Name + " is caught up in the embers. They are incinerated!")
                Incinerate(&pitch.Rotation[pitch.CurrentPitcher])
                output = append(output, "An umpire throws a body on the mound.")
                output = append(output, players[pitch.Rotation[pitch.CurrentPitcher]].Name + " is now pitching!")
            }
        }
    case "feedback":
        if rand.Float64() < 0.003 {
            FeedbackPlayers(&pitch.Rotation[pitch.CurrentPitcher], &bat.Lineup[batter])
            if rand.Float64() < 0.5 {
                output = append(output, players[pitch.Rotation[pitch.CurrentPitcher]].Name + " receives feedback.")
                output = append(output, players[bat.Lineup[batter]].Name + " is now a batter for the " + bat.Name + "!")
            } else {
                output = append(output, players[bat.Lineup[batter]].Name + " receives feedback.")
                output = append(output, players[pitch.Rotation[pitch.CurrentPitcher]].Name + " is now a pitcher for the " + pitch.Name + "!")
            }
        }
    }
    return output
}

func FeedbackPlayers(pointA *string, pointB *string) {
    *pointA, *pointB = *pointB, *pointA
}

func Incinerate(player *string) {
    if players[*player].Modifiers["still_alive"] == 0{
        newPlayer := NewPlayer(players[*player].Team)
        field = append(field, *player)
        *player = newPlayer
    } else {
        field = append(field, *player)
        players[*player].Modifiers["quantum"] = -1
        players[*player].Modifiers["still_alive"] = 0
    }
}

func ShowInCircles(n int, m int) string {
    output := ""
    for i := 0; i < m-1; i++ {
        if i < n {
            output += "🔶"
        } else {
            output += "🔷"
        }
    }
    return output
}

func Advance(bases *[3]string, homeBase string, from int) int{
    runsScored := 0
    if from == -1 {
        if bases[0] != "" {
            runsScored += Advance(bases, homeBase, 0)
        }
        bases[0] = homeBase
    } else {
        if from + 1 <= 2 {
            if bases[from + 1] != "" {
                runsScored += Advance(bases, homeBase, from + 1)
            }
            bases[from + 1] = bases[from]
        } else {
            runsScored += 1
        }
        bases[from] = ""
    }
    return runsScored
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
    }
    // If the channel the message was sent is the commands channel
    if m.ChannelID == CommandChannelId {
        switch m.Content {
        case "&help":
            emb := new(discordgo.MessageEmbed)
            emb.Title = "Zhanbun League Blasebot's Command List"
            emb.Color = 8651301
            AddField(emb, "&help", "Displays this list.", false)
            AddField(emb, "&st", "Shows a list of all teams.", false)
            AddField(emb, "&u", "Shows all the upcoming games.", false)
            AddField(emb, "&s", "Searches for a team or a player. &st [team name | icon] > (player name)", false)
            AddField(emb, "&b", "Shows you the store and lets you buy things. &b ([icon] > [amount to buy])", false)
            AddField(emb, "&f", "Lets you select your favorite team. &f [icon]. Requires Fairweather Flutes.", false)
            AddField(emb, "&e", "Participate in the election. &e ([icon] > [amount of votes])", false)
            /*emb.AddField("&help", "Sends this list of commands.")
            emb.AddField("&st", "Shows a list of all teams.")*/
            s.ChannelMessageSendEmbed(m.ChannelID, emb)
            /*s.ChannelMessageSend(m.ChannelID, `**How to use the Zhanbun League Blasebot**
                **$help:** Sends this message.
                **$st:** Shows all teams.`)*/
        case "&st":
            emb := new(discordgo.MessageEmbed)
            emb.Title = "ZHANBUN LEAGUE BLASEBALL"
            emb.Color = 8651301
            AddField(emb, "±± THE COOL LEAGUE ±±", "-+-+-+-+-", false)
            for k := range cool_league {
                AddField(emb, teams[cool_league[k]].Icon + " " + teams[cool_league[k]].Name, teams[cool_league[k]].Description, false)
            }
            s.ChannelMessageSendEmbed(m.ChannelID, emb)

            emb = new(discordgo.MessageEmbed)
            emb.Color = 8651301
            AddField(emb, "±± THE FUN LEAGUE ±±", "+-+-+-+-+", false)
            for k := range fun_league {
                AddField(emb, teams[fun_league[k]].Icon + " " + teams[fun_league[k]].Name, teams[fun_league[k]].Description, false)
            }
            s.ChannelMessageSendEmbed(m.ChannelID, emb)

        case "&u":
            emb := new(discordgo.MessageEmbed)
            emb.Color = 8651301
            text := ""
            fmt.Println("Uwu?")
            for _, k := range upcoming {
                text += teams[k.Home].Icon + " " + teams[k.Home].Name + " - " + teams[k.Away].Name + " " + teams[k.Away].Icon + "\n"
            }
            if text == "" {
                text = "No games prepared for now."
            }
            AddField(emb, "Upcoming Games", text, false)
            s.ChannelMessageSendEmbed(m.ChannelID, emb)
        case "&b":
           emb := new(discordgo.MessageEmbed)
           CheckForShopItem(m.Author.ID, "snoil", 0)
           emb.Title = "The Shop" + " (" + strconv.Itoa(fans[m.Author.ID].Coins) + ")"
           emb.Color = 8651301
           AddField(emb, "🐍 Snake Oil (" + strconv.Itoa(calculateGrowingPrice(fans[m.Author.ID].Shop["snoil"], 15, 1.8)) + ")", "Increase the maximum amount of money you can bet. It'd go from " + strconv.Itoa(calculateGrowingPrice(fans[m.Author.ID].Shop["snoil"], 10, 1.5)) +" to " + strconv.Itoa(calculateGrowingPrice(fans[m.Author.ID].Shop["snoil"] + 1, 10, 1.5)) + ".", false)
           AddField(emb, "🎟️ Vote (100)", "Participate in Democracy.", false)
           AddField(emb, "🥂 Fairweather Flute (2000)", "Change your favorite team. Your previous team will be a little sad, but they'll understand.", false)
           AddField(emb, "🐍 Snake Oil (" + strconv.Itoa(calculateGrowingPrice(fans[m.Author.ID].Shop["snoil"], 15, 1.8)) + ")", "Increase the maximum amount of money you can bet. It'd go from " + strconv.Itoa(calculateGrowingPrice(fans[m.Author.ID].Shop["snoil"], 10, 1.5)) +" to " + strconv.Itoa(calculateGrowingPrice(fans[m.Author.ID].Shop["snoil"] + 1, 10, 1.5)) + ".", false)
           AddField(emb, "🥺 Beg (FREE)", "Uses alchemy to convert your brokeness into a few coins.", false)
           s.ChannelMessageSendEmbed(m.ChannelID, emb)
       case "&e":
          emb := new(discordgo.MessageEmbed)
          CheckForShopItem(m.Author.ID, "snoil", 0)
          emb.Title = "The Election" + " (" + strconv.Itoa(fans[m.Author.ID].Votes) + ")"
          emb.Color = 8651301
          AddField(emb, "DECREES", "Decrees are permanent modifications to the rules of Blaseball.", false)
          AddField(emb, "👂 Communication", "Let's talk.", false)
          AddField(emb, "☎️ Discourse", "Let's listen.", false)
          AddField(emb, "BLESSINGS", "Blessings are modifications to teams.", false)
          AddField(emb, "✨ Magic", "Magic learn to do Magic Damage.", false)
          AddField(emb, "⚛️ Quantum Mechanics", "Quantum teams are affected by macroscopic quantum phenomena.", false)
          AddField(emb, "💤 Lullaby", "Lullaby teams learn to sing a lullaby.", false)
          s.ChannelMessageSendEmbed(m.ChannelID, emb)
        default:
            if strings.HasPrefix(m.Content, "&s ") {
                cont := strings.ToLower(m.Content[3:len(m.Content)])
                split := strings.Split(cont, ">")
                for i := range split {
                    split[i] = FixUnnecesarySpaces(split[i])
                }
                fmt.Println(cont)
                // Fan inserted more than three args on command
                if len(split) > 3 {
                    s.ChannelMessageSend(m.ChannelID, "Expected less than 3 arguments, got " + strconv.Itoa(len(split)))
                } else if len(split) == 1 { // Only look for a team
                    for i := range teams {
                        team := teams[i]
                        if strings.ToLower(team.Name) == strings.ToLower(split[0]) || team.Icon == split[0] {
                            emb := new(discordgo.MessageEmbed)
                            emb.Title = team.Name
                            emb.Color = 8651301
                            text := ""
                            for k := range team.Lineup {
                                text += players[team.Lineup[k]].Name + " " + GetModEmojis(*players[team.Lineup[k]]) + " " + ShowAsStars(players[team.Lineup[k]].Batting) + "\n"
                            }
                            AddField(emb, "Lineup", text, false)
                            text = ""
                            for k := range team.Rotation {
                                text += players[team.Rotation[k]].Name + " " + GetModEmojis(*players[team.Rotation[k]])  + " " + ShowAsStars(players[team.Rotation[k]].Pitching) + "\n"
                            }
                            AddField(emb, "Rotation", text, false)
                            s.ChannelMessageSendEmbed(m.ChannelID, emb)
                        }
                    }
                } else if len(split) == 2 { // Look for a player
                    for i := range teams {
                        team := teams[i]
                        if strings.ToLower(team.Name) == strings.ToLower(split[0]) || team.Icon == split[0] {
                            var player Player
                            for k := range team.Lineup {
                                if strings.ToLower(players[team.Lineup[k]].Name) == strings.ToLower(split[1]) {
                                    player = *players[team.Lineup[k]]
                                }
                            }
                            for k := range team.Rotation {
                                if strings.ToLower(players[team.Rotation[k]].Name) == strings.ToLower(split[1]) {
                                    player = *players[team.Rotation[k]]
                                }
                            }

                            emb := new(discordgo.MessageEmbed)
                            emb.Title = player.Name
                            emb.Color = 8651301
                            AddField(emb, "Modifiers", GetModEmojisAndNames(player), false)
                            AddField(emb, "Team", team.Icon + " " + team.Name, false)
                            stats := "Batting: " + ShowAsStars(player.Batting) + "\n"
                            stats += "Pitching: " + ShowAsStars(player.Pitching) + "\n"
                            stats += "Defense: " + ShowAsStars(player.Defense) + "\n"
                            stats += "Blaserunning: " + ShowAsStars(player.Blaserunning) + "\n"
                            AddField(emb, "Stats", stats, false)
                            AddField(emb, "Pregame Ritual", player.Ritual, false)
                            AddField(emb, "Blood Type", player.Blood + player.Rh, false)
                            AddField(emb, "Favorite Food", player.Food, false)
                            AddField(emb, "Favorite Drink", player.Drink, false)
                            s.ChannelMessageSendEmbed(m.ChannelID, emb)
                        }
                    }
                }
            } else if strings.HasPrefix(m.Content, "&f ") { // Favoriting teams
                cont := strings.ToLower(m.Content[3:len(m.Content)])
                cont = FixUnnecesarySpaces(cont)
                if CheckForShopItem(m.Author.ID, "flute", 1) > 0 {
                    for i := range teams {
                        team := teams[i]
                        if strings.ToLower(team.Name) == strings.ToLower(cont) || team.Icon == cont {
                            emb := new(discordgo.MessageEmbed)
                            emb.Title = "Favorite team set"
                            emb.Color = 8651301
                            AddField(emb, teams[i].Icon + " " + teams[i].Name, teams[i].Description, false)
                            s.ChannelMessageSendEmbed(m.ChannelID, emb)
                        }
                    }
                } else {
                    s.ChannelMessageSend(m.ChannelID, "Not enough flutes!")
                }
            } else if strings.HasPrefix(m.Content, "&m ") { // Favoriting teams
                cont := strings.ToLower(m.Content[3:len(m.Content)])
                cont = FixUnnecesarySpaces(cont)
                text := ""
                for _, i := range mods {
                    if strings.ToLower(modNames[i]) == cont || modIcons[i] == cont {
                        text += "**" + modIcons[i] + " " + modNames[i] + "**\n"
                        text += modDescs[i]
                    }
                }
            } else if strings.HasPrefix(m.Content, "&be ") { // Shop
                cont := strings.ToLower(m.Content[4:len(m.Content)])
                split := strings.Split(cont, ">")
                // This gets rid of final, initial and double spaces
                // Makes commands easier to type
                for i, k := range split {
                    split[i] = FixUnnecesarySpaces(string(k))
                }
                if len(split) > 2 {
                    s.ChannelMessageSend(m.ChannelID, "Expected less than 2 arguments, got " + strconv.Itoa(len(split)))
                } else {
                    if split[0] != "" {
                        amount := 1
                        if len(split) != 1 {
                            amt, err := strconv.Atoi(split[1])
                            if err != nil {
                                s.ChannelMessageSend(m.ChannelID, "Second argument is not an integer.")
                                return
                            }
                            amount = amt
                            CreateFanIfNotExist(m.Author.ID)
                            if fans[m.Author.ID].Coins >= amount {
                                for i, j := range teams {
                                    if strings.ToLower(j.Name) == split[0] || j.Icon == split[0] {
                                        for n, k := range upcoming {
                                            if i == k.Home {
                                                canVote := true
                                                for l := range k.betsAway {
                                                    if l == m.Author.ID {
                                                        canVote = false
                                                    }
                                                }
                                                if canVote {
                                                    s.ChannelMessageSend(m.ChannelID, "Bet placed!")
                                                    upcoming[n].betsHome[m.Author.ID] = amount
                                                    fans[m.Author.ID].Coins -= amount
                                                    return
                                                } else {
                                                    s.ChannelMessageSend(m.ChannelID, "You've already bet on this game.")
                                                    return
                                                }
                                            } else {
                                                canVote := true
                                                for l := range k.betsHome {
                                                    if l == m.Author.ID {
                                                        canVote = false
                                                    }
                                                }
                                                if canVote {
                                                    s.ChannelMessageSend(m.ChannelID, "Bet placed!")
                                                    upcoming[n].betsAway[m.Author.ID] = amount
                                                    fans[m.Author.ID].Coins -= amount
                                                    return
                                                } else {
                                                    s.ChannelMessageSend(m.ChannelID, "You've already bet on this game.")
                                                    return
                                                }
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            } else if strings.HasPrefix(m.Content, "&b ") { // Shop
                cont := strings.ToLower(m.Content[3:len(m.Content)])
                split := strings.Split(cont, ">")
                // This gets rid of final, initial and double spaces
                // Makes commands easier to type
                for i, k := range split {
                    split[i] = FixUnnecesarySpaces(string(k))
                }
                if len(split) > 2 {
                    s.ChannelMessageSend(m.ChannelID, "Expected less than 2 arguments, got " + strconv.Itoa(len(split)))
                } else {
                    print(split[0])
                    if split[0] != "" {
                        amount := 1
                        if len(split) != 1 {
                            amt, err := strconv.Atoi(split[1])
                            if err != nil {
                                s.ChannelMessageSend(m.ChannelID, "Second argument is not an integer.")
                                return
                            }
                            amount = amt
                        }

                        switch split[0] {
                        case "🐍":
                            CheckForShopItem(m.Author.ID, "snoil", 0)
                            if fans[m.Author.ID].Coins > calculateGrowingPrice(amount - 1 + fans[m.Author.ID].Shop["snoil"], 15, 1.8) {
                                fans[m.Author.ID].Coins -= calculateGrowingPrice(amount - 1 + fans[m.Author.ID].Shop["snoil"], 15, 1.8)
                                fans[m.Author.ID].Shop["snoil"] += amount
                                s.ChannelMessageSend(m.ChannelID, "Bought " + strconv.Itoa(amount) + " snake oil for " + strconv.Itoa(calculateGrowingPrice(fans[m.Author.ID].Shop["snoil"]-1, 15, 1.8)) + " coins.")
                            } else {
                                s.ChannelMessageSend(m.ChannelID, "Not enough coins.")
                            }
                        case "🎟️":
                            if fans[m.Author.ID].Coins > 100 * amount {
                                fans[m.Author.ID].Votes += amount
                                fans[m.Author.ID].Coins -= 100 * amount
                                s.ChannelMessageSend(m.ChannelID, "Bought " + strconv.Itoa(amount) + " votes for " + strconv.Itoa(100 * amount) + " coins.")
                            } else {
                                s.ChannelMessageSend(m.ChannelID, "Not enough coins.")
                            }
                        case "🥂":
                            CheckForShopItem(m.Author.ID, "flute", 1)
                            if fans[m.Author.ID].Coins > 2000 * amount {
                                fans[m.Author.ID].Shop["flute"] += amount
                                fans[m.Author.ID].Coins -= 2000 * amount
                                s.ChannelMessageSend(m.ChannelID, "Bought " + strconv.Itoa(amount) + " flutes for " + strconv.Itoa(2000 * amount) + " coins.")
                            } else {
                                s.ChannelMessageSend(m.ChannelID, "Not enough coins.")
                            }
                        case "🥺":
                            if fans[m.Author.ID].Coins <= 0 {
                                fans[m.Author.ID].Coins = rand.Intn(12)
                                s.ChannelMessageSend(m.ChannelID, "Conversion has succeeded! " + strconv.Itoa(fans[m.Author.ID].Coins) + " coins appear in your pockets.")
                            } else {
                                s.ChannelMessageSend(m.ChannelID, "Not enough brokeness.")
                            }
                        }
                    }
                }
            } else if strings.HasPrefix(m.Content, "&e ") { // Shop
                cont := strings.ToLower(m.Content[3:len(m.Content)])
                CreateFanIfNotExist(m.Author.ID)
                if fans[m.Author.ID].Team == "" {
                    return
                }
                split := strings.Split(cont, ">")
                // This gets rid of final, initial and double spaces
                // Makes commands easier to type
                for i, k := range split {
                    split[i] = FixUnnecesarySpaces(string(k))
                }
                if len(split) > 2 {
                    s.ChannelMessageSend(m.ChannelID, "Expected less than 2 arguments, got " + strconv.Itoa(len(split)))
                } else {
                    print(split[0])
                    if split[0] != "" {
                        amount := 1
                        if len(split) != 1 {
                            amt, err := strconv.Atoi(split[1])
                            if err != nil {
                                s.ChannelMessageSend(m.ChannelID, "Second argument is not an integer.")
                                return
                            }
                            amount = amt
                        }

                        switch split[0] {
                        case "👂":
                            if fans[m.Author.ID].Votes >= amount {
                                fans[m.Author.ID].Votes -= amount
                                election["comm"] = election["comm"] + 1
                                s.ChannelMessageSend(m.ChannelID, "Voted!")
                            } else {
                                s.ChannelMessageSend(m.ChannelID, "Not enough votes.")
                            }
                        case "☎️":
                            if fans[m.Author.ID].Votes >= amount {
                                fans[m.Author.ID].Votes -= amount
                                election["disc"] = election["disc"] + 1
                                s.ChannelMessageSend(m.ChannelID, "Voted!")
                            } else {
                                s.ChannelMessageSend(m.ChannelID, "Not enough votes.")
                            }
                        case "✨":
                            if fans[m.Author.ID].Votes >= amount {
                                fans[m.Author.ID].Votes -= amount
                                bless1[fans[m.Author.ID].Team] = bless1[fans[m.Author.ID].Team] + 1
                                s.ChannelMessageSend(m.ChannelID, "Voted!")
                            } else {
                                s.ChannelMessageSend(m.ChannelID, "Not enough votes.")
                            }
                        case "⚛️":
                            if fans[m.Author.ID].Votes >= amount {
                                fans[m.Author.ID].Votes -= amount
                                bless2[fans[m.Author.ID].Team] = bless2[fans[m.Author.ID].Team] + 1
                                s.ChannelMessageSend(m.ChannelID, "Voted!")
                            } else {
                                s.ChannelMessageSend(m.ChannelID, "Not enough votes.")
                            }
                        case "💤":
                            if fans[m.Author.ID].Votes >= amount {
                                fans[m.Author.ID].Votes -= amount
                                bless3[fans[m.Author.ID].Team] = bless3[fans[m.Author.ID].Team] + 1
                                s.ChannelMessageSend(m.ChannelID, "Voted!")
                            } else {
                                s.ChannelMessageSend(m.ChannelID, "Not enough votes.")
                            }
                        }
                    }
                }
            }
        }
    }
}

func calculateGrowingPrice(cAmount int, iPrice int, ex float64) int {
    cAf := float64(cAmount)
    return iPrice + int(math.Pi * cAf) + int(math.Pow(2.71828 * cAf, ex))
}

func CreateFanIfNotExist (id string) {
    sh := make(map[string]int)
    sh["flute"] = 1
    if _, ok := fans[id]; !ok {
        NewFan(id, "", 200, 1, sh, "")
    }
}

func GetAmountOf(id string, item string) int {
    CheckForShopItem(id, item, 0)
    return fans[id].Shop[item]
}

// This gets rid of final, initial and double spaces, making commands easier to type
func FixUnnecesarySpaces(str string) string {
    output := str
    if output != "" {
        for strings.Contains(output, "  ") {
            output = strings.ReplaceAll(output, "  ", " ")
        }
        for string(output[len(output)-1]) == " " {
            output = output[0:len(output)-1]
        }
        for string(output[0]) == " " {
            output = output[1:len(output)]
        }
    }
    return output
}

func CheckForShopItem (id string, item string, retval int) int {
    CreateFanIfNotExist(id)
    if valI, ok := fans[id].Shop[item]; ok {
        return valI
    }
    fans[id].Shop[item] = retval
    return retval
}

// Returns the emojis of a player's modifications
func GetModEmojis(player Player) string{
    output := ""
    for i, j := range player.Modifiers {
        if j != 0 {
            output += modIcons[i]
        }
    }
    return output
}

// Returns [EMOJI mod_name] of a player's modifications
func GetModEmojisAndNames(player Player) string{
    output := ""
    for i, j := range player.Modifiers {
        if j != 0 {
            output += "[ " + modIcons[i] + " " + modNames[i] + " ]"
        }
    }
    if output == "" {
        output = "[ 🍦 Vanilla ]" // No modifiers
    }
    return output
}

// Expresses a float32 as an int. Used showing stats
func ShowAsStars(n float32) string {
    i := int(math.Floor(float64(n)))
    output := ""
    for k := 0; k < i; k++ {
        output += "⭐"
    }
    if float64(i) < float64(n) {
        output += "✨"
    }
    return output
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

// Creates a sloce from a string in :; format
func StringSlice(str string) []string {
    var ret []string
    if str != "" {
        split := strings.Split(str, ";;")
        for k := range split {
            ret = append(ret, split[k])
        }
    }
    return ret
}

// Adds a field to an embed
func AddField (embed *discordgo.MessageEmbed, name string, value string, inline bool) {
    new_field := new(discordgo.MessageEmbedField)
    new_field.Name = name
    new_field.Value = value
    new_field.Inline = inline
    embed.Fields = append(embed.Fields, new_field)
}

func updateDatabases(db *sql.DB) {
    for k := range teams {
        team := teams[k]
        command := `INSERT INTO teams VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) ON CONFLICT (uuid) DO UPDATE SET modifiers = excluded.modifiers, lineup = excluded.lineup, rotation = excluded.rotation, current_pitcher = excluded.current_pitcher`
        _, err := db.Exec(command, team.UUID, strings.Replace(team.Name, "''", "'", -1), strings.Replace(team.Description, "''", "'", -1), team.Icon, SliceString(team.Lineup), SliceString(team.Rotation), MapString(team.Modifiers), team.AvgDef, team.CurrentPitcher)
        fmt.Println(MapString(team.Modifiers))
        CheckError(err)
    }
    for k := range players {
        player := players[k]
        command := `INSERT INTO players VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) ON CONFLICT (uuid) DO UPDATE SET name = excluded.name, modifiers = excluded.modifiers`
        for strings.Contains(player.Name, "''") {
            player.Name = strings.Replace(player.Name, "''", "'", -1)
        }
        _, err := db.Exec(command, player.UUID, strings.Replace(player.Name, "''", "'", -1), strings.Replace(player.Team, "''", "'", -1), player.Batting, player.Pitching, player.Defense, player.Blaserunning, MapString(player.Modifiers), player.Blood, player.Rh, player.Drink, player.Food, player.Ritual)
        CheckError(err)
    }
    for k := range fans {
        fan := fans[k]
        command := `INSERT INTO fans VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id) DO UPDATE SET favorite_team = excluded.favorite_team, coins = excluded.coins, votes = excluded.votes, shop = excluded.shop, stan = excluded.stan`
        _, err := db.Exec(command, fan.Id, fan.Team, fan.Coins, fan.Votes, MapString(fan.Shop), fan.Stan)
        CheckError(err)
    }
    for k := range field {
        player := field[k]
        command := `INSERT INTO fans VALUES ($1) ON CONFLICT (id) DO NOTHING`
        _, err := db.Exec(command, player)
        CheckError(err)
    }
    command := `INSERT INTO seasons VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) ON CONFLICT (id) DO UPDATE SET seasons day = excluded.day, tape = excluded.tape, election = excluded.election, wins = excluded.wins, losses = excluded.losses, b1 = excluded.b1, b2 = excluded.b2, b3 = excluded.b3`
    _, err := db.Exec(command, seasonNumber, day, tape, MapString(election), MapString(wins), MapString(losses), MapString(bless1), MapString(bless2), MapString(bless3))
    CheckError(err)
}

/* IN CASE THE DATABASE IS RESET

Comment the parts of the code that load the database, uncomment the parts that
generate the data, create code that insert them into the database. Once that is
done, comment the parts that generate and insert and uncomment the parts
that load it */
