package alexa

import (
	"context"
	"errors"
	"log"
	"math"
	"strconv"
	"time"
)

const sdkVersion = "1.0"
const launchRequestName = "LaunchRequest"
const intentRequestName = "IntentRequest"
const sessionEndedRequestName = "SessionEndedRequest"

var timestampTolerance = 150

// Alexa defines the primary interface to use to create an Alexa request handler.
type Alexa struct {
	ApplicationID       string
	RequestHandler      RequestHandler
	IgnoreApplicationID bool
	IgnoreTimestamp     bool
}

// RequestHandler defines the interface that must be implemented to handle
// Alexa Requests
type RequestHandler interface {
	OnSessionStarted(context.Context, *Request, *Session, *Context, *Response) error
	OnLaunch(context.Context, *Request, *Session, *Context, *Response) error
	OnIntent(context.Context, *Request, *Session, *Context, *Response) error
	OnSessionEnded(context.Context, *Request, *Session, *Context, *Response) error
}

// RequestEnvelope contains the data passed from Alexa to the request handler.
type RequestEnvelope struct {
	Version string   `json:"version"`
	Session *Session `json:"session"`
	Request *Request `json:"request"`
	Context *Context `json:"context"`
}

// Session containes the session data from the Alexa request.
type Session struct {
	New        bool   `json:"new"`
	SessionID  string `json:"sessionId"`
	Attributes struct {
		String map[string]interface{} `json:"string"`
	} `json:"attributes"`
	User struct {
		UserID      string `json:"userId"`
		AccessToken string `json:"accessToken"`
	} `json:"user"`
	Application struct {
		ApplicationID string `json:"applicationId"`
	} `json:"application"`
}

type Context struct {
	AudioPlayer struct {
		PlayerActivity string `json:"playerActivity"`
	} `json:"AudioPlayer"`
	Display struct {
		Token string `json:"token"`
	} `json:"Display"`
	System struct {
		Application struct {
			ApplicationID string `json:"applicationId"`
		} `json:"application"`
		User struct {
			UserID string `json:"userId"`
		} `json:"user"`
		Device struct {
			DeviceID            string `json:"deviceId"`
			SupportedInterfaces struct {
				AudioPlayer struct {
				} `json:"AudioPlayer"`
				Display struct {
					TemplateVersion string `json:"templateVersion"`
					MarkupVersion   string `json:"markupVersion"`
				} `json:"Display"`
			} `json:"supportedInterfaces"`
		} `json:"device"`
		APIEndpoint    string `json:"apiEndpoint"`
		APIAccessToken string `json:"apiAccessToken"`
	} `json:"System"`
}

// Request contines the data in the request within the main request.
type Request struct {
	Locale      string `json:"locale"`
	Timestamp   string `json:"timestamp"`
	Type        string `json:"type"`
	RequestID   string `json:"requestId"`
	DialogState string `json:"dialogState"`
	Intent      Intent `json:"intent"`
	Name        string `json:"name"`
}

// Intent contains the data about the Alexa Intent requested.
type Intent struct {
	Name               string                `json:"name"`
	ConfirmationStatus string                `json:"confirmationStatus,omitempty"`
	Slots              map[string]IntentSlot `json:"slots"`
}

// IntentSlot contains the data for one Slot
type IntentSlot struct {
	Name               string      `json:"name"`
	ConfirmationStatus string      `json:"confirmationStatus,omitempty"`
	Value              string      `json:"value"`
	ID                 string      `json:"id,omitempty"`
	Resolutions        *Resolution `json:"resolutions,omitempty"`
}

type Resolution struct {
	ResolutionsPerAuthority []Authority `json:"resolutionsPerAuthority"`
}

type Authority struct {
	Authority string            `json:"authority"`
	Status    *ResolutionStatus `json:"status"`
	Values    []ResolutionValue `json:"values"`
}

type ResolutionStatus struct {
	Code string `json:"code"`
}

type ResolutionValue struct {
	Value *SlotValue `json:"value"`
}

type SlotValue struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

// ResponseEnvelope contains the Response and additional attributes.
type ResponseEnvelope struct {
	Version           string                 `json:"version"`
	SessionAttributes map[string]interface{} `json:"sessionAttributes,omitempty"`
	Response          *Response              `json:"response"`
}

