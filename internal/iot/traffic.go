package iot

import (
	"encoding/json"
	"math/rand"
	"time"
)

// MessageType represents the type of IoT message
type MessageType string

const (
	TypeSensor MessageType = "sensor"
	TypeEvent  MessageType = "event"
	TypeStatus MessageType = "status"
)

// Message represents an IoT message (both real and fake)
type Message struct {
	Type      MessageType            `json:"type"`
	DeviceID  string                 `json:"device_id"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	IsReal    bool                   `json:"-"` // Not serialized, internal use only
}

// Device represents a fake IoT device
type Device struct {
	ID       string
	Type     string
	lastTemp float64
}

var (
	devices = []Device{
		{ID: "temp_living_room", Type: "temperature", lastTemp: 21.0},
		{ID: "temp_bedroom", Type: "temperature", lastTemp: 19.5},
		{ID: "temp_kitchen", Type: "temperature", lastTemp: 22.0},
		{ID: "humidity_bathroom", Type: "humidity", lastTemp: 60.0},
		{ID: "humidity_bedroom", Type: "humidity", lastTemp: 55.0},
		{ID: "motion_hallway", Type: "motion"},
		{ID: "motion_living_room", Type: "motion"},
		{ID: "door_front", Type: "door"},
		{ID: "door_back", Type: "door"},
		{ID: "window_bedroom", Type: "window"},
	}
)

// GenerateFakeTraffic creates a realistic fake IoT message
func GenerateFakeTraffic() Message {
	device := devices[rand.Intn(len(devices))]

	msg := Message{
		DeviceID:  device.ID,
		Timestamp: time.Now().Unix(),
		Data:      make(map[string]interface{}),
		IsReal:    false,
	}

	switch device.Type {
	case "temperature":
		// Temperature drifts slowly
		drift := (rand.Float64() - 0.5) * 0.3
		device.lastTemp += drift
		if device.lastTemp < 15 {
			device.lastTemp = 15
		}
		if device.lastTemp > 28 {
			device.lastTemp = 28
		}

		msg.Type = TypeSensor
		msg.Data["value"] = roundFloat(device.lastTemp, 1)
		msg.Data["unit"] = "celsius"

	case "humidity":
		// Humidity drifts slowly
		drift := (rand.Float64() - 0.5) * 2
		device.lastTemp += drift
		if device.lastTemp < 30 {
			device.lastTemp = 30
		}
		if device.lastTemp > 80 {
			device.lastTemp = 80
		}

		msg.Type = TypeSensor
		msg.Data["value"] = roundFloat(device.lastTemp, 0)
		msg.Data["unit"] = "percent"

	case "motion":
		// Motion sensors trigger randomly (10% chance)
		msg.Type = TypeEvent
		if rand.Float64() < 0.1 {
			msg.Data["motion"] = "detected"
		} else {
			msg.Data["motion"] = "clear"
		}

	case "door", "window":
		// Door/window events are rare (5% chance of state change)
		msg.Type = TypeEvent
		if rand.Float64() < 0.05 {
			if rand.Float64() < 0.5 {
				msg.Data["state"] = "opened"
			} else {
				msg.Data["state"] = "closed"
			}
		} else {
			msg.Data["state"] = "closed"
		}
	}

	return msg
}

// CreateRealMessage creates a message containing encrypted chat data
func CreateRealMessage(encryptedData []byte) Message {
	// Disguise as a door/window event with encoded data
	device := devices[7+rand.Intn(3)] // Use door/window devices

	msg := Message{
		Type:      TypeEvent,
		DeviceID:  device.ID,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"state":   encodeState(encryptedData),
			"payload": base64Encode(encryptedData),
		},
		IsReal: true,
	}

	return msg
}

// ExtractRealMessage checks if a message contains real encrypted data
func ExtractRealMessage(msg Message) ([]byte, bool) {
	if msg.Type != TypeEvent {
		return nil, false
	}

	payload, ok := msg.Data["payload"].(string)
	if !ok || payload == "" {
		return nil, false
	}

	data, err := base64Decode(payload)
	if err != nil {
		return nil, false
	}

	return data, true
}

// ToJSON serializes message to JSON
func (m Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON deserializes message from JSON
func FromJSON(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// Helper functions

func roundFloat(val float64, precision int) float64 {
	ratio := 1.0
	for i := 0; i < precision; i++ {
		ratio *= 10
	}
	return float64(int(val*ratio+0.5)) / ratio
}

func encodeState(data []byte) string {
	// Use first byte to determine fake state
	if len(data) > 0 && data[0]%2 == 0 {
		return "opened"
	}
	return "closed"
}

func base64Encode(data []byte) string {
	// Simple encoding wrapper
	encoded := ""
	for _, b := range data {
		encoded += string(rune(b))
	}
	return encoded
}

func base64Decode(s string) ([]byte, error) {
	data := make([]byte, len(s))
	for i, r := range s {
		data[i] = byte(r)
	}
	return data, nil
}
