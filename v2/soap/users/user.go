package users

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	EmailAddress  string             `json:"email" bson:"email"`
	Password      string             `json:"password" bson:"password"`
	BadAttempts   int                `json:"badAttempts" bson:"badAttempts"`
	FirstName     string             `json:"firstName" bson:"firstName"`
	MiddleName    string             `json:"middleName,omitempty" bson:"middleName,omitempty"`
	LastName      string             `json:"lastName" bson:"lastName"`
	ResetToken    string             `json:"-" bson:"resettoken,omitempty"`
	ResetTokenExp time.Time          `json:"-" bson:"resettokenexp,omitempty"`
	Administrator bool               `json:"administrator" bson:"administrator"`
	PlanID        primitive.ObjectID `json:"planId,omitempty" bson:"planId,omitempty"`
	BibleVersion  string             `json:"bibleVersion,omitempty" bson:"bibleVersion,omitempty"`
}

type ByUser []User

func (c ByUser) Len() int { return len(c) }
func (c ByUser) Less(i, j int) bool {
	if c[i].LastName == c[j].LastName {
		if c[i].FirstName == c[j].LastName {
			return c[i].MiddleName < c[j].MiddleName
		}
		return c[i].FirstName < c[j].FirstName
	}
	return c[i].LastName < c[j].LastName
}
func (c ByUser) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (u *User) SetPassword(passwd string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(passwd), 12)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	u.BadAttempts = 0
	return nil
}

func (u *User) Authenticate(passwd string) error {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(passwd))
	if err != nil {
		u.BadAttempts++
		return errors.New("email address/password mismatch")
	}
	if u.BadAttempts > 2 {
		return errors.New("account locked")
	}

	u.BadAttempts = 0
	return nil
}

func (u *User) CreateRandomPassword() (string, error) {
	result := ""
	upCharacters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowCharacters := "abcdefghijklmnopqrstuvwxyz"
	numbers := "1234567890"

	for i := 0; i < 16; i++ {
		switch {
		case i%3 == 0:
			upPos := rand.Intn(len(upCharacters))
			result = result + string(upCharacters[upPos])
		case i%3 == 1:
			lowPos := rand.Intn(len(lowCharacters))
			result = result + string(lowCharacters[lowPos])
		case i%3 == 2:
			numPos := rand.Intn(len(numbers))
			result = result + string(numbers[numPos])
		}
	}
	err := u.SetPassword(result)
	if err != nil {
		return "", err
	}
	u.BadAttempts = -1
	return result, nil
}

func (u *User) Unlock() {
	u.BadAttempts = 0
}

func (u *User) GetFullName() string {
	if u.MiddleName != "" {
		return fmt.Sprintf("%s %s. %s", u.FirstName, u.MiddleName[:1], u.LastName)
	}
	return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
}

func (u *User) GetLastFirst() string {
	if u.MiddleName != "" {
		return fmt.Sprintf("%s, %s %s.", u.LastName, u.FirstName, u.MiddleName[:1])
	}
	return fmt.Sprintf("%s, %s", u.LastName, u.FirstName)
}

func (u *User) CreateResetToken() string {
	result := ""
	upCharacters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowCharacters := "abcdefghijklmnopqrstuvwxyz"
	numbers := "1234567890"

	for i := 0; i < 16; i++ {
		switch {
		case i%3 == 0:
			upPos := rand.Intn(len(upCharacters))
			result = result + string(upCharacters[upPos])
		case i%3 == 1:
			lowPos := rand.Intn(len(lowCharacters))
			result = result + string(lowCharacters[lowPos])
		case i%3 == 2:
			numPos := rand.Intn(len(numbers))
			result = result + string(numbers[numPos])
		}
	}
	now := time.Now()
	u.ResetTokenExp = now.AddDate(0, 0, 1)
	u.ResetToken = result
	return result
}

func (u *User) CheckResetToken(token string) error {
	now := time.Now()
	if u.ResetToken == token && u.ResetTokenExp.After(now) {
		return nil
	} else {
		if !u.ResetTokenExp.After(now) {
			return errors.New("reset token expired")
		}
		return errors.New("reset token error")
	}
}