// Response contains the body of the response.
type Response struct {
	OutputSpeech     *OutputSpeech `json:"outputSpeech,omitempty"`
	Card             *Card         `json:"card,omitempty"`
	Reprompt         *Reprompt     `json:"reprompt,omitempty"`
	Directives       []interface{} `json:"directives,omitempty"`
	ShouldSessionEnd *bool         `json:"shouldEndSession,omitempty"`
}

// OutputSpeech contains the data the defines what Alexa should say to the user.
type OutputSpeech struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	SSML string `json:"ssml,omitempty"`
}

// Card contains the data displayed to the user by the Alexa app.
type Card struct {
	Type    string `json:"type"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
	Text    string `json:"text,omitempty"`
	Image   *Image `json:"image,omitempty"`
}

// Image provides URL(s) to the image to display in resposne to the request.
type Image struct {
	SmallImageURL string `json:"smallImageUrl,omitempty"`
	LargeImageURL string `json:"largeImageUrl,omitempty"`
}

// Reprompt contains data about whether Alexa should prompt the user for more data.
type Reprompt struct {
	OutputSpeech *OutputSpeech `json:"outputSpeech,omitempty"`
}

// AudioPlayerDirective contains device level instructions on how to handle the response.
type AudioPlayerDirective struct {
	Type         string     `json:"type"`
	PlayBehavior string     `json:"playBehavior,omitempty"`
	AudioItem    *AudioItem `json:"audioItem,omitempty"`
}

// AudioItem contains an audio Stream definition for playback.
type AudioItem struct {
	Stream Stream `json:"stream,omitempty"`
}

// VideoAppDirective contains device level instructions on how to handle the response.
type VideoAppDirective struct {
	Type      string     `json:"type"`
	VideoItem *VideoItem `json:"videoItem"`
}

// VideoItem contains a video file for playback
type VideoItem struct {
	Source   string    `json:"source"`
	Metadata *Metadata `json:"metadata,omitempty"`
}

// Metadata contains additional information about the VideoItem
type Metadata struct {
	Title    string `json:"title,omitempty"`
	Subtitle string `json:"subtitle,omitempty"`
}

// Stream contains instructions on playing an audio stream.
type Stream struct {
	Token                string `json:"token"`
	URL                  string `json:"url"`
	OffsetInMilliseconds int    `json:"offsetInMilliseconds"`
}

// DialogDirective contains directives for use in Dialog prompts.
type DialogDirective struct {
	Type          string  `json:"type"`
	SlotToElicit  string  `json:"slotToElicit,omitempty"`
	SlotToConfirm string  `json:"slotToConfirm,omitempty"`
	UpdatedIntent *Intent `json:"updatedIntent,omitempty"`
}

// DelegateDirective contains directives for use in delegating Dialog prompts to Alexa
type DelegateDirective struct {
	Type          string  `json:"type"`
	UpdatedIntent *Intent `json:"updatedIntent,omitempty"`
}

// DisplayDirective contains directives for use in template rendering for Echo devices with screens
type DisplayDirective struct {
	Type     string    `json:"type"`
	Template *Template `json:"template"`
}

type DisplayTemplate struct {
	Type            string              `json:"type"`
	Token           string              `json:"token"`
	BackButton      string              `json:"backButton"`
	BackgroundImage *DisplayImage       `json:"backgroundImage"`
	Title           string              `json:"title"`
	TextContent     *DisplayTextContent `json:"textContent"`
}

type DisplayImage struct {
	ContentDescription string          `json:"contentDescription"`
	Sources            []DisplaySource `json:"sources"`
}

type DisplaySource struct {
	Url          string  `json:"url"`
	Size         *string `json:"size"`
	WidthPixels  *int    `json:"widthPixels"`
	HeightPixels *int    `json:"heightPixels"`
}

type DisplayTextContent struct {
	PrimaryText struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"primaryText"`
	SecondaryText struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"secondaryText"`
	TertiaryText struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"tertiaryText"`
}

