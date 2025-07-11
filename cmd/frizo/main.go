// Main func entry point.
package main

import (
	"fmt"
	"frizo-blockchain/crypto"
	"log"
)

func main() {
	wallet, priv, err := crypto.CreateWallet()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Wallet created: %s\n", wallet)
	fmt.Printf("Private key: %s\n", priv)
}
