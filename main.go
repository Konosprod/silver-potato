package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Chunck int

const (
	CCL = iota
	CCT
	GN
	PROPREL
	VERBE
	ADVERBE
	COD
	CCM
	LAST
)

var (
	Token string
)

const (
	MAXPLAYER = 8
)

type entry struct {
	chunck Chunck
	userID string
	answer string
}

var chunck Chunck = CCL
var entries []entry = make([]entry, 0, LAST)
var initialChannel string = ""

func (c Chunck) String() string {
	return [...]string{
		"complément circonstanciel de lieu",
		"complément circonstanciel de temps",
		"groupe nominal",
		"propisition relative",
		"verbe",
		"adverbe",
		"complément d'objet direct",
		"complément circonstenciel de manière",
	}[c]
}

func reset() {
	chunck = CCL
	initialChannel = ""
	entries = make([]entry, 0, LAST)
}

func indexOf(e []entry, s string) int {

	for i, v := range e {
		if v.userID == s && v.answer == "" {
			return i
		}
	}
	return -1
}

func getAnswer(chunck int) string {
	for _, v := range entries {
		if int(v.chunck) == chunck {
			return v.answer
		}
	}

	return ""
}

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	dg, err := discordgo.New("Bot " + Token)

	if err != nil {
		fmt.Println("error creating Discord session, ", err)
		return
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()

	if err != nil {
		fmt.Println("Error while opening connection, ", err)
		return
	}

	fmt.Println("Bot is running, press CTRL+C to exit")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()

}

func isDM(s *discordgo.Session, m *discordgo.MessageCreate) (bool, error) {
	channel, err := s.State.Channel(m.ChannelID)

	if err != nil {
		if channel, err = s.Channel(m.ChannelID); err != nil {
			return false, err
		}
	}

	return channel.Type == discordgo.ChannelTypeDM, nil
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	if dm, err := isDM(s, m); err == nil {
		if dm {

			index := indexOf(entries, m.Author.ID)

			if index != -1 {
				entries[index].answer = m.Content
				s.ChannelMessageSend(m.ChannelID, "Merci !")
			}

			done := true
			for _, e := range entries {
				if e.answer == "" {
					done = false
				}
			}

			if done {
				rep := ""
				for answer := 0; answer < LAST; answer++ {
					rep += getAnswer(answer) + " "
				}
				s.ChannelMessageSend(initialChannel, rep)
				reset()
			}

		} else {

			if m.Content == "!ce" {
				//Add a player only if it doesn't exist in
				if indexOf(entries, m.Author.ID) == -1 {
					initialChannel = m.ChannelID

					//users = append(users, m.Author.ID)
					entries = append(entries, entry{-1, m.Author.ID, ""})
					s.ChannelMessageSend(initialChannel, fmt.Sprintf("Inscrit ! Il y a %d joueur(s). %d joueurs max", len(entries), MAXPLAYER))
				}

				if len(entries) == MAXPLAYER {
					startAsking(s)
				}
			}

			if m.Content == "!cs" {
				startAsking(s)
			}
		}
	}
}

func fillUsers() {
	missing := MAXPLAYER - len(entries)
	fmt.Println(missing)
	for i := 0; i < missing; i++ {
		index := rand.Intn(len(entries))
		user := entries[index]
		entries = append(entries, user)
	}
}

func startAsking(s *discordgo.Session) {
	rand.Seed(time.Now().UnixNano())

	if len(entries) < MAXPLAYER {
		fillUsers()
	}

	for i := len(entries) - 1; i > 0; i-- { // Fisher–Yates shuffle
		j := rand.Intn(i + 1)
		entries[i], entries[j] = entries[j], entries[i]
	}

	for index, user := range entries {
		dm, err := s.UserChannelCreate(user.userID)

		entries[index].chunck = chunck

		if err != nil {
			fmt.Println("Error whle creating DM channel, ", err)
			return
		}

		s.ChannelMessageSend(dm.ID, "Il me faudrait un "+chunck.String())
		chunck++
	}
}
