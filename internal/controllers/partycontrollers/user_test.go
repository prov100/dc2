package partycontrollers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	partyproto "github.com/prov100/dc2/internal/protogen/party/v1"
	"github.com/prov100/dc2/test"
	_ "github.com/go-sql-driver/mysql"
)

func TestGetUsers(t *testing.T) {
	var err error
	err = test.LoadSQL(log, dbService)
	if err != nil {
		t.Error(err)
		return
	}

	tokenString, _ := LoginUser()

	w := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "http://localhost:6060/v0.1/users", nil)
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+tokenString)

	mux.ServeHTTP(w, req)

	resp := w.Result()
	// Check the status code is what we expect.
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code %d", resp.StatusCode)
		return
	}

	usersResponse := &partyproto.GetUsersResponse{}
	err = json.Unmarshal(w.Body.Bytes(), usersResponse)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println("usersResponse", usersResponse)

	user1, err := GetUser("auth0|66fd06d0bfea78a82bb42459", "sprov300@gmail.com", "https://s.gravatar.com/avatar/52ab1cc37bb42deb67ea939fd68ff7d4?s=480&r=pg&d=https%3A%2F%2Fcdn.auth0.com%2Favatars%2Fsp.png", "sprov300@gmail.com")
	if err != nil {
		t.Error(err)
		return
	}

	user2, err := GetUser("auth0|66fcdfb6d20dcb68e3fcbc3b", "sprov200@gmail.com", "https://s.gravatar.com/avatar/06004bcbe9705b0ba5d7c4923fef0061?s=480&r=pg&d=https%3A%2F%2Fcdn.auth0.com%2Favatars%2Fsp.png", "sprov200@gmail.com")
	if err != nil {
		t.Error(err)
		return
	}

	users := []*partyproto.User{}
	users = append(users, user1, user2)

	expected := &partyproto.GetUsersResponse{}
	expected.Users = users

	fmt.Println("expected", expected)

	if !reflect.DeepEqual(usersResponse, expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			usersResponse, expected)
		return
	}
}

func TestGetUser(t *testing.T) {
	err := test.LoadSQL(log, dbService)
	if err != nil {
		t.Error(err)
		return
	}

	tokenString, _ := LoginUser()

	w := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "https://localhost:6060/v0.1/users/auth0|66fd06d0bfea78a82bb42459/", nil)
	if err != nil {
		t.Error(err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+tokenString)
	mux.ServeHTTP(w, req)

	resp := w.Result()
	// Check the status code is what we expect.
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code %d", resp.StatusCode)
		return
	}
	fmt.Println("resp", resp)
	fmt.Println("w.w.Body.String()", w.Body.String())
	fmt.Println("w.Body.Bytes()", w.Body.Bytes())

	userResponse := &partyproto.GetUserResponse{}
	err = json.Unmarshal(w.Body.Bytes(), userResponse)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("userResponse", userResponse)

	user, err := GetUser("auth0|66fd06d0bfea78a82bb42459", "sprov300@gmail.com", "https://s.gravatar.com/avatar/52ab1cc37bb42deb67ea939fd68ff7d4?s=480&r=pg&d=https%3A%2F%2Fcdn.auth0.com%2Favatars%2Fsp.png", "sprov300@gmail.com")
	if err != nil {
		t.Error(err)
		return
	}

	expected := &partyproto.GetUserResponse{}
	expected.User = user

	fmt.Println("expected", expected)

	if !reflect.DeepEqual(userResponse, expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			userResponse, expected)
		return
	}
}

func TestGetUserByEmail(t *testing.T) {
	err := test.LoadSQL(log, dbService)
	if err != nil {
		t.Error(err)
		return
	}

	tokenString, _ := LoginUser()

	w := httptest.NewRecorder()

	data := []byte(`{"email" : "sprov300@gmail.com"}`)

	req, err := http.NewRequest("POST", "https://localhost:6060/v0.1/users/getuserbyemail", bytes.NewBuffer(data))
	if err != nil {
		t.Error(err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+tokenString)
	mux.ServeHTTP(w, req)

	resp := w.Result()
	// Check the status code is what we expect.
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code %d", resp.StatusCode)
		return
	}
	fmt.Println("resp", resp)
	fmt.Println("w.w.Body.String()", w.Body.String())
	fmt.Println("w.Body.Bytes()", w.Body.Bytes())

	userResponse := &partyproto.GetUserByEmailResponse{}
	err = json.Unmarshal(w.Body.Bytes(), userResponse)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("userResponse", userResponse)

	user, err := GetUser("auth0|66fd06d0bfea78a82bb42459", "sprov300@gmail.com", "https://s.gravatar.com/avatar/52ab1cc37bb42deb67ea939fd68ff7d4?s=480&r=pg&d=https%3A%2F%2Fcdn.auth0.com%2Favatars%2Fsp.png", "sprov300@gmail.com")
	if err != nil {
		t.Error(err)
		return
	}

	expected := &partyproto.GetUserByEmailResponse{}
	expected.User = user

	fmt.Println("expected", expected)

	if !reflect.DeepEqual(userResponse, expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			userResponse, expected)
		return
	}
}

func TestChangePassword(t *testing.T) {
	var err error
	err = test.LoadSQL(log, dbService)
	if err != nil {
		t.Error(err)
		return
	}

	tokenString, _ := LoginUser()

	w := httptest.NewRecorder()

	data := []byte(`{"email" : "sprov300@gmail.com"}`)

	req, err := http.NewRequest("POST", "http://localhost:6060/v0.1/users/change_password", bytes.NewBuffer(data))
	if err != nil {
		t.Error(err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+tokenString)

	mux.ServeHTTP(w, req)

	resp := w.Result()
	fmt.Println("resp", resp)
	// Check the status code is what we expect.
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code %d", resp.StatusCode)
		return
	}

	expected := string(`"We've just sent you an email to reset your password."` + "\n")

	if w.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			w.Body.String(), expected)
		return
	}
}

func GetUser(id string, email string, picture string, name string) (*partyproto.User, error) {
	user := new(partyproto.User)
	user.Id = id
	user.Email = email
	user.Picture = picture
	user.Name = name
	return user, nil
}