// ProcessRequest handles a request passed from Alexa
func (alexa *Alexa) ProcessRequest(ctx context.Context, requestEnv *RequestEnvelope) (*ResponseEnvelope, error) {

	if !alexa.IgnoreApplicationID {
		err := alexa.verifyApplicationID(requestEnv)
		if err != nil {
			return nil, err
		}
	}
	if !alexa.IgnoreTimestamp {
		err := alexa.verifyTimestamp(requestEnv)
		if err != nil {
			return nil, err
		}
	} else {
		log.Println("Ignoring timestamp verification.")
	}

	request := requestEnv.Request
	session := requestEnv.Session
	context := requestEnv.Context

	responseEnv := &ResponseEnvelope{}
	responseEnv.Version = sdkVersion
	responseEnv.Response = &Response{}
	//	responseEnv.Response.ShouldSessionEnd = true // Set default value.

	response := responseEnv.Response

	// If it is a new session, invoke onSessionStarted
	if session.New {
		err := alexa.RequestHandler.OnSessionStarted(ctx, request, session, context, response)
		if err != nil {
			log.Println("Error handling OnSessionStarted.", err.Error())
			return nil, err
		}
	}

	switch requestEnv.Request.Type {
	case launchRequestName:
		err := alexa.RequestHandler.OnLaunch(ctx, request, session, context, response)
		if err != nil {
			log.Println("Error handling OnLaunch.", err.Error())
			return nil, err
		}
	case intentRequestName:
		err := alexa.RequestHandler.OnIntent(ctx, request, session, context, response)
		if err != nil {
			log.Println("Error handling OnIntent.", err.Error())
			return nil, err
		}
	case sessionEndedRequestName:
		err := alexa.RequestHandler.OnSessionEnded(ctx, request, session, context, response)
		if err != nil {
			log.Println("Error handling OnSessionEnded.", err.Error())
			return nil, err
		}
	}

	return responseEnv, nil
}

// SetTimestampTolerance sets the maximum number of seconds to allow between
// the current time and the request Timestamp.  Default value is 150 seconds.
func (alexa *Alexa) SetTimestampTolerance(seconds int) {
	timestampTolerance = seconds
}

// SetSimpleCard creates a new simple card with the specified content.
func (r *Response) SetSimpleCard(title string, content string) {
	r.Card = &Card{Type: "Simple", Title: title, Content: content}
}

// SetStandardCard creates a new standard card with the specified content.
func (r *Response) SetStandardCard(title string, text string, smallImageURL string, largeImageURL string) {
	r.Card = &Card{Type: "Standard", Title: title, Text: text}
	r.Card.Image = &Image{SmallImageURL: smallImageURL, LargeImageURL: largeImageURL}
}

// SetLinkAccountCard creates a new LinkAccount card.
func (r *Response) SetLinkAccountCard() {
	r.Card = &Card{Type: "LinkAccount"}
}

// SetOutputText sets the OutputSpeech type to text and sets the value specified.
func (r *Response) SetOutputText(text string) {
	r.OutputSpeech = &OutputSpeech{Type: "PlainText", Text: text}
}

// SetOutputSSML sets the OutputSpeech type to ssml and sets the value specified.
func (r *Response) SetOutputSSML(ssml string) {
	r.OutputSpeech = &OutputSpeech{Type: "SSML", SSML: ssml}
}

// SetRepromptText created a Reprompt if needed and sets the OutputSpeech type to text and sets the value specified.
func (r *Response) SetRepromptText(text string) {
	if r.Reprompt == nil {
		r.Reprompt = &Reprompt{}
	}
	r.Reprompt.OutputSpeech = &OutputSpeech{Type: "PlainText", Text: text}
}

// SetRepromptSSML created a Reprompt if needed and sets the OutputSpeech type to ssml and sets the value specified.
func (r *Response) SetRepromptSSML(ssml string) {
	if r.Reprompt == nil {
		r.Reprompt = &Reprompt{}
	}
	r.Reprompt.OutputSpeech = &OutputSpeech{Type: "SSML", SSML: ssml}
}

// AddAudioPlayer adds an AudioPlayer directive to the Response.
func (r *Response) AddAudioPlayer(playerType, playBehavior, streamToken, url string, offsetInMilliseconds int) {
	d := AudioPlayerDirective{
		Type:         playerType,
		PlayBehavior: playBehavior,
		AudioItem: &AudioItem{
			Stream: Stream{
				Token:                streamToken,
				URL:                  url,
				OffsetInMilliseconds: offsetInMilliseconds,
			},
		},
	}
	r.Directives = append(r.Directives, d)
}

