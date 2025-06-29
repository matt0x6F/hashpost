package main

import (
	"fmt"

	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/matt0x6f/hashpost/internal/poc"
)

func main() {
	fmt.Println("=== Identity-Based Encryption (IBE) Proof of Concept ===")
	fmt.Println("Demonstrating pseudonymous social platform with administrative correlation")

	// Initialize the IBE system
	ibeSystem := ibe.NewIBESystem()

	// Initialize storage
	db := poc.NewMemoryDB()

	// Run the demonstration
	demo := NewDemo(ibeSystem, db)
	demo.Run()
}
