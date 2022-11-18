package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Event struct {
	ID        int64     `json:"id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	Initiator string    `json:"initiator,omitempty"`
	Subject   string    `json:"subject,omitempty"`
	Action    string    `json:"action,omitempty"`
	Rollback  any       `json:"rollback,omitempty"`
	Update    any       `json:"update,omitempty"`
}

type User struct {
	ID      int64  `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Age     int    `json:"age,omitempty"`
	IsAdult bool   `json:"is_adult,omitempty"`
}

func main() {
	u, err := getUser(1)
	if err != nil {
		log.Fatal(err)
	}

	err = updateUser(u)
	if err != nil {
		log.Fatal(err)
	}
}

func updateUser(u *User) error {
	testNames := []string{
		"Sam",
		"Serj",
		"Aaron",
		"Henry",
		"Steven",
		"Jackie",
		"Alex",
		"Johan",
		"Bjorn",
		"Anders",
	}
	for idx, name := range testNames {
		if name == u.Name {
			log.Fatal(errors.New("old and new names are the same"))
		}
		oldData := *u
		u.Name = name
		newData := u
		err := addEvent(idx+1, "admin_user", "some_user", "user_update", &oldData, newData)
		if err != nil {
			return err
		}

		resp, err := sendUser(newData)
		if err != nil {
			return err
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		log.Printf("Update user status: %s", body)
	}
	return nil
}

func addEvent(idx int, initiator, subject, action string, oldData, newData any) error {
	e := &Event{
		ID:        int64(idx),
		Initiator: initiator,
		Subject:   subject,
		Action:    action,
		Rollback:  oldData,
		Update:    newData,
	}
	resp, err := sendEvent(e)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Printf("Add event status: %s", body)
	return nil
}

func sendEvent(e *Event) (*http.Response, error) {
	serialized, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(serialized)
	req, err := http.NewRequest("POST", "http://localhost:8080/event/add", buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	c := &http.Client{}

	return c.Do(req)
}

func sendUser(u *User) (*http.Response, error) {
	return resp, nil
}

func getUser(id int64) (*User, error) {
	url := fmt.Sprintf("http://localhost:8080/user/%d", id)
	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	serialized, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	u := &User{}
	err = json.Unmarshal(serialized, u)
	if err != nil {
		return nil, err
	}

	return u, nil
}
