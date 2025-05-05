package metrics

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/erneap/models/v2/systemdata"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MissionSensorOutage struct {
	TotalOutageMinutes     uint `json:"totalOutageMinutes"`
	PartialLBOutageMinutes uint `json:"partialLBOutageMinutes"`
	PartialHBOutageMinutes uint `json:"partialHBOutageMinutes"`
}

type MissionSensor struct {
	SensorID          string                  `json:"sensorID"`
	SensorType        systemdata.GeneralTypes `json:"sensorType"`
	PreflightMinutes  uint                    `json:"preflightMinutes"`
	ScheduledMinutes  uint                    `json:"scheduledMinutes"`
	ExecutedMinutes   uint                    `json:"executedMinutes"`
	PostflightMinutes uint                    `json:"postflightMinutes"`
	AdditionalMinutes uint                    `json:"additionalMinutes"`
	FinalCode         uint                    `json:"finalCode"`
	KitNumber         string                  `json:"kitNumber"`
	SensorOutage      MissionSensorOutage     `json:"sensorOutage"`
	GroundOutage      uint                    `json:"groundOutage"`
	HasHap            bool                    `json:"hasHap"`
	TowerID           uint                    `json:"towerID,omitempty"`
	SortID            uint                    `json:"sortID"`
	Comments          string                  `json:"comments"`
	Images            []systemdata.ImageType  `json:"images"`
}

type ByMissionSensor []MissionSensor

func (c ByMissionSensor) Len() int { return len(c) }
func (c ByMissionSensor) Less(i, j int) bool {
	return c[i].SortID < c[j].SortID
}
func (c ByMissionSensor) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type MissionData struct {
	Exploitation   string          `json:"exploitation"`
	TailNumber     string          `json:"tailNumber"`
	Communications string          `json:"communications"`
	PrimaryDCGS    string          `json:"primaryDCGS"`
	Cancelled      bool            `json:"cancelled"`
	Executed       bool            `json:"executed,omitempty"`
	Aborted        bool            `json:"aborted"`
	IndefDelay     bool            `json:"indefDelay"`
	MissionOverlap uint            `json:"missionOverlap"`
	Comments       string          `json:"comments"`
	Sensors        []MissionSensor `json:"sensors,omitempty"`
}

type Mission struct {
	ID                   primitive.ObjectID `json:"id" bson:"_id"`
	MissionDate          time.Time          `json:"missionDate" bson:"missionDate"`
	PlatformID           string             `json:"platformID" bson:"platformID"`
	SortieID             uint               `json:"sortieID" bson:"sortieID"`
	EncryptedMissionData string             `json:"-" bson:"encryptedMissionData"`
	MissionData          *MissionData       `json:"missionData" bson:"-"`
}

type ByMission []Mission

func (c ByMission) Len() int { return len(c) }
func (c ByMission) Less(i, j int) bool {
	if c[i].MissionDate.Equal(c[j].MissionDate) {
		if strings.EqualFold(c[i].PlatformID, c[j].PlatformID) {
			return c[i].SortieID < c[j].SortieID
		}
		return c[i].PlatformID < c[j].PlatformID
	}
	return c[i].MissionDate.Before(c[j].MissionDate)
}
func (c ByMission) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (m *Mission) Encrypt() error {
	data, err := json.Marshal(m.MissionData)
	if err != nil {
		return err
	}

	// get the security key from the environment and create a byte array from
	// it for the cipher
	keyString := os.Getenv("SECURITY_KEY")
	key := []byte(keyString)

	// create the aes cipher using our security key
	c, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// create the GCM for the symetric key
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return err
	}

	// create a new byte array to hold the nonce which must be passed to create
	// the encrypted value.
	nonce := make([]byte, gcm.NonceSize())
	// and populate it with a random code
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	// lastly, encrypt the value and store in problem property above
	m.EncryptedMissionData = string(gcm.Seal(nonce, nonce, data, nil))

	return nil
}

func (m *Mission) Decrypt() error {
	prob := []byte(m.EncryptedMissionData)
	if len(prob) == 0 {
		return errors.New("no encrypted mission data")
	}
	// get the security key from the environment and create a byte array from
	// it for the cipher
	keyString := os.Getenv("SECURITY_KEY")
	key := []byte(keyString)

	// create the aes cipher using our security key
	c, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// create the GCM for the symetric key
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return err
	}

	nonceSize := gcm.NonceSize()
	if len(prob) < nonceSize {
		return errors.New("encrypted data too small")
	}

	nonce, prob := prob[:nonceSize], prob[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, prob, nil)
	if err != nil {
		return err
	}
	json.Unmarshal(plainText, &m.MissionData)
	return nil
}
