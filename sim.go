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
    "math"
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
}
type Fan struct {
    Id string
    Team string
    Coins int
    Votes int
    Shop map[string]int
    Stan string
}

const CommandChannelId = "823421429601140756"
const GamesChannelId = "823421360768417824"

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

var upcoming []Game
var games []Game

var validuuids []string

var Weathers []string = []string{"Sunny"}

var modNames map[string]string = map[string]string {
    "intangible" : "Intangible",
    "still_alive" : "Still Alive",
}

var modDescs map[string]string = map[string]string {
    "intangible" : "This player is intangible.",
    "still_alive" : "When this player's dying they'll be Still Alive.",
}
var modIcons map[string]string = map[string]string{
    "intangible" : "🌫️",
    "still_alive" : "🐱",
}

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
    // Players have a 20% chance of being ghosts
    if rand.Float32() < 0.2 {
        player.Modifiers["intangible"] = -1
    }
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
    game.Weather = "rain"
    game.Inning = 1
    game.MaxInnings = innings
    game.Top = true
    game.RunsAway, game.RunsHome = 0, 0
    game.MaxInnings = innings
    game.InningState = "starting"
    game.Announcements = []string{}
    game.AnnouncementStates = []Game{}
    game.ChangeBatter = true
    upcoming = append(upcoming, *game)
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
    // Make sure the RNG is random
    rand.Seed(time.Now().Unix())
    // Load the .env file, this has to be discarded for heroku releases
    envs, err := godotenv.Read(".env")
    CheckError(err)

    // Set up the bot using the bot key thats found in the environment variable
    discord, err := discordgo.New("Bot " + envs["BOT_KEY"])
    // discord, err := discordgo.New("Bot " + os.Getenv("BOT_KEY"))
    CheckError(err)

    db, err := sql.Open("postgres", envs["DATABASE_URL"])
    // db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
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
    _, err = db.Exec(`DROP TABLE fans`)
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS fans(
        id TEXT PRIMARY KEY,
        favorite_team TEXT,
        coins INTEGER,
        votes INTEGER,
        shop TEXT,
        stan TEXT,
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
        modifieds := StringMap(modifiers)
        if name == "Adjacent Alyssa" {
            name = "Normal Cowboy"
            modifieds["still_alive"] = -1
            modifieds["intangible"] = -1
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
        CheckError(err)
        TeamWithData(name, description, icon, uuid, StringSlice(lineup), StringSlice(rotation), StringMap(modifiers), AvgDef, currentPitcher)
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
    go HandleGames(discord)
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

func HandleGames(session *discordgo.Session) {
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
            upcoming = make([]Game, 0)

        } else if len(games) == 0 { // If there are no upcoming games nor current games
            i := 0
            var TeamA string
            var TeamB string
            for k := range teams {
                if i % 2 == 0 {
                    TeamA = string(k)
                } else {
                    TeamB = string(k)
                    NewGame(TeamA, TeamB, 9)
                }
                i += 1
            }
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
            time.Sleep(10 * time.Millisecond)
        }

        time.Sleep(0)
    }
}