// AddVideoApp adds a VideoApp directive to the Response
func (r *Response) AddVideoApp(appType, sourceFile, title, subtitle string) {
	d := VideoAppDirective{
		Type: appType,
		VideoItem: &VideoItem{
			Source: sourceFile,
			Metadata: &Metadata{
				Title:    title,
				Subtitle: subtitle,
			},
		},
	}
	r.Directives = append(r.Directives, d)
}

// AddDialogDirective adds a Dialog directive to the Response.
func (r *Response) AddDialogDirective(dialogType, slotToElicit, slotToConfirm string, intent *Intent) {
	d := DialogDirective{
		Type:          dialogType,
		SlotToElicit:  slotToElicit,
		SlotToConfirm: slotToConfirm,
		UpdatedIntent: intent,
	}
	r.Directives = append(r.Directives, d)
}

// AddDelegateDirective adds a Delegate directive to the Response
func (r *Response) AddDelegateDirective(dialogType string, intent *Intent) {
	d := DelegateDirective{
		Type:          dialogType,
		UpdatedIntent: intent,
	}
	r.Directives = append(r.Directives, d)
}

// AddDisplayDirective adds a Display directive to the Response
func (r *Response) AddDisplayDirective(templateType, token, backButton, bkgrndImgDesc, bkgrndUrl, bkgrndSize, bkgrndWidth, bkgrndHeight, imgDesc, imgUrl, imgSize, imgWidth, imgHeight, primaryTxt, secondaryTxt, tertiaryTxt string) {
	d := DisplayDirective{
		Type: "Display.RenderTemplate",
		Template: &Template{
			Type:       templateType,
			Token:      token,
			BackButton: backButton,
			BackgroundImage: BackgroundImage{
				ContentDescription: bkgrndImgDesc,
				Sources: []DisplaySource{
					Url:          bkgrndUrl,
					Size:         bkgrndSize,
					WidthPixels:  bkgrndWidth,
					HeightPixels: bkgrndHeight,
				},
				Title: title,
				Image: &DisplayImage{
					ContentDescription: imgDesc,
					Sources: []DisplaySource{
						Url:          imgUrl,
						Size:         imgSize,
						WidthPixels:  imgWidth,
						HeightPixels: imgHeight,
					},
				},
				TextContent: &DisplayTextContent{
					PrimaryText: PrimaryText{
						Text: primaryTxt,
						Type: "PlainText",
					},
					SecondaryText: SecondaryText{
						Text: secondaryTxt,
						Type: "PlainText",
					},
					TertiaryText: TertiaryText{
						Text: tertiaryTxt,
						Type: "PlainText",
					},
				},
			},
		},
	}
	r.Directives = append(r.Directives, d)
}

// verifyApplicationId verifies that the ApplicationID sent in the request
// matches the one configured for this skill.
func (alexa *Alexa) verifyApplicationID(request *RequestEnvelope) error {
	appID := alexa.ApplicationID
	requestAppID := request.Session.Application.ApplicationID
	if appID == "" {
		return errors.New("application ID was set to an empty string")
	}
	if requestAppID == "" {
		return errors.New("request Application ID was set to an empty string")
	}
	if appID != requestAppID {
		return errors.New("request Application ID does not match expected ApplicationId")
	}

	return nil
}

// verifyTimestamp compares the request timestamp to the current timestamp
// and returns an error if they are too far apart.
func (alexa *Alexa) verifyTimestamp(request *RequestEnvelope) error {

	timestamp, err := time.Parse(time.RFC3339, request.Request.Timestamp)
	if err != nil {
		return errors.New("Unable to parse request timestamp.  Err: " + err.Error())
	}
	now := time.Now()
	delta := now.Sub(timestamp)
	deltaSecsAbs := math.Abs(delta.Seconds())
	if deltaSecsAbs > float64(timestampTolerance) {
		return errors.New("Invalid Timestamp. The request timestap " + timestamp.String() + " was off the current time " + now.String() + " by more than " + strconv.FormatInt(int64(timestampTolerance), 10) + " seconds.")
	}

	return nil
}
