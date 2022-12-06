package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type HVACSet struct {
	Param string `json:"param"`
	Value string `json:"value"`
}

type HVACStatus struct {
	Device struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		FamilyID       int    `json:"familyId"`
		ModelID        int    `json:"modelId"`
		InstallationID int    `json:"installationId"`
		ZoneID         int    `json:"zoneId"`
		Order          int    `json:"order"`
		Widgets        []int  `json:"widgets"`
	} `json:"device"`
	Status struct {
		Num187           int `json:"187"`
		Num188           int `json:"188"`
		Num189           int `json:"189"`
		Num190           int `json:"190"`
		Num50008         int `json:"50008"`
		Num50010         int `json:"50010"`
		AlarmStatus      int `json:"alarm_status"`
		ConfigConfirmOff int `json:"config_confirm_off"`
		ConfigFanMap     struct {
			Num0 string `json:"0"`
			Num1 string `json:"1"`
			Num2 string `json:"2"`
			Num3 string `json:"3"`
			Num4 string `json:"4"`
		} `json:"config_fan_map"`
		ConfigModeMap             int    `json:"config_mode_map"`
		ConfigQuiet               int    `json:"config_quiet"`
		ConfigVerticalVanes       int    `json:"config_vertical_vanes"`
		CoolTemperatureMax        int    `json:"cool_temperature_max"`
		CoolTemperatureMin        int    `json:"cool_temperature_min"`
		ErrorAddress              int    `json:"error_address"`
		ErrorCode                 int    `json:"error_code"`
		ExternalLed               string `json:"external_led"`
		FanSpeed                  int    `json:"fan_speed"`
		FilterClean               int    `json:"filter_clean"`
		FilterDueHours            int    `json:"filter_due_hours"`
		HeatTemperatureMin        int    `json:"heat_temperature_min"`
		InternalLed               string `json:"internal_led"`
		InternalTemperatureOffset int    `json:"internal_temperature_offset"`
		MainenanceWReset          int    `json:"mainenance_w_reset"`
		MainenanceWoReset         int    `json:"mainenance_wo_reset"`
		Mode                      string `json:"mode"`
		Power                     string `json:"power"`
		QuietMode                 string `json:"quiet_mode"`
		RemoteControllerLock      int    `json:"remote_controller_lock"`
		Rssi                      int    `json:"rssi"`
		RuntimeModeRestrictions   int    `json:"runtime_mode_restrictions"`
		Setpoint                  int    `json:"setpoint"`
		SetpointMax               int    `json:"setpoint_max"`
		SetpointMin               int    `json:"setpoint_min"`
		TempLimitation            string `json:"temp_limitation"`
		Temperature               int    `json:"temperature"`
		UID185                    int    `json:"uid_185"`
		UID186                    int    `json:"uid_186"`
		Vvane                     string `json:"vvane"`
		WorkingHours              int    `json:"working_hours"`
	} `json:"status"`
}

// TODO: name me something else

type Slack struct {
	client  *slack.Client
	socket  *socketmode.Client
	channel *slack.Channel
	intesis string
	device  string
}

var (
	device   string
	intesis  string
	logger   *log.Logger
	endpoint string
)

const (
	errGetArgs   string = "expected a key of some kind, something like ...\n`@hvac get power`"
	errSetArgs   string = "expected a key & a value, something like ...\n`@hvac set power on`"
	errCmdArgs   string = ":shrug: no command specified"
	errNotImpl   string = ":construction: not implemented"
	msgHelp      string = "I'm expecting something like\n`@hvac (help|status|set|get) key [value]`"
	msgYes       string = "Yes, I'm listening"
	healthListen string = "0.0.0.0:2113"
)

