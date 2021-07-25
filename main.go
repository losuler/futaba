package main

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"
)

type Config struct {
	Discord Discord `yaml:"discord"`
	Roles   Roles   `yaml:"roles"`
	Users   []Users `yaml:"users"`
}

type Roles struct {
	MuteID string `yaml:"muteid"`
}

type Discord struct {
	Token string `yaml:"token"`
}

type Users struct {
	Username  string `yaml:"username"`
	UserID    string `yaml:"userid"`
	Timezone  string `yaml:"timezone"`
	Nicknames string `yaml:"nicknames"`
	Admin     bool   `yaml:"admin"`
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
		if strings.ToLower(suffix) == strings.ToLower(user.Username) {
			return user, user.Username, nil
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
	log.Printf("[INFO] Command match for user \"%s\".\n", suffix[2])

	// [0] = whole match, [1] = command, [2] = username
	account, userName, err := getAcc(conf, suffix[2])
	if err != nil {
		return
	}

	dayTime := getTime(account)

	msg := fmt.Sprintf("It's %s where %s is.",
		dayTime, userName)

	_, err = s.ChannelMessageSend(m.ChannelID, msg)
	if err != nil {
		fmt.Println(err)
	}
}

func memberInConf(member *discordgo.Member, conf Config) bool {
	for _, confUser := range conf.Users {
		if member.User.ID == confUser.UserID {
			return true
		}
	}
	return false
}

func userIsAdmin(user *discordgo.User, conf Config) bool {
	for _, confUser := range conf.Users {
		if user.ID == confUser.UserID && confUser.Admin == true {
			return true
		}
	}
	return false
}

func updateConfig(conf Config, members []*discordgo.Member) {
	for _, member := range members {
		if !memberInConf(member, conf) {
			var confUser Users

			confUser.Username = member.User.Username
			confUser.UserID = member.User.ID
			confUser.Timezone = ""
			confUser.Nicknames = member.Nick
			confUser.Admin = false

			conf.Users = append(conf.Users, confUser)
		}
	}

	data, err := yaml.Marshal(&conf)
	if err != nil {
		log.Println("[ERROR] ", err)
	}

	err = ioutil.WriteFile("/etc/futaba.yml", data, 0)
	if err != nil {
		log.Println("[ERROR] ", err)
	}
}

func sendUpdate(conf Config, cmd *regexp.Regexp,
	s *discordgo.Session, m *discordgo.MessageCreate) {
	// Check if allowed to execute command
	if !userIsAdmin(m.Author, conf) {
		return
	}

	guildsList := s.State.Guilds
	log.Printf("[INFO] Number of users in the guild: %d.\n", guildsList[0].MemberCount)
	log.Printf("[INFO] Usernames: %d.\n", len(guildsList[0].Members))

	for _, member := range guildsList[0].Members {
		log.Printf("[INFO] Username: %s.\n", member.User.Username)
	}

	updateConfig(conf, guildsList[0].Members)

	// TODO: Include details about how many updated, etc.
	msg := fmt.Sprintf("Updated users list.")

	_, err := s.ChannelMessageSend(m.ChannelID, msg)
	if err != nil {
		fmt.Println(err)
	}
}

func setMute(addRole bool, conf Config, cmd *regexp.Regexp,
	s *discordgo.Session, m *discordgo.MessageCreate) {
	// Check if allowed to execute command
	if !userIsAdmin(m.Author, conf) {
		log.Printf("[WARN] %s is not allowed to mute users.\n", m.Author.Username)
		return
	}

	suffix := cmd.FindStringSubmatch(m.Content)
	log.Printf("[INFO] Mute command match for user \"%s\".\n", suffix[2])

	// [0] = whole match, [1] = command, [2] = username
	account, userName, err := getAcc(conf, suffix[2])
	if err != nil {
		log.Println("[ERROR] ", err)
		return
	}

	if addRole {
		err = s.GuildMemberRoleAdd(s.State.Guilds[0].ID, account.UserID, conf.Roles.MuteID)
		if err != nil {
			log.Println("[ERROR]", err)
			return
		}

		log.Printf("[INFO] %s has been muted.\n", userName)
	} else {
		err = s.GuildMemberRoleRemove(s.State.Guilds[0].ID, account.UserID, conf.Roles.MuteID)
		if err != nil {
			log.Println("[ERROR]", err)
			return
		}

		log.Printf("[INFO] %s has been un-muted.\n", userName)
	}
}

func messageRecieve(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Check the message is not from the bot.
	if m.Author.ID == s.State.User.ID {
		return
	}

	conf := readConfig("/etc/futaba.yml")

	// Regexp for each command.
	timeUpdate := regexp.MustCompile(`^(t|time)\.(update)`)
	timeCheck := regexp.MustCompile(`^(t|time)\.(.+)`)
	muteUser := regexp.MustCompile(`^(m|mute)\.(.+)`)
	unmuteUser := regexp.MustCompile(`^(\!m|unmute)\.(.+)`)

	switch {
	case timeUpdate.MatchString(m.Content):
		sendUpdate(conf, timeCheck, s, m)
	case timeCheck.MatchString(m.Content):
		sendTime(conf, timeCheck, s, m)
	case muteUser.MatchString(m.Content):
		setMute(true, conf, muteUser, s, m)
	case unmuteUser.MatchString(m.Content):
		setMute(false, conf, unmuteUser, s, m)
	}
}

func main() {
	c := readConfig("/etc/futaba.yml")

	dg, err := discordgo.New("Bot " + c.Discord.Token)
	if err != nil {
		panic(err)
	}

	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)

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
