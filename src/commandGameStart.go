/*
	Sent when the owner of a table clicks on the "Start Game" button
	(the client will send a "hello" message after getting "gameStart")

	"data" is empty
*/

package main

import (
	"hash/crc64"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func commandGameStart(s *Session, d *CommandData) {
	/*
		Validation
	*/

	// Validate that the game exists
	gameID := s.CurrentGame()
	var g *Game
	if v, ok := games[gameID]; !ok {
		s.Error("Game " + strconv.Itoa(gameID) + " does not exist.")
		return
	} else {
		g = v
	}

	// Validate that the game has at least 2 players
	if len(g.Players) < 2 {
		s.Error("You need at least 2 players before you can start a game.")
		return
	}

	// Validate that the game is not started yet
	if g.Running {
		s.Error("That game has already started, so you cannot start it.")
		return
	}

	// Validate that this is the owner of the game
	if g.Owner != s.UserID() {
		s.Error("Only the owner of a game can start it.")
		return
	}

	/*
		Start
	*/

	// Create the deck
	// (it will have 60 cards if playing no variant,
	// 65 cards if playing a one of each variant,
	// and 70 cards when playing the other variants)
	suits := []int{0, 1, 2, 3, 4}
	if g.Options.Variant > 0 {
		suits = append(suits, 5)
	}
	for _, suit := range suits {
		ranks := []int{1, 2, 3, 4, 5}
		for _, rank := range ranks {
			var amountToAdd int
			if suit == 5 && (g.Options.Variant == 2 || g.Options.Variant == 7) {
				// Black one of each or Crazy (which includes black one of each)
				amountToAdd = 1
			} else if rank == 1 {
				amountToAdd = 3
			} else if rank == 5 {
				amountToAdd = 1
			} else {
				amountToAdd = 2
			}

			for i := 0; i < amountToAdd; i++ {
				// Add the card to the deck
				card := &Card{
					Suit: suit,
					Rank: rank,
					// We can't set the order here because the deck will be shuffled later
				}
				g.Deck = append(g.Deck, card)
			}
		}
	}

	// Create the stacks
	for i := 0; i < len(suits); i++ {
		g.Stacks = append(g.Stacks, 0)
	}

	// Parse the game name to see if the players want to play a specific deal (read from a text file)
	var presetRegExp *regexp.Regexp
	if v, err := regexp.Compile(`^!preset (.+)$`); err != nil {
		log.Error("Failed to create the preset regular expression:", err)
		s.Error("Failed to create the game. Please contact an administrator.")
		return
	} else {
		presetRegExp = v
	}

	match1 := presetRegExp.FindStringSubmatch(g.Name)
	if match1 != nil {
		// The players want to play a specific deal, so don't bother getting a seed or shuffling the deck
		g.Seed = match1[1]
		filePath := path.Join(projectPath, "specific-deals", g.Seed+".txt")

		if _, err := os.Stat(filePath); err != nil {
			s.Error("That preset deal does not exist on the server.")
			return
		}

		var lines []string
		if v, err := ioutil.ReadFile(filePath); err != nil {
			log.Error("Failed to read \""+filePath+"\":", err)
			s.Error("Failed to create the game. Please contact an administrator.")
			return
		} else {
			lines = strings.Split(string(v), "\n")
		}

		log.Info("Using a preset deal of:", g.Seed)

		var cardRegExp *regexp.Regexp
		if v, err := regexp.Compile(`^(\w)(\d)$`); err != nil {
			log.Error("Failed to create the card regular expression:", err)
			s.Error("Failed to create the game. Please contact an administrator.")
			return
		} else {
			cardRegExp = v
		}

		for i, line := range lines {
			if line == "" {
				continue
			}

			match2 := cardRegExp.FindStringSubmatch(line)
			if match2 == nil {
				log.Error("Failed to parse line "+strconv.Itoa(i+1)+":", line)
				s.Error("Failed to create the game. Please contact an administrator.")
				return
			}

			// Change the suit of all of the cards in the deck
			suit := match2[1]
			newSuit := -1
			if suit == "b" {
				newSuit = 0
			} else if suit == "g" {
				newSuit = 1
			} else if suit == "y" {
				newSuit = 2
			} else if suit == "r" {
				newSuit = 3
			} else if suit == "p" {
				newSuit = 4
			} else if suit == "m" {
				newSuit = 5
			} else {
				log.Error("Failed to parse the suit on line "+strconv.Itoa(i+1)+":", suit)
				s.Error("Failed to create the game. Please contact an administrator.")
				return
			}
			g.Deck[i].Suit = newSuit

			// Change the rank of all of the cards in the deck
			rank := match2[2]
			newRank := -1
			if v, err := strconv.Atoi(rank); err != nil {
				log.Error("Failed to parse the rank on line "+strconv.Itoa(i+1)+":", rank)
				s.Error("Failed to create the game. Please contact an administrator.")
				return
			} else {
				newRank = v
			}
			g.Deck[i].Rank = newRank
		}
	} else {
		// We are not playing on a preset deal
		// Parse the game name to see if the players want to play a specific seed
		var seedRegExp *regexp.Regexp
		if v, err := regexp.Compile(`^!seed (.+)$`); err != nil {
			log.Error("Failed to create the seed regular expression:", err)
			s.Error("Failed to create the game. Please contact an administrator.")
			return
		} else {
			seedRegExp = v
		}

		seedPrefix := "p" + strconv.Itoa(len(g.Players)) + "v" + strconv.Itoa(g.Options.Variant) + "s"
		match2 := seedRegExp.FindStringSubmatch(g.Name)
		if match2 != nil {
			g.Seed = seedPrefix + match2[1]
		} else {
			// Get a list of all the seeds that these players have played before
			seedMap := make(map[string]bool)
			for _, p := range g.Players {
				var seeds []string
				if v, err := db.Games.GetPlayerSeeds(p.ID); err != nil {
					log.Error("Failed to get the past seeds for \""+s.Username()+"\":", err)
					s.Error("Failed to create the game. Please contact an administrator.")
					return
				} else {
					seeds = v
				}

				for _, v := range seeds {
					seedMap[v] = true
				}
			}

			// Find a seed that no-one has played before
			seedNum := 0
			looking := true
			for looking {
				seedNum++
				g.Seed = seedPrefix + strconv.Itoa(seedNum)
				if !seedMap[g.Seed] {
					looking = false
				}
			}
		}

		// Shuffle the deck
		// From: https://stackoverflow.com/questions/12264789/shuffle-array-in-go
		// Convert the string to an uint64 (seeding with negative numbers will not work)
		// We use the CRC64 hash function to do this
		// https://www.socketloop.com/references/golang-hash-crc64-checksum-and-maketable-functions-example
		crc64Table := crc64.MakeTable(crc64.ECMA)
		intSeed := crc64.Checksum([]byte(g.Seed), crc64Table)
		rand.Seed(int64(intSeed))

		for i := range g.Deck {
			j := rand.Intn(i + 1)
			g.Deck[i], g.Deck[j] = g.Deck[j], g.Deck[i]
		}
	}

	log.Info(g.GetName() + "Using seed \"" + g.Seed + "\", timed is " + strconv.FormatBool(g.Options.Timed) + ".")

	// Log the deal (so that it can be distributed to others if necessary)
	log.Info("--------------------------------------------------")
	log.Info("Deal for seed: " + g.Seed + " (from top to bottom)")
	log.Info("(cards are dealt to a player until their hand fills up before moving on to the next one)")
	for i, c := range g.Deck {
		log.Info(strconv.Itoa(i+1) + ") " + c.SuitName(g) + " " + strconv.Itoa(c.Rank))
	}
	log.Info("--------------------------------------------------")

	// Now that we have finished building the deck,
	// initialize all of the players notes based on the number of cards in the deck
	for _, p := range g.Players {
		p.Notes = make([]string, len(g.Deck))
	}

	// Deal the cards
	handSize := 5
	if len(g.Players) > 3 {
		handSize = 4
	}
	for _, p := range g.Players {
		for i := 0; i < handSize; i++ {
			p.DrawCard(g)
		}
	}

	// Get a random player to start first
	g.ActivePlayer = rand.Intn(len(g.Players))
	text := g.Players[g.ActivePlayer].Name + " goes first"
	g.Actions = append(g.Actions, Action{
		Text: text,
	})
	g.Actions = append(g.Actions, Action{
		Type: "turn",
		Num:  0,
		Who:  g.ActivePlayer,
	})
	log.Info(g.GetName() + text)

	// Set the game to running
	g.Running = true
	g.DatetimeStarted = time.Now()

	// Send a "gameStart" message to everyone in the game
	for _, p := range g.Players {
		p.Session.NotifyGameStart()
	}

	// Let everyone know that the game has started, which will turn the
	// "Join Game" button into "Spectate"
	notifyAllTable(g)

	// Set the status for all of the users in the game
	for _, p := range g.Players {
		p.Session.Set("status", "Playing")
		notifyAllUser(p.Session)
	}

	// Start the timer
	g.TurnBeginTime = time.Now()
	if g.Options.Timed {
		go g.CheckTimer(0, g.Players[g.ActivePlayer])
	}

	// Send the list of people who are connected
	// (this governs if a player's name is red or not)
	g.NotifyConnected()

	// Make a sound effect
	g.NotifySound()
}
