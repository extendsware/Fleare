package auth

import (
	"fmt"
	"sync"
)

// Role type represents the authorization level for database access
type Role string

// Define roles with different access levels
const (
	Basic Role = "Basic" // Regular user with basic access
	Read  Role = "Read"  // Read-only access
	Write Role = "Write" // Write access
	Admin Role = "Admin" // Full administrative access
)

type User struct {
	Username string
	Password string
	Role     Role
}

type Users struct {
	store  map[string]*User
	stLock *sync.RWMutex
}

var UserStore *Users

func init() {
	UserStore = NewUsersStore()
}

func NewUsersStore() (users *Users) {
	users = &Users{
		store:  make(map[string]*User),
		stLock: &sync.RWMutex{},
	}
	return
}

func (users *Users) Validate(username, password string) (user *User, err error) {
	users.stLock.RLock()
	defer users.stLock.RUnlock()
	isPresent := false
	if user, isPresent = users.store[username]; !isPresent {
		return nil, fmt.Errorf("The username or password you entered is incorrect. Please try again with valid credentials.")
	}
	if user.Password != password {
		return nil, fmt.Errorf("The username or password you entered is incorrect. Please try again with valid credentials.")
	}
	return
}

func (users *Users) ValidateDefault(username string) (user *User, err error) {
	users.stLock.RLock()
	defer users.stLock.RUnlock()
	isPresent := false
	if user, isPresent = users.store[username]; !isPresent {
		return nil, fmt.Errorf("user not found %s", username)
	}
	return
}

func (users *Users) Get(username string) (user *User, err error) {
	users.stLock.RLock()
	defer users.stLock.RUnlock()
	isPresent := false
	if user, isPresent = users.store[username]; !isPresent {
		return nil, fmt.Errorf("user not found %s", username)
	}
	return
}

func (users *Users) Add(username string, password string, role string) (user *User, err error) {
	user = &User{
		Username: username,
		Password: password,
		Role:     Role(role),
	}
	users.stLock.Lock()
	defer users.stLock.Unlock()
	users.store[username] = user
	return
}

func (users *Users) Delete(username string) (err error) {
	users.stLock.Lock()
	defer users.stLock.Unlock()
	if _, isPresent := users.store[username]; !isPresent {
		return fmt.Errorf("user not found %s", username)
	}
	delete(users.store, username)
	return
}
