package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

// slack_channels = web.get('https://slack.com/api/channels.list?token=' + api_key + '&exclude_archived=1&pretty=1').json()
// slack_users = web.get('https://slack.com/api/users.list?token=' + api_key + '&pretty=1').json()
// slack_groups = web.get('https://slack.com/api/groups.list?token=' + api_key + '&pretty=1').json()

var client = &http.Client{}

// Session represents an active connection to the Slack REST API.
type Session struct {
	APIToken string
}

// Presence is a possible user presence state
type Presence string

const (
	// PresenceAuto is the 'auto' presence state
	PresenceAuto Presence = "auto"

	// PresenceAway is the 'away' presence state
	PresenceAway Presence = "away"
)

const SlackAPI = "https://api.slack.com/api/"

// Auth represents slack authentication info
type Auth struct {
	Ok     bool   `json:"ok"`
	URL    string `json:"url"`
	Team   string `json:"team"`
	User   string `json:"user"`
	TeamID string `json:"team_id"`
	UserID string `json:"user_id"`
}

// Channel represents a channel
type Channel struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Members []string `json:"members"`
	Topic   struct {
		Value   string `json:"value"`
		Creator string `json:"creator"`
	} `json:"topic"`
	Purpose struct {
		Value   string `json:"value"`
		Creator string `json:"creator"`
	} `json:"topic"`
}

// PinnedMessage is a pinned message
type PinnedMessage struct {
	Ts        string   `json:"ts"`
	Type      string   `json:"type"`
	Permalink string   `json:"permalink"`
	User      string   `json:"user"`
	Text      string   `json:"text"`
	PinnedTo  []string `json:"pinned_to"`
}

// PinnedFile is a pinned file
type PinnedFile struct {
	Channels   []string `json:"channels"`
	FileType   string   `json:"filetype"`
	MimeType   string   `json:"mimetype"`
	Permalink  string   `json:"permalink"`
	PrettyType string   `json:"pretty_type"`
	Thumb64    string   `json:"thumb_64"`
	Title      string   `json:"title"`
	PrivateURL string   `json:"url_private"`
}

// Pin represents a pinned item
type Pin struct {
	Channel string         `json:"channel"`
	Created int64          `json:"created"`
	Message *PinnedMessage `json:"message,omitempty"`
	File    *PinnedFile    `json:"file,omitempty"`
}

// User is a Slack user
type User struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Deleted bool   `json:"deleted"`
	Profile struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		RealName  string `json:"real_name"`
		Email     string `json:"email"`
	} `json:"profile"`
	Presence string `json:"presence"`
}

// Title returns the title of a Pin
func (p *Pin) Title() string {
	if p.Message != nil {
		return p.Message.Text
	}
	return p.File.Title
}

// OpenSession opens a session using an existing API token.
func OpenSession(token string) Session {
	return Session{APIToken: token}
}

// GetAuth returns auth data
func (session *Session) GetAuth() (Auth, error) {
	params := map[string]string{"token": session.APIToken}

	data, err := session.get(SlackAPI, "auth.test", params)
	if err != nil {
		return Auth{}, err
	}

	var response Auth
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&response); err != nil {
		return Auth{}, err
	}

	return response, nil
}

// GetChannels returns a list of channels
func (session *Session) GetChannels() (channels []Channel, err error) {
	params := map[string]string{
		"token":            session.APIToken,
		"exclude_archived": "1",
	}

	var data []byte
	if data, err = session.get(SlackAPI, "channels.list", params); err != nil {
		return
	}

	var response struct {
		Ok       bool      `json:"ok"`
		Channels []Channel `json:"channels"`
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	if err = dec.Decode(&response); err != nil {
		return
	}

	return response.Channels, nil
}

// GetPins for a channel
func (session *Session) GetPins(channelID string) (pins []Pin, err error) {
	params := map[string]string{
		"token":   session.APIToken,
		"channel": channelID,
	}

	var data []byte
	if data, err = session.get(SlackAPI, "pins.list", params); err != nil {
		return
	}

	var response struct {
		Ok    bool  `json:"ok"`
		Items []Pin `json:"items"`
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	if err = dec.Decode(&response); err != nil {
		return
	}

	for _, item := range response.Items {
		pins = append(pins, item)
	}

	return
}

// GetUsers for a channel
func (session *Session) GetUsers() (users []User, err error) {
	params := map[string]string{
		"token": session.APIToken,
	}

	var data []byte
	if data, err = session.get(SlackAPI, "users.list", params); err != nil {
		return
	}

	var response struct {
		Ok    bool   `json:"ok"`
		Users []User `json:"members"`
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	if err = dec.Decode(&response); err != nil {
		return
	}

	return response.Users, nil
}

// GetPresence returns the presence status of a given user
func (session *Session) GetPresence(userID string) (presence string, err error) {
	params := map[string]string{
		"token": session.APIToken,
		"user":  userID,
	}

	var data []byte
	if data, err = session.get(SlackAPI, "users.getPresence", params); err != nil {
		return
	}

	var response struct {
		Ok       bool   `json:"ok"`
		Presence string `json:"presence"`
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	if err = dec.Decode(&response); err != nil {
		return
	}

	return response.Presence, nil
}

// SetPresence updates the presence of the authenticated user
func (session *Session) SetPresence(presence Presence) (err error) {
	params := map[string]string{
		"token":    session.APIToken,
		"presence": string(presence),
	}

	var data []byte
	if data, err = session.get(SlackAPI, "users.setPresence", params); err != nil {
		return
	}

	dlog.Printf("response: %s", data)

	var response struct {
		Ok    bool   `json:"ok"`
		Error string `json:"error"`
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	if err = dec.Decode(&response); err != nil {
		return
	}

	if !response.Ok {
		return fmt.Errorf("Unable to set presence: %s", response.Error)
	}

	return
}

// OpenDirectMessage opens a direct message channel and returns the channel ID
func (session *Session) OpenDirectMessage(userID string) (channelID string, err error) {
	params := map[string]string{
		"token": session.APIToken,
		"user":  userID,
	}

	var data []byte
	if data, err = session.get(SlackAPI, "im.open", params); err != nil {
		return
	}

	var response struct {
		Ok      bool `json:"ok"`
		Channel struct {
			ID string `json:"id"`
		} `json:"channel"`
		Error string `json:"error"`
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	if err = dec.Decode(&response); err != nil {
		return
	}

	if !response.Ok {
		return "", fmt.Errorf("Unable to set presence: %s", response.Error)
	}

	return response.Channel.ID, nil
}

func (session *Session) request(method string, requestURL string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, requestURL, body)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return content, fmt.Errorf(resp.Status)
	}

	return content, nil
}

func (session *Session) get(requestURL string, path string, params map[string]string) ([]byte, error) {
	requestURL += path

	if params != nil {
		data := url.Values{}
		for key, value := range params {
			data.Set(key, value)
		}
		requestURL += "?" + data.Encode()
	}

	dlog.Printf("GETing from URL: %s", requestURL)
	return session.request("GET", requestURL, nil)
}

func (session *Session) post(requestURL string, path string, data interface{}) ([]byte, error) {
	requestURL += path

	var body []byte
	var err error

	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}

	dlog.Printf("POSTing to URL: %s", requestURL)
	dlog.Printf("data: %s", body)
	return session.request("POST", requestURL, bytes.NewBuffer(body))
}
