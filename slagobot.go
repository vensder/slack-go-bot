package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/slack-go/slack"
	"gopkg.in/yaml.v2"
)

const (
	templateReport = `
_My outbound IP is_: *%s*
_My external IP is_: *%s*
_My hostname is_: *%s*
_My latency is_: *%s*
_Runtime OS/Arch_: *%s*`
)

type conf struct {
	Admins  []string `yaml:"admins"`
	Admin   string   `yaml:"admin"`
	Channel string   `yaml:"channel"`
}

func (c *conf) getConf(configPath string) *conf {
	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Printf("Read config error: %v\n", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		// log.Fatalf("Unmarshal: %v", err)
		log.Printf("Unmarshal error: %v\n", err)
	}

	return c
}

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Printf("Getting outbound address error: %v\n", err)
		return "Can't obtain the outbound IP address"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return fmt.Sprintf("%v", localAddr.IP)
}

func getExternalIP() string {
	resp, err := http.Get("http://checkip.amazonaws.com")
	if err != nil {
		log.Printf("Error during the get request to checkip.amazonaws.com: %v\n", err)
		return "Can't obtain the external IP address"
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error during the response body reading: %v\n", err)
		return "Can't read response body"
	}
	log.Println(string(body))
	return strings.TrimSpace(string(body))
}

func getHostname() string {
	name, err := os.Hostname()
	if err != nil {
		log.Printf("Getting hostname error: %v\n", err)
		return "can't obtain hostname"
	}
	return fmt.Sprintf("%v", name)
}

func getOsArch() string {
	return fmt.Sprintf("%v/%v", runtime.GOOS, runtime.GOARCH)
}

func main() {
	configPathPtr := flag.String("config-path", "config.yaml", "path to the config file")
	slackTokenPtr := flag.String("slack-token", "xoxb", "slack bot token")
	flag.Parse()
	var slackToken string
	var defaultChannelName string = "random"
	var defaultChannelID string
	var configuration conf
	// var currentLatency time.Duration
	var currentLatencyStr string = "Not checked yet"
	channelChIDMap := make(map[string]string)
	chIDChannelMap := make(map[string]string)
	sendMessageAfterTypingMap := make(map[string]bool)

	configuration.getConf(*configPathPtr)
	outboundIP := getOutboundIP()
	externalIP := getExternalIP()
	hostname := getHostname()
	osNameArch := getOsArch()
	fmt.Printf("Outbound IP: %s\n", outboundIP)
	fmt.Printf("External IP: %s\n", externalIP)
	fmt.Printf("Hostname: %s\n", hostname)

	if *slackTokenPtr == "xoxb" {
		fmt.Println("slack-token flag not passed")
		fmt.Println("checking environment variable...")
		var ok bool
		slackToken, ok = os.LookupEnv("SLACK_TOKEN")
		if !ok {
			fmt.Println("SLACK_TOKEN environment variable is not set too")
			os.Exit(1)

		}
	} else {
		slackToken = *slackTokenPtr
	}

	api := slack.New(
		slackToken,
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)),
	)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	channels, err := api.GetChannels(false)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	for _, channel := range channels {
		fmt.Println(channel.Name, channel.ID)
		channelChIDMap[channel.Name] = channel.ID
		chIDChannelMap[channel.ID] = channel.Name
	}
	fmt.Println("channelChIDMap:", channelChIDMap)
	fmt.Println("chIDChannelMap:", chIDChannelMap)

	if configuration.Channel != "" {
		defaultChannelName = configuration.Channel
	}

	defaultChannelID = channelChIDMap[defaultChannelName]

	fmt.Printf("Admin: %v, Default channel: %v, Admins: %v\n",
		configuration.Admin,
		configuration.Channel,
		configuration.Admins,
	)

	users, err := api.GetUsers()
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	for _, user := range users {
		fmt.Println(user.Name, user.RealName, user.ID)
	}

	for msg := range rtm.IncomingEvents {
		fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			fmt.Printf("Hello event: %v\n", ev)
			rtm.SendMessage(rtm.NewOutgoingMessage(
				fmt.Sprintf(templateReport,
					outboundIP,
					externalIP,
					hostname,
					currentLatencyStr,
					osNameArch),
				defaultChannelID))

		case *slack.ConnectedEvent:
			fmt.Println("Infos:", ev.Info)
			fmt.Println("Connection counter:", ev.ConnectionCount)
			rtm.SendMessage(rtm.NewOutgoingMessage("Hi, I'm connected!", defaultChannelID))

		case *slack.MessageEvent:
			fmt.Printf("Message: %v\n", ev)
			fmt.Println("Msg:", ev.Msg)
			fmt.Println("Msg.User:", ev.Msg.User)
			fmt.Println("Text:", ev.Text)

			sendMessageAfterTypingMap[ev.User] = true

			if ev.Msg.User == configuration.Admin && ev.Text == "!ip" {
				rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf("My ip: %s. Latency: %v",
					outboundIP, currentLatencyStr),
					defaultChannelID))
			}
			if ev.Msg.User == configuration.Admin && ev.Text == "!report" {
				rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf(templateReport,
					outboundIP,
					externalIP,
					hostname,
					currentLatencyStr,
					osNameArch),
					defaultChannelID))
			}
			if strings.HasPrefix(ev.Text, "!tr ") {
				rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf("Translation of *%s*: xxxxx",
					strings.Replace(ev.Text, "!tr ", "", 1)),
					defaultChannelID))
			}

		case *slack.PresenceChangeEvent:
			fmt.Printf("Presence Change: %v\n", ev)

		case *slack.LatencyReport:
			currentLatencyStr = fmt.Sprintf("%v", ev.Value)
			fmt.Printf("Current latency: %v\n", currentLatencyStr)

		case *slack.DesktopNotificationEvent:
			fmt.Printf("Desktop Notification: %v\n", ev)

		case *slack.RTMError:
			fmt.Printf("RTM Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return

		case *slack.UserTypingEvent:
			typingUserInfo, err := api.GetUserInfo(ev.User)
			if err != nil {
				fmt.Printf("%s\n", err)
			}
			typingUserRealName := typingUserInfo.RealName
			typingUserChannel := ev.Channel

			fmt.Printf("User typing event: %v, User: %v, Real Name: %v\n",
				ev,
				ev.User,
				typingUserRealName,
			)

			if sendMessageAfterTyping, ok := sendMessageAfterTypingMap[ev.User]; ok {
				if sendMessageAfterTyping == true {
					rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf("Wow! %v is typing! Say something wisdom!",
						typingUserRealName),
						typingUserChannel))
					sendMessageAfterTypingMap[ev.User] = false
				}
			} else {
				sendMessageAfterTypingMap[ev.User] = true
			}

		default:
			// Other events..
			fmt.Printf("Unexpected event %v with content: %v\n", ev, msg.Data)
		}
	}
}
