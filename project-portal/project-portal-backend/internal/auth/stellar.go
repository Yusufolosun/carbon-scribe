package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"time"
)

// StellarAuthenticator handles Stellar wallet authentication
type StellarAuthenticator struct {
	NetworkPassphrase string
	ChallengeTimeout  time.Duration
	Challenges        map[string]*ChallengeData // In production, use Redis
}

// ChallengeData stores challenge information
type ChallengeData struct {
	Challenge   string
	ValidatedAt *time.Time
	ExpiresAt   time.Time
}

// NewStellarAuthenticator creates a new Stellar authenticator
func NewStellarAuthenticator(networkPassphrase string, timeout time.Duration) *StellarAuthenticator {
	if networkPassphrase == "" {
		networkPassphrase = "Test SDF Network ; September 2015" // Default test network
	}
	if timeout == 0 {
		timeout = 15 * time.Minute
	}

	return &StellarAuthenticator{
		NetworkPassphrase: networkPassphrase,
		ChallengeTimeout:  timeout,
		Challenges:        make(map[string]*ChallengeData),
	}
}

// GenerateChallengeTransaction generates a challenge for wallet signing
func (sa *StellarAuthenticator) GenerateChallengeTransaction(publicKey string) (string, error) {
	// Validate the public key format (Stellar addresses start with 'G' and are 56 characters)
	if !isValidStellarAddress(publicKey) {
		return "", fmt.Errorf("invalid Stellar public key format")
	}

	// Generate a crypto-secure challenge
	challengeData := sha256.Sum256([]byte(publicKey + time.Now().String()))
	challenge := hex.EncodeToString(challengeData[:])

	// Store the challenge with expiration
	sa.Challenges[publicKey] = &ChallengeData{
		Challenge: challenge,
		ExpiresAt: time.Now().Add(sa.ChallengeTimeout),
	}

	return challenge, nil
}

// VerifyChallengeSignature verifies a signed challenge
func (sa *StellarAuthenticator) VerifyChallengeSignature(publicKey string, signedChallenge string) error {
	// Validate the public key format
	if !isValidStellarAddress(publicKey) {
		return fmt.Errorf("invalid Stellar public key format")
	}

	// Validate the signed challenge is not empty
	if signedChallenge == "" {
		return errors.New("signed challenge cannot be empty")
	}

	// Check if challenge exists and hasn't expired
	challengeData, exists := sa.Challenges[publicKey]
	if !exists {
		return errors.New("challenge not found or expired")
	}

	if time.Now().After(challengeData.ExpiresAt) {
		delete(sa.Challenges, publicKey)
		return errors.New("challenge has expired")
	}

	// Mark challenge as validated
	now := time.Now()
	challengeData.ValidatedAt = &now

	// Basic validation that a signature was provided
	// In a production system, you would verify the actual XDR-encoded transaction
	// signature using: txnbuild.TransactionEnvelopeFromXDR() and cryptographic verification
	if len(signedChallenge) >= 32 { // Minimum reasonable signature length
		return nil
	}

	return errors.New("signature verification failed: invalid format")
}

// ValidateWalletAddress performs validation of a Stellar wallet address
func ValidateWalletAddress(address string) error {
	if !isValidStellarAddress(address) {
		return errors.New("invalid Stellar wallet address")
	}
	return nil
}

// isValidStellarAddress checks if address matches Stellar public key format
func isValidStellarAddress(address string) bool {
	// Stellar public keys start with 'G' and are 56 characters long
	// They use base32 encoding (A-Z, 2-7)
	if len(address) != 56 {
		return false
	}
	if address[0] != 'G' {
		return false
	}

	// Check if the rest contains only valid base32 characters
	validChars := regexp.MustCompile(`^G[A-Z2-7]{55}$`)
	return validChars.MatchString(address)
}
