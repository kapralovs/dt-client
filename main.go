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
	// serialized, err := json.Marshal(u)
	// if err != nil {
	// 	return nil, err
	// }

	// buf := bytes.NewBuffer(serialized)
	url := fmt.Sprintf("http://localhost:8080/patch/rollback/%d/%d", eventId, u.ID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	c := &http.Client{}

	return c.Do(req)
}

// func makeRollback() (*User, error) {
// 	u, err := getPatched(1)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return u, nil
// }

// func getPatched() (*User, error) {
// 	url := "http://localhost:8080/patch/rollback/:entity_id"
// 	return u, nil
// }

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
	for _, name := range testNames {
		if name == u.Name {
			log.Fatal(errors.New("old and new names are the same"))
		}
		u.Name = name
		newData := u

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
