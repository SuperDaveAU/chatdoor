package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"chatdoor/internal/client"
	"chatdoor/internal/crypto"
	"chatdoor/internal/server"
)

const banner = `
 ██████╗██╗  ██╗ █████╗ ████████╗██████╗  ██████╗  ██████╗ ██████╗ 
██╔════╝██║  ██║██╔══██╗╚══██╔══╝██╔══██╗██╔═══██╗██╔═══██╗██╔══██╗
██║     ███████║███████║   ██║   ██║  ██║██║   ██║██║   ██║██████╔╝
██║     ██╔══██║██╔══██║   ██║   ██║  ██║██║   ██║██║   ██║██╔══██╗
╚██████╗██║  ██║██║  ██║   ██║   ██████╔╝╚██████╔╝╚██████╔╝██║  ██║
 ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   ╚═════╝  ╚═════╝  ╚═════╝ ╚═╝  ╚═╝

  Covert P2P Chat - Disguised as IoT Traffic`

func main() {
	fmt.Println(banner)

	// Subcommands
	serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
	clientCmd := flag.NewFlagSet("client", flag.ExitOnError)

	// Server flags
	serverPort := serverCmd.Int("port", 8080, "Port to listen on")
	serverKeyFile := serverCmd.String("key", "", "Load existing private key (optional)")

	// Client flags
	clientAddr := clientCmd.String("addr", "", "Server address (host:port)")
	clientKeyFile := clientCmd.String("pubkey", "", "Server's public key file")
	clientKeyString := clientCmd.String("key", "", "Server's public key (base64 string)")

	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  chatdoor server [options]  - Start server mode")
		fmt.Println("  chatdoor client [options]  - Start client mode")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  chatdoor server -port 8080")
		fmt.Println("  chatdoor client -addr localhost:8080 -pubkey server.pub")
		fmt.Println("  chatdoor client -addr localhost:8080 -key \"base64key...\"")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		serverCmd.Parse(os.Args[2:])
		runServer(*serverPort, *serverKeyFile)

	case "client":
		clientCmd.Parse(os.Args[2:])
		runClient(*clientAddr, *clientKeyFile, *clientKeyString)

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runServer(port int, keyFile string) {
	var keypair *crypto.KeyPair
	var err error

	if keyFile != "" {
		// Load existing key
		fmt.Printf("📂 Loading private key from: %s\n", keyFile)
		keypair, err = crypto.LoadPrivateKey(keyFile)
		if err != nil {
			log.Fatalf("Failed to load private key: %v", err)
		}
		fmt.Println("✓ Private key loaded successfully")
	} else {
		// Generate new keypair
		fmt.Println("🔑 Generating new keypair...")
		keypair, err = crypto.GenerateKeyPair()
		if err != nil {
			log.Fatalf("Failed to generate keypair: %v", err)
		}

		// Save private key
		privKeyPath := "chatdoor_private.key"
		if err := crypto.SavePrivateKey(keypair, privKeyPath); err != nil {
			log.Fatalf("Failed to save private key: %v", err)
		}
		fmt.Printf("✓ Private key saved to: %s\n", privKeyPath)

		// Save public key
		pubKeyPath := "chatdoor_public.key"
		if err := crypto.SavePublicKey(keypair.PublicKey(), pubKeyPath); err != nil {
			log.Fatalf("Failed to save public key: %v", err)
		}
		fmt.Printf("✓ Public key saved to: %s\n", pubKeyPath)
	}

	// Display public key for sharing
	pubKeyB64 := crypto.PublicKeyToBase64(keypair.PublicKey())
	fmt.Println()
	fmt.Println(strings.Repeat("━", 60))
	fmt.Println("📤 Share this public key with your peer:")
	fmt.Println(strings.Repeat("━", 60))
	fmt.Println(pubKeyB64)
	fmt.Println(strings.Repeat("━", 60))
	fmt.Println()

	// Start server
	srv := server.New(keypair, port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func runClient(addr, keyFile, keyString string) {
	var publicKey []byte
	var err error

	// Get server's public key
	if keyFile != "" {
		fmt.Printf("📂 Loading public key from: %s\n", keyFile)
		publicKey, err = crypto.LoadPublicKeyFromFile(keyFile)
		if err != nil {
			log.Fatalf("Failed to load public key: %v", err)
		}
	} else if keyString != "" {
		fmt.Println("🔑 Parsing public key from argument...")
		publicKey, err = crypto.Base64ToPublicKey(keyString)
		if err != nil {
			log.Fatalf("Invalid public key: %v", err)
		}
	} else {
		// Prompt for public key
		fmt.Println("Enter server's public key (base64):")
		fmt.Print("> ")
		var input string
		fmt.Scanln(&input)
		publicKey, err = crypto.Base64ToPublicKey(input)
		if err != nil {
			log.Fatalf("Invalid public key: %v", err)
		}
	}

	// Validate public key
	if !crypto.ValidatePublicKey(publicKey) {
		log.Fatal("Invalid public key format")
	}

	fmt.Println("✓ Valid public key loaded")

	// Prompt for address if not provided
	if addr == "" {
		fmt.Println("Enter server address (host:port):")
		fmt.Print("> ")
		fmt.Scanln(&addr)
	}

	// Start client
	cli := client.New(publicKey, addr)
	if err := cli.Start(); err != nil {
		log.Fatalf("Client error: %v", err)
	}
}
