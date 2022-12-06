package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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
	ID      int64     `json:"id,omitempty"`
	Name    string    `json:"name,omitempty"`
	Age     int       `json:"age,omitempty"`
	Bag     *Backpack `json:"bag,omitempty"`
	IsAdult bool      `json:"is_adult,omitempty"`
}

type Backpack struct {
	Phone string `json:"phone,omitempty"`
	Food  string `json:"food,omitempty"`
	Gun   string `json:"gun,omitempty"`
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

	u, err = getUser(1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("before rollback: %v\n", u)

	u, err = rollbackToID(u, 4)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("after rollback: %v\n", u)

	strResp, err := sendDate()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("sendDate result: %s\n", strResp)

	events, err := getEventsList()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("events after required date: %d\n", len(events))
}

func makeQueryParam(t time.Time) (string, error) {
	// Example: 2006-01-02T15:04:05+07:00
	// var month string
	t = t.Add(time.Hour * 72)
	fmt.Printf("makeQueryParam: raw time: %v\n", t)
	formatted := t.Format(time.RFC3339)
	fmt.Printf("makeQueryParam: formatted time: %v\n", t)

	param := url.QueryEscape(formatted)
	return param, nil

	// layout := "2006-01-02T15:04:05+00:00"
	// format := t.Format(layout)
	// t, _ = time.Parse(layout, format)
	// fmt.Println(t)
	// param := url.QueryEscape(t.String())
	// return param, nil
}

func getEventsList() ([]*Event, error) {
	t := time.Now()
	param, err := makeQueryParam(t)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://localhost:8080/events?created_at=%s", param)
	fmt.Println(url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// fmt.Println(string(body))

	events := make([]*Event, 0)
	err = json.Unmarshal(body, &events)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func sendDate() (string, error) {
	var month string
	y, m, d := time.Now().Date()
	if m == time.November {
		month = "11"
	}
	date := fmt.Sprintf("%d-%s-%d", y, month, d)
	queryParam := url.QueryEscape(date)
	url := fmt.Sprintf("http://localhost:8080/parse_date?created_at=%s", queryParam)
	fmt.Printf("sendDate url: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func rollbackToID(u *User, id int64) (*User, error) {
	resp, err := makeRollback(u, id)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	patched := &User{}
	err = json.Unmarshal(body, patched)
	if err != nil {
		return nil, err
	}

	return patched, nil
}

func makeRollback(u *User, eventId int64) (*http.Response, error) {
	url := fmt.Sprintf("http://localhost:8080/patch/rollback/%d/%d", eventId, u.ID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	c := &http.Client{}

	return c.Do(req)
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
		if idx == 4 {
			age := u.Age
			u.Age = age + 2
		}
		u.Name = name
		newData := u

		if idx == 6 {
			u.Bag.Food = "Cola"
		}
		if idx == 8 {
			u.Bag.Gun = ""
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

func sendUser(u *User) (*http.Response, error) {
	serialized, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(serialized)
	url := fmt.Sprintf("http://localhost:8080/user/update/%d", u.ID)
	req, err := http.NewRequest("PUT", url, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	c := &http.Client{}

	return c.Do(req)
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
