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
}

const CommandChannelId = "823421429601140756"
const GamesChannelId = "823421360768417824"

var CurrentGamesId string
var CurrentGamesId2 string
var CurrentGamesId3 string

var cool_league [10]string
var fun_league [10]string

var players map[string]Player
var teams map[string]Team

var name_pool []string
var drink_pool []string
var food_pool []string
var blood_pool []string
var rh_pool []string
var ritual_pool []string

var upcoming []Game
var games []Game

var validuuids []string

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
    teams[team.UUID] = *team
    return team.UUID
}

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

    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS leagues(
        cool TEXT UNIQUE,
        fun TEXT UNIQUE,
    )`)

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
        err := rows.Scan(&uuid, &name, &description, &icon, &lineup, &rotation, &modifiers, &AvgDef, &currentPitcher)
        CheckError(err)
        TeamWithData(name, description, icon, uuid, StringSlice(lineup), StringSlice(rotation), StringMap(modifiers), AvgDef, currentPitcher)
    }

    // TODO get all the teams for the leagues from the database
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



    go HandleGames(discord)
    time.Sleep(0)

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

func HandleGames(session *discordgo.Session) {
    for true {
        if len(upcoming) > 0 && len(games) == 0 {
            msg, _ := session.ChannelMessageSend(GamesChannelId, "**Preparing Games**")
            CurrentGamesId = msg.ID
            msg, _ = session.ChannelMessageSend(GamesChannelId, "**Preparing Games**")
            CurrentGamesId2 = msg.ID
            msg, _ = session.ChannelMessageSend(GamesChannelId, "**Preparing Games**")
            CurrentGamesId3 = msg.ID
            for k := range upcoming {
                games = append(games, upcoming[k])
                // go Handle(&games[len(games)-1], k, session)

            }
            upcoming = make([]Game, 0)
        } else if len(games) == 0 {
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

        if len(games) > 0 {
            HandlePlays(session, CurrentGamesId, 0, 3)
            HandlePlays(session, CurrentGamesId2, 3, 6)
            HandlePlays(session, CurrentGamesId3, 6, len(games))
            time.Sleep(1 * time.Second)
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
                game.BattingTeam = teams[game.Home]
                game.PitchingTeam = teams[game.Away]
                game.Batter = players[game.BattingTeam.Lineup[game.BatterHome]]
                game.Pitcher = players[game.PitchingTeam.Rotation[game.PitchingTeam.CurrentPitcher]]
                if !game.Top {
                    game.BattingTeam  = teams[game.Away]
                    game.PitchingTeam = teams[game.Home]
                    game.Batter = players[game.BattingTeam.Lineup[game.BatterAway]]
                    game.Pitcher = players[game.PitchingTeam.Rotation[game.PitchingTeam.CurrentPitcher]]
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
                        game.Batter = players[game.BattingTeam.Lineup[game.BatterAway]]
                    } else {
                        game.BatterHome = (game.BatterHome + 1) % len(game.BattingTeam.Lineup)
                        game.Batter = players[game.BattingTeam.Lineup[game.BatterHome]]
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
                        } else if happen < game.Batter.Batting * 0.7 { // Singlet
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
                                fmt.Println("Finished")
                                continue Loop
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
            this_output += "1️⃣ " + players[game.AnnouncementStates[0].Bases[0]].Name + "\n"
            this_output += "2️⃣ " + players[game.AnnouncementStates[0].Bases[1]].Name + "\n"
            this_output += "3️⃣ " + players[game.AnnouncementStates[0].Bases[2]].Name + "\n\n"
            this_output += game.Announcements[0] + "\n\n"
            game.LastMessage = this_output
            output += this_output
            game.Announcements = append(game.Announcements[:0], game.Announcements[1:]...)
            game.AnnouncementStates = append(game.AnnouncementStates[:0], game.AnnouncementStates[1:]...)

        }
    }
    if output != "" {
        fmt.Println(len(output))
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

func Handle(game *Game, index int, session *discordgo.Session) {
    InningState := "starting" // starting, inning, outing
    /* When a half-inning inning starts, the innning's state is starting
    When a half-inning is ongoing, the inning's state is inning
    When the whole inning ends, the inning's state is outing*/
    var announcements []string
    var announcement_states []Game
    changeBatter := true
    var batting_team Team
    var pitching_team Team
    var batter Player
    var pitcher Player
    Loop:
    for true {
        /* Nothing can go on until the announcements are finished
        This is done to prevent the game from announcing the half-inning
        After its player steps up to bat (error from the previous iteration of the sim)
        Among other things*/
        if len(game.Announcements) == 0 {
            switch InningState {
            case "inning", "starting":
                // If it's the top of the inning, the batter is the home team
                // And the pitcher is the away team
                batting_team = teams[game.Home]
                pitching_team = teams[game.Away]
                batter = players[batting_team.Lineup[game.BatterHome]]
                pitcher = players[pitching_team.Rotation[pitching_team.CurrentPitcher]]
                if !game.Top {
                    batting_team = teams[game.Away]
                    pitching_team = teams[game.Home]
                    batter = players[batting_team.Lineup[game.BatterAway]]
                    pitcher = players[pitching_team.Rotation[pitching_team.CurrentPitcher]]
                }

                if InningState == "starting" {
                    // Reset everything when the half-inning is starting and announce the half-inning
                    if game.Top {
                        announcements = append(announcements, "Top of " + strconv.Itoa(game.Inning))
                        announcement_states = append(announcement_states, *game)
                    } else {
                        announcements = append(announcements, "Bottom of " + strconv.Itoa(game.Inning))
                        announcement_states = append(announcement_states, *game)
                    }
                    game.Bases = [3]string{"", "", ""}
                    game.Outs = 0
                    game.Strikes = 0
                    game.Balls = 0

                    InningState = "inning"
                }

                // If in the last tick it was determined that the batter should change
                if changeBatter {
                    game.Strikes = 0
                    game.Balls = 0
                    if !game.Top {
                        game.BatterAway = (game.BatterAway + 1) % len(batting_team.Lineup)
                        batter = players[batting_team.Lineup[game.BatterAway]]
                    } else {
                        game.BatterHome = (game.BatterHome + 1) % len(batting_team.Lineup)
                        batter = players[batting_team.Lineup[game.BatterHome]]
                    }
                    announcements = append(announcements, batter.Name + " batting for the " + batting_team.Name)
                    announcement_states = append(announcement_states, *game)
                    changeBatter = false
                }

                if InningState == "inning" {
                    probability := batter.Batting + pitcher.Pitching //The range of probabilities
                    happen := rand.Float32() * probability // What actually happens
                    // The batter manages to bat
                    if happen < batter.Batting {
                        var runsScored int
                        if happen < batter.Batting * 0.01 { // Homer
                            if batter.UUID == game.Bases[0] && game.Bases[0] == game.Bases[1] && game.Bases[1] == game.Bases[2] {
                                announcements = append(announcements, batter.Name + " hits a solo grand slam!?!?!?")
                            } else {
                                if game.Bases[0] != "" && game.Bases[1] != "" && game.Bases[2] != "" {
                                    announcements = append(announcements, batter.Name + " hits a grand slam!!!!!")
                                } else {

                                    announcements = append(announcements, batter.Name + " hits a home run!!!")
                                }
                            }
                            runsScored += Advance(&game.Bases, batter.UUID, -1)
                            runsScored += Advance(&game.Bases, "", 0)
                            runsScored += Advance(&game.Bases, "", 1)
                            runsScored += Advance(&game.Bases, "", 2)

                            game.Strikes = 0
                            game.Balls = 0
                        } else if happen < batter.Batting * 0.025 { // Triplet
                            announcements = append(announcements, batter.Name + " hits a triple!!!")
                            runsScored += Advance(&game.Bases, batter.UUID, -1)
                            runsScored += Advance(&game.Bases, "", 0)
                            runsScored += Advance(&game.Bases, "", 1)
                            game.Strikes = 0
                            game.Balls = 0
                        } else if happen < batter.Batting * 0.1 { // Twin
                            announcements = append(announcements, batter.Name + " hits a double!!")
                            runsScored += Advance(&game.Bases, batter.UUID, -1)
                            runsScored += Advance(&game.Bases, "", 0)
                            game.Strikes = 0
                            game.Balls = 0
                        } else if happen < batter.Batting * 0.7 { // Singlet
                            announcements = append(announcements, batter.Name + " hits a single!")
                            runsScored += Advance(&game.Bases, batter.UUID, -1)
                            game.Strikes = 0
                            game.Balls = 0

                        } else if happen < batter.Batting * 0.999 {
                            announcements = append(announcements, batter.Name + " hits a flyout!")
                            game.Outs += 1
                            changeBatter = true
                        } else {
                            announcements = append(announcements, batter.Name + " hits the ball so hard it goes to another game!")
                        }
                        announcement_states = append(announcement_states, *game)
                        if runsScored != 0 {
                            if game.Top {
                                game.RunsHome += runsScored
                            } else {
                                game.RunsAway += runsScored
                            }
                            if runsScored == 1 {
                                announcements = append(announcements, strconv.Itoa(runsScored) + " run scored.")
                            } else {
                                announcements = append(announcements, strconv.Itoa(runsScored) + " runs scored.")
                            }
                            announcement_states = append(announcement_states, *game)
                        }
                        changeBatter = true

                    // The batter fails to bat
                    } else {
                        if happen < batter.Batting + pitcher.Pitching * 0.1 {
                            announcements = append(announcements, "Ball.")
                        } else if happen < batter.Batting + pitcher.Pitching * 0.25 {
                            announcements = append(announcements, "Strike, swinging.")
                            game.Strikes += 1
                        } else if happen < batter.Batting + pitcher.Pitching * 0.65 {
                            announcements = append(announcements, "Strike, looking.")
                            game.Strikes += 1
                        } else if happen < batter.Batting + pitcher.Pitching * 0.99999 {
                            announcements = append(announcements, "Strike, flinching.")
                            game.Strikes += 1
                        } else {
                            announcements = append(announcements, "Strike, knows too much.")
                            game.Strikes += 1
                        }
                        announcement_states = append(announcement_states, *game)
                    }

                    if game.Balls >= 4 {
                        game.Strikes = 0
                        game.Balls = 0
                        announcements = append(announcements, batter.Name + " gets a walk.")
                        announcement_states = append(announcement_states, *game)
                        Advance(&game.Bases, batter.UUID, -1)
                        changeBatter = true
                    }

                    if game.Strikes >= 3 {
                        game.Strikes = 0
                        game.Balls = 0
                        game.Outs += 1
                        announcements = append(announcements, batter.Name + " strikes out.")
                        announcement_states = append(announcement_states, *game)
                        changeBatter = true
                    }
                    if game.Outs >= 3 {
                        if !game.Top {
                            game.Inning += 1
                            // Game is Over
                            if game.RunsAway != game.RunsHome && game.Inning >= 9 {
                                games = append(games[:index], games[index+1:]...)
                                fmt.Println("Finished")
                                break Loop
                            }
                            announcements = append(announcements, "Inning " + strconv.Itoa(game.Inning-1) + " is now an outing.")
                            announcement_states = append(announcement_states, *game)
                        }
                        game.Top = !game.Top
                        InningState = "starting"
                        changeBatter = true
                    }
                }
            }
        } else {
            // fmt.Println(announcements[0])

            emb := new(discordgo.MessageEmbed)
            foot := new(discordgo.MessageEmbedFooter)
            foot.Text = "React with any of the emotes shown to get more information about them."
            if announcement_states[0].Top {
                emb.Title = "🔺"
            } else {
                emb.Title = "🔻"
            }
            emb.Title += strconv.Itoa(announcement_states[0].Inning)
            emb.Color = 8651301
            emb.Footer = foot
            AddField(emb, teams[announcement_states[0].Home].Icon + " " + teams[announcement_states[0].Home].Name, strconv.Itoa(announcement_states[0].RunsHome), true)
            AddField(emb, teams[announcement_states[0].Away].Icon + " " + teams[announcement_states[0].Away].Name, strconv.Itoa(announcement_states[0].RunsAway), true)
            if announcement_states[0].Top {
                AddField(emb, "Inning", " 🔺" + strconv.Itoa(announcement_states[0].Inning), true)
            } else {
                AddField(emb, "Inning", " 🔻" + strconv.Itoa(announcement_states[0].Inning), true)
            }
            AddField(emb, "Outs", strconv.Itoa(announcement_states[0].Outs), true)
            AddField(emb, "Strikes", strconv.Itoa(announcement_states[0].Strikes), true)
            AddField(emb, "Balls", strconv.Itoa(announcement_states[0].Balls), true)
            AddField(emb, "🏏 Batting", batter.Name, true)
            AddField(emb, "⚾ Pitching", pitcher.Name, true)
            if announcement_states[0].Bases[0] != "" {
                AddField(emb, "Base 1️⃣", players[announcement_states[0].Bases[0]].Name, false)
            } else {
                AddField(emb, "Base 1️⃣", "Empty", false)
            }
            if announcement_states[0].Bases[1] != "" {
                AddField(emb, "Base 2️⃣", players[announcement_states[0].Bases[1]].Name, false)
            } else {
                AddField(emb, "Base 2️⃣", "Empty", false)
            }
            if announcement_states[0].Bases[2] != "" {
                AddField(emb, "Base 3️⃣", players[announcement_states[0].Bases[2]].Name, false)
            } else {
                AddField(emb, "Base 3️⃣", "Empty", false)
            }
            AddField(emb, "🍿 Events", announcements[0], false)
            _, err := session.ChannelMessageEditEmbed(GamesChannelId, announcement_states[0].MessageId, emb)
            CheckError(err)
            announcements = append(announcements[:0], announcements[1:]...)
            announcement_states = append(announcement_states[:0], announcement_states[1:]...)

        }
        time.Sleep(500 * time.Millisecond)
    }
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
            for k := range teams {
                AddField(emb, teams[k].Icon + " " + teams[k].Name, teams[k].Description, false)
            }
            s.ChannelMessageSendEmbed(m.ChannelID, emb)
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
