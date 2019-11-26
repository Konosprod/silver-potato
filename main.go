package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Chunck int

const (
	SUJET = iota
	VERBE
	COMPLEMENT
	LAST
)

var (
	Token string
)

const (
	MINPLAYER = 3
)

var chunck Chunck = SUJET
var answers []string = make([]string, LAST, LAST)
var users []string = make([]string, 0, LAST)
var initialChannel string = ""

func (c Chunck) String() string {
	return [...]string{"sujet", "verbe", "complément"}[c]
}

func reset() {
	chunck = SUJET
	answers = make([]string, LAST, LAST)
	users = make([]string, 0, LAST)
	initialChannel = ""
}

func contains(strings []string, s string) bool {
	for _, v := range strings {
		if v == s {
			return true
		}
	}

	return false
}

func indexOf(strings []string, s string) int {
	for i, v := range strings {
		if v == s {
			return i
		}
	}

	return -1
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

			index := indexOf(users, m.Author.ID)

			if index != -1 {
				answers[index] = m.Content
				s.ChannelMessageSend(m.ChannelID, "Merci !")
			}

			done := true
			for _, a := range answers {
				if a == "" {
					done = false
				}
			}

			if done {
				s.ChannelMessageSend(initialChannel, strings.Join(answers, ", "))
				reset()
			}

		} else {

			if m.Content == "!ce" {
				//Add a player only if it doesn't exist in
				if !contains(users, m.Author.ID) {
					initialChannel = m.ChannelID
					users = append(users, m.Author.ID)
					s.ChannelMessageSend(initialChannel, fmt.Sprintf("Inscrit ! Il y a %d/%d joueur(s)", len(users), MINPLAYER))
				}

				if len(users) == MINPLAYER {
					startAsking(s)
				}
			}
		}
	}
}

func startAsking(s *discordgo.Session) {
	rand.Seed(time.Now().UnixNano())

	for i := len(users) - 1; i > 0; i-- { // Fisher–Yates shuffle
		j := rand.Intn(i + 1)
		users[i], users[j] = users[j], users[i]
	}

	for _, user := range users {
		dm, err := s.UserChannelCreate(user)

		if err != nil {
			fmt.Println("Error whle creating DM channel, ", err)
			return
		}

		s.ChannelMessageSend(dm.ID, "Il me faudrait un "+chunck.String())
		chunck++
	}
}
