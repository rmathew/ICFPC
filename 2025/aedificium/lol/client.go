package lol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func sendJSON(path string, in interface{}, out interface{}) error {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	url := "https://31pwr5t6ij.execute-api.eu-west-2.amazonaws.com" + path
	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusBadRequest {
			log.Printf("ERROR: Request JSON\n%s\n", string(b))
			respBody, _ := ioutil.ReadAll(resp.Body)
			log.Printf("ERROR: Response Body\n%s\n", string(respBody))
		}
		return fmt.Errorf("HTTP error %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

type Client struct {
	id string
}

type RoomDoor struct {
	Room int `json:"room"`
	Door int `json:"door"`
}

type Connection struct {
	From RoomDoor `json:"from"`
	To   RoomDoor `json:"to"`
}

type GuessedMap struct {
	Rooms        []int        `json:"rooms"`
	StartingRoom int          `json:"startingRoom"`
	Connections  []Connection `json:"connections"`
}

func NewClient(id string) *Client {
	return &Client{id: id}
}

func maybeAnnotateError(err error, extraErr string) error {
	if extraErr == "" {
		return err
	}
	return fmt.Errorf("%w (%s)", err, extraErr)
}

func (c *Client) SelectProblem(prob string) error {
	reqBody := struct {
		Id          string `json:"id"`
		ProblemName string `json:"problemName"`
	}{c.id, prob}
	resBody := struct {
		ProblemName string `json:"problemName"`
		Error       string `json:"error,omitempty"`
	}{}
	if err := sendJSON("/select", reqBody, &resBody); err != nil {
		return maybeAnnotateError(err, resBody.Error)
	}
	selProb := resBody.ProblemName
	log.Printf("INFO: Selected Problem: '%s'", selProb)
	return nil
}

func (c *Client) Explore(plans []string) ([][]int, error) {
	reqBody := struct {
		Id    string   `json:"id"`
		Plans []string `json:"plans"`
	}{c.id, plans}
	resBody := struct {
		Results    [][]int `json:"results"`
		QueryCount int     `json:"queryCount"`
		Error      string  `json:"error,omitempty"`
	}{}
	if err := sendJSON("/explore", reqBody, &resBody); err != nil {
		return nil, maybeAnnotateError(err, resBody.Error)
	}
	log.Printf("INFO: Post-exploration query-count: %d", resBody.QueryCount)
	return resBody.Results, nil
}

func (c *Client) Guess(guessedMap *GuessedMap) (bool, error) {
	reqBody := struct {
		Id  string     `json:"id"`
		Map GuessedMap `json:"map"`
	}{c.id, *guessedMap}
	resBody := struct {
		Correct bool   `json:"correct"`
		Error   string `json:"error,omitempty"`
	}{}
	if err := sendJSON("/guess", reqBody, &resBody); err != nil {
		return false, maybeAnnotateError(err, resBody.Error)
	}
	log.Printf("INFO: Guessed map was correct: %v", resBody.Correct)
	return resBody.Correct, nil
}
