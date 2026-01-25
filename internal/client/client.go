package client

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"chatdoor/internal/crypto"
	"chatdoor/internal/iot"
)

type Client struct {
	serverPubKey []byte
	addr         string
	conn         *websocket.Conn
	keypair      *crypto.KeyPair
	connMutex    sync.Mutex
	stopFake     chan bool
}

func New(serverPubKey []byte, addr string) *Client {
	return &Client{
		serverPubKey: serverPubKey,
		addr:         addr,
		stopFake:     make(chan bool),
	}
}

func (c *Client) Start() error {
	// Generate ephemeral keypair for this session
	keypair, err := crypto.GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("generating keypair: %w", err)
	}
	c.keypair = keypair

	// Connect to server
	u := url.URL{Scheme: "ws", Host: c.addr, Path: "/ws"}
	fmt.Printf("🔌 Connecting to %s...\n", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	c.connMutex.Lock()
	c.conn = conn
	c.connMutex.Unlock()

	fmt.Println("✓ Connected!")

	// Exchange keys
	if err := c.exchangeKeys(); err != nil {
		return fmt.Errorf("key exchange failed: %w", err)
	}

	fmt.Println("🔐 Secure channel established")
	fmt.Println("🏠 Observing IoT devices...")
	fmt.Println()
	fmt.Println("Type your message and press Enter to send:")
	fmt.Println()

	// Start fake traffic generator
	go c.generateFakeTraffic()

	// Start message receiver
	go c.receiveMessages()

	// Handle user input
	c.handleInput()

	return nil
}

func (c *Client) exchangeKeys() error {
	// Receive server's public key (verification)
	var serverKey map[string]string
	if err := c.conn.ReadJSON(&serverKey); err != nil {
		return fmt.Errorf("receiving server key: %w", err)
	}

	if serverKey["type"] != "handshake" {
		return fmt.Errorf("unexpected message type")
	}

	receivedPubKey, err := crypto.Base64ToPublicKey(serverKey["public_key"])
	if err != nil {
		return fmt.Errorf("invalid server public key: %w", err)
	}

	// Verify it matches what we were given
	if !bytesEqual(receivedPubKey, c.serverPubKey) {
		return fmt.Errorf("server public key mismatch - possible MITM attack!")
	}

	// Send our public key
	keyMsg := map[string]string{
		"type":       "handshake",
		"public_key": crypto.PublicKeyToBase64(c.keypair.PublicKey()),
	}

	if err := c.conn.WriteJSON(keyMsg); err != nil {
		return fmt.Errorf("sending public key: %w", err)
	}

	return nil
}

func (c *Client) generateFakeTraffic() {
	ticker := time.NewTicker(time.Duration(30+rand.Intn(30)) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopFake:
			return
		case <-ticker.C:
			msg := iot.GenerateFakeTraffic()
			c.sendMessage(msg)

			// Show fake traffic occasionally (10% chance)
			if rand.Float64() < 0.1 {
				fmt.Printf("\r\033[K[IoT] %s: %v\n> ", msg.DeviceID, msg.Data)
			}

			// Reset ticker with random interval
			ticker.Reset(time.Duration(30+rand.Intn(30)) * time.Second)
		}
	}
}

func (c *Client) sendMessage(msg iot.Message) error {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if c.conn == nil {
		return fmt.Errorf("no connection")
	}

	return c.conn.WriteJSON(msg)
}

func (c *Client) receiveMessages() {
	for {
		var msg iot.Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Connection closed: %v", err)
			}
			c.cleanup()
			return
		}

		// Check if it's a real message
		if encryptedData, isReal := iot.ExtractRealMessage(msg); isReal {
			decrypted, err := c.keypair.Decrypt(encryptedData, c.serverPubKey)
			if err != nil {
				log.Printf("Decryption failed: %v", err)
				continue
			}

			fmt.Printf("\r\033[K[Them]: %s\n> ", string(decrypted))
		}
	}
}

func (c *Client) handleInput() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")

	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			fmt.Print("> ")
			continue
		}

		if text == "/quit" {
			c.cleanup()
			os.Exit(0)
		}

		// Encrypt and send
		encrypted, err := c.keypair.Encrypt([]byte(text), c.serverPubKey)
		if err != nil {
			log.Printf("Encryption failed: %v", err)
			fmt.Print("> ")
			continue
		}

		msg := iot.CreateRealMessage(encrypted)
		if err := c.sendMessage(msg); err != nil {
			log.Printf("Send failed: %v", err)
		}

		fmt.Print("> ")
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Input error: %v", err)
	}
}

func (c *Client) cleanup() {
	c.stopFake <- true
	c.connMutex.Lock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.connMutex.Unlock()
	fmt.Println()
	fmt.Println("Connection closed")
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