func HandlePlays (session *discordgo.Session, message string, start int, end int) {
    output := ""
    Loop:
    for k := start; k < end; k++ {
        game := &games[k]
        /* Nothing can go on until the announcements are finished
        This is done to prevent the game from announcing the half-inning
        After its player steps up to bat (error from the previous iteration of the sim)
        Among other things*/
        if len(game.Announcements) == 0 {
            switch game.InningState {
            case "inning", "starting":
                // If it's the top of the inning, the batter is the home team
                // And the pitcher is the away team
                game.BattingTeam = *teams[game.Home]
                game.PitchingTeam = *teams[game.Away]
                game.Batter = *players[game.BattingTeam.Lineup[game.BatterHome]]
                game.Pitcher = *players[game.PitchingTeam.Rotation[game.PitchingTeam.CurrentPitcher]]
                if !game.Top {
                    game.BattingTeam  = *teams[game.Away]
                    game.PitchingTeam = *teams[game.Home]
                    game.Batter = *players[game.BattingTeam.Lineup[game.BatterAway]]
                    game.Pitcher = *players[game.PitchingTeam.Rotation[game.PitchingTeam.CurrentPitcher]]
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
                    } else {
                        game.BatterHome = (game.BatterHome + 1) % len(game.BattingTeam.Lineup)
                        game.Batter = *players[game.BattingTeam.Lineup[game.BatterHome]]
                    }
                    game.Announcements = append(game.Announcements, game.Batter.Name + " batting for the " + game.BattingTeam.Name)
                    game.AnnouncementStates = append(game.AnnouncementStates, *game)
                    game.ChangeBatter = false
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
                        game.ChangeBatter = true

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
                            if game.RunsAway != game.RunsHome && game.Inning >= 9 {
                                games = append(games[:k], games[k+1:]...)
                                // TODO give money to the players that bet
                                // TODO send "game ended" message
                                if game.RunsAway > game.RunsHome {
                                    for i, j := range game.betsAway {
                                        fans[i].Coins += j * 2 + rand.Intn(7) - 1
                                    }
                                } else {
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
        } else {
            // fmt.Println(game.Announcements[0])
            this_output := ""
            this_output += "**" + teams[game.Home].Name + " at " + teams[game.Away].Name + "**" + "\n"
            this_output += teams[game.Home].Icon + " **" + teams[game.Home].Name + "** (" + strconv.Itoa(game.AnnouncementStates[0].RunsHome) + ")\n"
            if game.AnnouncementStates[0].Top {
                this_output += "🔺" + strconv.Itoa(game.AnnouncementStates[0].Inning) + "\n"
            } else {
                this_output += "🔻" + strconv.Itoa(game.AnnouncementStates[0].Inning) + "\n"
            }
            this_output += teams[game.Away].Icon + " **" + teams[game.Away].Name + "** (" + strconv.Itoa(game.AnnouncementStates[0].RunsAway) + ")\n\n"
            this_output += "🏏 " + game.AnnouncementStates[0].Batter.Name + "\n"
            this_output += "⚾ " + game.AnnouncementStates[0].Pitcher.Name + "\n\n"
            this_output += "**Outs:** " + ShowInCircles(game.AnnouncementStates[0].Outs, 3) + "\n"
            this_output += "**Strikes:** " + ShowInCircles(game.AnnouncementStates[0].Strikes, 3) + "\n"
            this_output += "**Balls:** " + ShowInCircles(game.AnnouncementStates[0].Balls, 4) + "\n\n"
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
            game.Announcements = append(game.Announcements[:0], game.Announcements[1:]...)
            game.AnnouncementStates = append(game.AnnouncementStates[:0], game.AnnouncementStates[1:]...)
        }
    }
    if output != "" {
        _, err := session.ChannelMessageEdit(GamesChannelId, message, output)
        CheckError(err)
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
            for k := range games {
                text += teams[games[k].Home].Icon + " " + teams[games[k].Home].Name + " - " + teams[games[k].Away].Name + " " + teams[games[k].Away].Icon + "\n"
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
           AddField(emb, "🥺 Beg (FREE)", "Uses alchemy to convert your brokeness into a few coins.", false)
           s.ChannelMessageSendEmbed(m.ChannelID, emb)
        default:
            if strings.HasPrefix(m.Content, "&s ") {
                cont := strings.ToLower(m.Content[3:len(m.Content)])
                split := strings.Split(cont, ">")
                for i := range split {
                    FixUnnecesarySpaces(split[i])
                }
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
                FixUnnecesarySpaces(cont)
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
                            for _, i := range teams {
                                if strings.ToLower(i.Name) == split[0] || i.Icon == split[0] {
                                    for i, k := range upcoming {
                                        if i == k.Home {
                                            canBet := true
                                            for i := range k.betsAway {
                                                if i == m.Author.ID {
                                                    canBet = false
                                                }
                                            }
                                            if canBet {
                                                games[i].betsHome[m.Author.ID] = amount
                                                s.ChannelMessageSend(m.ChannelID, "Bet placed!")
                                            } else {
                                                s.ChannelMessageSend(m.ChannelID, "You've already bet for that game.")
                                            }
                                        } else if i == k.Away {
                                            canBet := true
                                            for i := range k.betsHome {
                                                if i == m.Author.ID {
                                                    canBet = false
                                                }
                                            }
                                            if canBet {
                                                games[i].betsAway[m.Author.ID] = amount
                                                s.ChannelMessageSend(m.ChannelID, "Bet placed!")
                                            } else {
                                                s.ChannelMessageSend(m.ChannelID, "You've already bet for that game.")
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
                                fans[m.Author.ID].Shop["snoil"] += amount
                                fans[m.Author.ID].Coins -= calculateGrowingPrice(amount - 1 + fans[m.Author.ID].Shop["snoil"], 15, 1.8)
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
    for i := range player.Modifiers {
        output += modIcons[i]
    }
    return output
}

// Returns [EMOJI mod_name] of a player's modifications
func GetModEmojisAndNames(player Player) string{
    output := ""
    for i := range player.Modifiers {
        output += "[ " + modIcons[i] + " " + modNames[i] + " ]"
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

/* IN CASE THE DATABASE IS RESET

Comment the parts of the code that load the database, uncomment the parts that
generate the data, create code that insert them into the database. Once that is
done, comment the parts that generate and insert and uncomment the parts
that load it */