func New(botToken, appToken, channel, i, d string) (*Slack, error) {
	s := &Slack{}
	intesis = i
	device = d
	logger = log.New(os.Stdout, "", log.Lshortfile|log.LstdFlags)
	s.client = slack.New(
		botToken,
		slack.OptionAppLevelToken(appToken),
		slack.OptionLog(logger),
	)
	s.socket = socketmode.New(
		s.client,
		socketmode.OptionDebug(false),
		socketmode.OptionLog(logger),
	)
	c, err := s.GetChannelByName(channel)
	if err != nil {
		return nil, err
	}
	s.channel = c
	err = s.JoinChannel()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s Slack) PostMessage(m string) error {
	_, _, err := s.client.PostMessage(s.channel.ID, slack.MsgOptionText(m, false))
	if err != nil {
		return err
	}
	return nil
}

func (s Slack) GetChannelByName(channel string) (*slack.Channel, error) {
	p := slack.GetConversationsParameters{ExcludeArchived: true}
	channels, _, err := s.client.GetConversations(&p)
	if err != nil {
		return &slack.Channel{}, err
	}
	for _, c := range channels {
		if c.Name == channel {
			logger.Printf("resolved channel: %s to: %s", channel, c.ID)
			return &c, nil
		}
	}
	logger.Printf("unable to resolve channel: %s", channel)
	return &slack.Channel{}, fmt.Errorf("channel_not_found")
}

