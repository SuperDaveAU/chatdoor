package server

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"chatdoor/internal/crypto"
	"chatdoor/internal/iot"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

type Server struct {
	keypair      *crypto.KeyPair
	port         int
	conn         *websocket.Conn
	connMutex    sync.Mutex
	clientPubKey []byte
	stopFake     chan bool
}

func New(keypair *crypto.KeyPair, port int) *Server {
	return &Server{
		keypair:  keypair,
		port:     port,
		stopFake: make(chan bool),
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/ws", s.handleWebSocket)

	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("🏠 IoT Dashboard running on ws://localhost%s/ws\n", addr)
	fmt.Println("💬 Waiting for connection...")
	fmt.Println()

	return http.ListenAndServe(addr, nil)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	s.connMutex.Lock()
	s.conn = conn
	s.connMutex.Unlock()

	fmt.Println("✓ Client connected!")

	// Exchange public keys
	if err := s.exchangeKeys(); err != nil {
		log.Printf("Key exchange failed: %v", err)
		conn.Close()
		return
	}

	fmt.Println("🔐 Secure channel established")
	fmt.Println("📊 Generating fake IoT traffic...")
	fmt.Println()
	fmt.Println("Type your message and press Enter to send:")
	fmt.Println()

	// Start fake traffic generator
	go s.generateFakeTraffic()

	// Start message receiver
	go s.receiveMessages()

	// Handle user input
	s.handleInput()
}

func (s *Server) exchangeKeys() error {
	// Send our public key
	keyMsg := map[string]string{
		"type":       "handshake",
		"public_key": crypto.PublicKeyToBase64(s.keypair.PublicKey()),
	}

	if err := s.conn.WriteJSON(keyMsg); err != nil {
		return fmt.Errorf("sending public key: %w", err)
	}

	// Receive client's public key
	var clientKey map[string]string
	if err := s.conn.ReadJSON(&clientKey); err != nil {
		return fmt.Errorf("receiving public key: %w", err)
	}

	if clientKey["type"] != "handshake" {
		return fmt.Errorf("unexpected message type")
	}

	pubKey, err := crypto.Base64ToPublicKey(clientKey["public_key"])
	if err != nil {
		return fmt.Errorf("invalid client public key: %w", err)
	}

	s.clientPubKey = pubKey
	return nil
}

func (s *Server) generateFakeTraffic() {
	ticker := time.NewTicker(time.Duration(30+rand.Intn(30)) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopFake:
			return
		case <-ticker.C:
			msg := iot.GenerateFakeTraffic()
			s.sendMessage(msg)

			// Show fake traffic occasionally (10% chance)
			if rand.Float64() < 0.1 {
				fmt.Printf("\r\033[K[IoT] %s: %v\n> ", msg.DeviceID, msg.Data)
			}

			// Reset ticker with random interval
			ticker.Reset(time.Duration(30+rand.Intn(30)) * time.Second)
		}
	}
}

func (s *Server) sendMessage(msg iot.Message) error {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()

	if s.conn == nil {
		return fmt.Errorf("no connection")
	}

	return s.conn.WriteJSON(msg)
}

func (s *Server) receiveMessages() {
	for {
		var msg iot.Message
		err := s.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Connection closed: %v", err)
			}
			s.cleanup()
			return
		}

		// Check if it's a real message
		if encryptedData, isReal := iot.ExtractRealMessage(msg); isReal {
			decrypted, err := s.keypair.Decrypt(encryptedData, s.clientPubKey)
			if err != nil {
				log.Printf("Decryption failed: %v", err)
				continue
			}

			fmt.Printf("\r\033[K[Them]: %s\n> ", string(decrypted))
		}
	}
}

func (s *Server) handleInput() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")

	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			fmt.Print("> ")
			continue
		}

		if text == "/quit" {
			s.cleanup()
			os.Exit(0)
		}

		// Encrypt and send
		encrypted, err := s.keypair.Encrypt([]byte(text), s.clientPubKey)
		if err != nil {
			log.Printf("Encryption failed: %v", err)
			fmt.Print("> ")
			continue
		}

		msg := iot.CreateRealMessage(encrypted)
		if err := s.sendMessage(msg); err != nil {
			log.Printf("Send failed: %v", err)
		}

		fmt.Print("> ")
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Input error: %v", err)
	}
}

func (s *Server) cleanup() {
	s.stopFake <- true
	s.connMutex.Lock()
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	s.connMutex.Unlock()
	fmt.Println()
	fmt.Println("Connection closed")
}
