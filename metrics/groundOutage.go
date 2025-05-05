package metrics

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GroundOutage struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id"`
	OutageDate         time.Time          `json:"outageDate" bson:"outageDate"`
	GroundSystem       string             `json:"groundSystem" bson:"groundSystem"`
	Classification     string             `json:"classification" bson:"classification"`
	OutageNumber       uint               `json:"outageNumber" bson:"outageNumber"`
	OutageMinutes      uint               `json:"outageMinutes" bson:"outageMinutes"`
	Subsystem          string             `json:"subSystem" bson:"subSystem"`
	ReferenceID        string             `json:"referenceId" bson:"referenceId"`
	MajorSystem        string             `json:"majorSystem" bson:"majorSystem"`
	EncryptedProblem   string             `json:"-" bson:"problem"`
	ExcryptedFixAction string             `json:"-" bson:"fixAction"`
	Problem            string             `json:"problem" bson:"-"`
	FixAction          string             `json:"fixAction" bson:"-"`
	MissionOutage      bool               `json:"missionOutage" bson:"missionOutage"`
	Capability         string             `json:"capability,omitempty" bson:"capability,omitempty"`
}

type ByOutage []GroundOutage

func (c ByOutage) Len() int { return len(c) }
func (c ByOutage) Less(i, j int) bool {
	if c[i].OutageDate.Equal(c[j].OutageDate) {
		return c[i].OutageNumber < c[j].OutageNumber
	}
	return c[i].OutageDate.Before(c[j].OutageDate)
}
func (c ByOutage) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (g *GroundOutage) Encrypt() {
	// get the security key from the environment and create a byte array from
	// it for the cipher
	keyString := os.Getenv("SECURITY_KEY")
	key := []byte(keyString)

	// create the aes cipher using our security key
	c, err := aes.NewCipher(key)
	if err != nil {
		log.Println(err)
	}

	// create the GCM for the symetric key
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		log.Println(err)
	}

	// create a new byte array to hold the nonce which must be passed to create
	// the encrypted value.
	nonce := make([]byte, gcm.NonceSize())
	// and populate it with a random code
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		log.Println(err)
	}

	if len(g.Problem) == 0 || g.Problem == "" {
		g.EncryptedProblem = ""
	} else {
		g.EncryptedProblem = string(gcm.Seal(nonce, nonce, []byte(g.Problem), nil))
	}

	if len(g.FixAction) == 0 || g.FixAction == "" {
		g.ExcryptedFixAction = ""
	} else {
		g.ExcryptedFixAction = string(gcm.Seal(nonce, nonce, []byte(g.FixAction), nil))
	}
}

func (g *GroundOutage) Decrypt() {
	// get the security key from the environment and create a byte array from
	// it for the cipher
	keyString := os.Getenv("SECURITY_KEY")
	key := []byte(keyString)

	// create the aes cipher using our security key
	c, err := aes.NewCipher(key)
	if err != nil {
		log.Println(err)
	}

	// create the GCM for the symetric key
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		log.Println(err)
	}

	nonceSize := gcm.NonceSize()
	if len(g.EncryptedProblem) < nonceSize {
		g.Problem = ""
	} else {
		nonce, prob := g.EncryptedProblem[:nonceSize], g.EncryptedProblem[nonceSize:]
		plainText, err := gcm.Open(nil, []byte(nonce), []byte(prob), nil)
		if err != nil {
			log.Println(err)
		}
		g.Problem = string(plainText)
	}
	if len(g.ExcryptedFixAction) < nonceSize {
		g.FixAction = ""
	} else {
		nonce, fix := g.ExcryptedFixAction[:nonceSize], g.ExcryptedFixAction[nonceSize:]
		plaintext, err := gcm.Open(nil, []byte(nonce), []byte(fix), nil)
		if err != nil {
			log.Println(err)
		}
		g.FixAction = string(plaintext)
	}
}

func (g *GroundOutage) GetProblem() string {
	prob := []byte(g.Problem)
	if len(prob) == 0 {
		return ""
	}
	// get the security key from the environment and create a byte array from
	// it for the cipher
	keyString := os.Getenv("SECURITY_KEY")
	key := []byte(keyString)

	// create the aes cipher using our security key
	c, err := aes.NewCipher(key)
	if err != nil {
		log.Println(err)
	}

	// create the GCM for the symetric key
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		log.Println(err)
	}

	nonceSize := gcm.NonceSize()
	if len(prob) < nonceSize {
		return ""
	}

	nonce, prob := prob[:nonceSize], prob[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, prob, nil)
	if err != nil {
		log.Println(err)
	}
	return string(plainText)
}