func (s Slack) JoinChannel() error {
	_, _, _, err := s.client.JoinConversation(s.channel.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s Slack) ListenForMention() {
	go func() {
		gin.SetMode(gin.ReleaseMode)
		router := gin.Default()
		router.GET("/health", healthHandler)
		log.Fatal(router.Run(healthListen))
	}()
	handler := socketmode.NewSocketmodeHandler(s.socket)
	handler.Handle(socketmode.EventTypeConnecting, middlewareConnecting)
	handler.Handle(socketmode.EventTypeConnectionError, middlewareConnectionError)
	handler.Handle(socketmode.EventTypeConnected, middlewareConnected)
	handler.HandleEvents(slackevents.AppMention, middlewareAppMentionEvent)
	handler.RunEventLoop()
}

func healthHandler(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func middlewareConnecting(evt *socketmode.Event, client *socketmode.Client) {
	logger.Print("socketmode connectng")
}

func middlewareConnectionError(evt *socketmode.Event, client *socketmode.Client) {
	logger.Print("socketmode connection error")
}

func middlewareConnected(evt *socketmode.Event, client *socketmode.Client) {
	logger.Print("socketmode connected")
}

func middlewareAppMentionEvent(evt *socketmode.Event, client *socketmode.Client) {
	logger.Print("socketmode AppMentionEvent")
	eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
	if !ok {
		logger.Printf("WARN: ignored %+v", evt)
		return
	}

	client.Ack(*evt.Request)

	ev, ok := eventsAPIEvent.InnerEvent.Data.(*slackevents.AppMentionEvent)
	if !ok {
		logger.Printf("WARN: ignored %+v", ev)
		return
	}

	logger.Printf(`received mention with text: "%s" in channel: %v`, ev.Text, ev.Channel)
	mention := strings.Split(ev.Text, " ")
	if len(mention) < 2 {
		threadedReply(client, ev.Channel, ev.EventTimeStamp, errCmdArgs)
		threadedReply(client, ev.Channel, ev.EventTimeStamp, msgHelp)
		return
	}
	switch mention[1] {
	case "hi", "hello", "ping":
		threadedReply(client, ev.Channel, ev.EventTimeStamp, msgYes)
	case "status":
		statusHandler(client, ev.Channel, ev.EventTimeStamp)
	case "get":
		if len(mention) < 3 {
			threadedReply(client, ev.Channel, ev.EventTimeStamp, errGetArgs)
			return
		}
		threadedReply(client, ev.Channel, ev.EventTimeStamp, errNotImpl)
	case "set":
		if len(mention) < 4 {
			threadedReply(client, ev.Channel, ev.EventTimeStamp, errSetArgs)
			return
		}
		setHandler(client, ev.Channel, ev.EventTimeStamp, mention[2], mention[3])
	default:
		threadedReply(client, ev.Channel, ev.EventTimeStamp, msgHelp)
	}
}

func threadedReply(client *socketmode.Client, channel, thread, message string) {
	_, _, err := client.Client.PostMessage(
		channel,
		slack.MsgOptionTS(thread),
		slack.MsgOptionText(message, false),
	)
	if err != nil {
		logger.Printf("ERROR: problem sending message: %s with: %s", message, err)
		return
	}
}

func blockReply(client *socketmode.Client, channel, thread string, opt slack.MsgOption) {
	_, _, err := client.Client.PostMessage(
		channel,
		slack.MsgOptionTS(thread),
		opt,
	)
	if err != nil {
		logger.Printf("ERROR: problem sending block: %s", err)
		return
	}
}

// TODO: turn the output into a struct with a toString method?
func statusHandler(client *socketmode.Client, channel, thread string) {
	endpoint := fmt.Sprint(intesis + "/hvac/" + device)
	resp, err := http.Get(endpoint)
	if err != nil {
		logger.Printf("ERROR: calling intesis: %s with: %s", endpoint, err)
		m := fmt.Sprintf(":x: error contacting intesis: `%s`", err)
		threadedReply(client, channel, thread, m)
		return
	}
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		logger.Printf("ERROR: calling intesis: %s with: %s", endpoint, string(body))
		m := fmt.Sprintf(":x: error contacting intesis: `%s`", string(body))
		threadedReply(client, channel, thread, m)
		return
	}
	status := &HVACStatus{}
	if err = json.Unmarshal(body, &status); err != nil {
		logger.Printf("ERROR: unable to decode: %s ", string(body))
		m := fmt.Sprintf(":x: unable to decode: `%s`", string(body))
		threadedReply(client, channel, thread, m)
		return
	}
	blockReply(client, channel, thread, status.blockOutput())
}

// TODO: consolidate this
func setHandler(client *socketmode.Client, channel, thread, param, value string) {
	endpoint := fmt.Sprint(intesis + "/hvac/" + device)
	payload := &HVACSet{Param: param, Value: value}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(payload)
	if err != nil {
		logger.Printf("ERROR: encoding command: %s %s to json: %s", payload.Param, payload.Value, err)
		m := fmt.Sprintf(":x: error encoding command: `%s`", err)
		threadedReply(client, channel, thread, m)
		return
	}
	resp, err := http.Post(endpoint, "application/json", &buf)
	if err != nil {
		logger.Printf("ERROR: calling intesis: %s with: %s", endpoint, err)
		m := fmt.Sprintf(":x: error contacting intesis: `%s`", err)
		threadedReply(client, channel, thread, m)
		return
	}
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusAccepted {
		logger.Printf("ERROR: calling intesis: %s with: %s", endpoint, string(body))
		m := fmt.Sprintf(":x: error contacting intesis: `%s`", string(body))
		threadedReply(client, channel, thread, m)
		return
	}
	m := fmt.Sprintf(":+1: `%s`", string(body))
	threadedReply(client, channel, thread, m)
}

func (h *HVACStatus) blockOutput() slack.MsgOption {
	fields := []*slack.TextBlockObject{}
	title := &slack.TextBlockObject{Type: slack.MarkdownType, Text: "*Device:* " + h.Device.ID}
	fields = append(fields, &slack.TextBlockObject{Type: slack.MarkdownType, Text: fmt.Sprintf("*Power:* %s", h.Status.Power)})
	fields = append(fields, &slack.TextBlockObject{Type: slack.MarkdownType, Text: fmt.Sprintf("*Mode:* %s", h.Status.Mode)})
	fields = append(fields, &slack.TextBlockObject{Type: slack.MarkdownType, Text: fmt.Sprintf("*Fan Speed:* %d", h.Status.FanSpeed)})
	fields = append(fields, &slack.TextBlockObject{Type: slack.MarkdownType, Text: fmt.Sprintf("*Set Point:* %.2fC", float64(h.Status.Setpoint/10))})
	fields = append(fields, &slack.TextBlockObject{Type: slack.MarkdownType, Text: fmt.Sprintf("*Room Temp:* %.2fC", float64(h.Status.Temperature/10))})
	section := slack.NewSectionBlock(title, fields, nil)
	return slack.MsgOptionBlocks(section)
}
