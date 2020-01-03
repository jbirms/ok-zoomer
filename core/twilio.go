package main

import (
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type TwilioClient struct {
	AccountId 	string
	APIKey 		string
	PhoneNumber	string
}

func LoadTwilioConfigFromEnv() (TwilioClient, error) {
	//
	acctId, ok := os.LookupEnv("TWILIO_ACCT_ID")
	if !ok {
		acctId = "AC6711b0d93c5edbcd070cba8e377ac1b4" // jbirms default value
	}
	apiKey, ok := os.LookupEnv("TWILIO_API_KEY")
	if !ok {
		return TwilioClient{}, fmt.Errorf("missing TWILIO_API_KEY")
	}
	phoneNumber, ok := os.LookupEnv("TWILIO_PHONE_NUMBER")
	if !ok {
		return TwilioClient{}, fmt.Errorf("missing TWILIO_PHONE_NUMBER")
	}
	return TwilioClient{acctId, apiKey, phoneNumber}, nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func (tw *TwilioClient) SendMessage(toNumber, messageText string) error {
	sendUrl := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json",
		tw.AccountId)
	data := url.Values{}
	data.Set("Body", messageText)
	data.Set("From", tw.PhoneNumber)
	data.Set("To", toNumber)

	client := &http.Client{}
	r, err := http.NewRequest("POST", sendUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("had an issue posting to twilio: %s", err.Error())
	}
	r.Header.Add("Authorization", basicAuth(tw.AccountId, tw.APIKey))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, err := client.Do(r)
	if err != nil {
		log.Printf("had an issue making the request to twilio, %s", err.Error())
	} else if resp.StatusCode > 299 {
		log.Printf("got a non-200-level response code: %v", resp.StatusCode)
	}
	log.Println("successfully posted to twilio!")
	return nil
}

func GetTwilioHandler(sess *session.Session) func(w http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		twilioClient, err := LoadTwilioConfigFromEnv()
		if err != nil {
			panic(err)
		}
		numMedia, err := strconv.Atoi(req.FormValue("NumMedia"))
		if err != nil {
			panic(err)
		}
		fromNumber := req.FormValue("From")
		if err := req.ParseForm(); err != nil {
			fmt.Fprintf(rw, "ParseForm() err: %v", err)
			return
		}
		log.Printf("received a message from %s, %s. NumMedia: %v, Mime Type: %s",
			req.Form.Get("FromCity"), req.Form.Get("FromState"),
			numMedia,
			req.Form.Get("MediaContentType0"))

		rw.WriteHeader(200)
		if numMedia == 0 {
			log.Println("Got no media in this message, sending error reply")
			err = twilioClient.SendMessage(fromNumber, "You didn't send any media with your previous message!")
			if err != nil {
				log.Fatalf("hit an error trying to text a reply: %s", err.Error())
			}
			return
		} else {
			if numMedia > 1 {
				// warn about us only handling the first message
				err = twilioClient.SendMessage(fromNumber, "You sent multiple pieces of media, only handling the first one!")
			}
			dataUrl := req.FormValue("MediaUrl0")
			gifUrl, err := UrlToUrl(sess, dataUrl, fromNumber)
			if err != nil {
				log.Fatalf("had trouble generating the url: %s", err.Error())
			}
			twilioClient.SendMessage(fromNumber, "here's your gif: " + gifUrl)
		}

	}
}
