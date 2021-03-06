package main

import (
	"fmt"
	"time"
	"os"
	"os/signal"
	"syscall"
	"io/ioutil"
	"regexp"
	"strings"
	"errors"
	"github.com/go-yaml/yaml"
	"github.com/bwmarrin/discordgo"
)

type Config struct {
	Discord Discord `yaml:"discord"`
	Users []Users `yaml:"users"`
}

type Discord struct {
	Token string `yaml:"token"`
}

type Users struct {
	Username string `yaml:"username"`
	UserID string `yaml:"userid"`
	Timezone string	`yaml:"timezone"`
	Nicknames string `yaml:"nicknames"`
	Commands string `yaml:"commands"`
}

func readConfig(configFile string) Config {
	var c Config
    
	raw, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(err)
	} 

	err = yaml.Unmarshal([]byte(raw), &c)

	return c
}

func getAcc(c Config, suffix string) (Users, string, error) {
	var user Users

	for _, user := range c.Users {
		if strings.ToLower(suffix) == strings.ToLower(user.Commands) {
			return user, user.Username, nil
		} else if strings.ToLower(suffix) == strings.ToLower(user.Username) {
			return user, user.Username, nil
		// TODO: Check for more than one nickname.
		} else if strings.ToLower(suffix) == strings.ToLower(user.Nicknames) {
			return user, user.Username, nil
		}
	}

	return user, suffix, errors.New("Account doesn't exist.")
}

func getTime(account Users) string {
	tz, err := time.LoadLocation(account.Timezone)
	if err != nil {
		fmt.Println(err)
	}

	timeNow := time.Now().In(tz)
	dayTime := timeNow.Format("Monday 3:04PM")

	return dayTime
}

func sendTime(conf Config, cmd *regexp.Regexp, 
	s *discordgo.Session, m *discordgo.MessageCreate) {

	suffix := cmd.FindStringSubmatch(m.Content)

	// [1] is the first group in ().
	account, userName, err := getAcc(conf, suffix[1])
	if err != nil {
		return
	}

	dayTime := getTime(account)

	msg := fmt.Sprintf("It's %s where %s is.",
		dayTime, strings.Title(userName))

	_, err = s.ChannelMessageSend(m.ChannelID, msg)
	if err != nil {
		fmt.Println(err)
	}
}

func messageRecieve(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Check the message is not from the bot.
	if m.Author.ID == s.State.User.ID {
		return
	}

	conf := readConfig("config.yml")

	// Regexp for each command.
	timeFull := regexp.MustCompile(`time\.(.+)`)
	timePart := regexp.MustCompile(`t\.(.+)`)

	switch {
		case timeFull.MatchString(m.Content):
			sendTime(conf, timeFull, s, m)
		case timePart.MatchString(m.Content):
			sendTime(conf, timePart, s, m)
	}
}

func main() {
	c := readConfig("/etc/futaba.yml")

	dg, err := discordgo.New("Bot " + c.Discord.Token)
	if err != nil {
		panic(err)
	}

	err = dg.Open()
	if err != nil {
		panic(err)
	}

	dg.AddHandler(messageRecieve)

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sigChan

	dg.Close()
}
