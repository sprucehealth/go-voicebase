package erx

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"math/rand"
	"time"
)

var (
	alphanum = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
)

func generateSingleSignOn(clinicKey, userId, clinicId string) singleSignOn {
	rand.Seed(time.Now().UnixNano())

	singleSignOn := singleSignOn{}

	// STEP 1: Create a random phrase 32 characters long in UTF8
	phrase := generateRandomAlphaNumString(32)
	singleSignOn.Code = string(createSingleSignOn(phrase, clinicKey))
	singleSignOn.UserIdVerify = string(createSingleSignOnUserIdVerify(phrase, clinicKey, userId))
	singleSignOn.ClinicId = clinicId
	singleSignOn.UserId = userId
	singleSignOn.PhraseLength = 32
	return singleSignOn
}

func generateRandomAlphaNumString(n int) []byte {
	randomBytes := make([]byte, n)
	for i := 0; i < n; i++ {
		randomBytes[i] = alphanum[rand.Intn(len(alphanum))]
	}
	return randomBytes
}

// Steps to create the singleSignOnUserIdVerify is spelled out in the
func createSingleSignOnUserIdVerify(phrase []byte, clinicKey, userId string) []byte {

	// STEP 1: first 22 characters from phrase
	first22Pharse := phrase[:22]

	// STEPS 2-5: Compute the hash of the userId + first 22 characters from phrase + key
	sha512Hash := sha512.New()
	sha512Hash.Write([]byte(userId))
	sha512Hash.Write(first22Pharse)
	sha512Hash.Write([]byte(clinicKey))
	hashedBytes := sha512Hash.Sum(nil)

	// STEP 6: Get a Base64String out of the hash that you created
	encodedBytes := make([]byte, base64.StdEncoding.EncodedLen(len(hashedBytes)))
	base64.StdEncoding.Encode(encodedBytes, hashedBytes)

	// STEP 7: if there are two = signs at the end, then remove them.
	base64Encoded := removeTwoEqualSignsIfPresent(encodedBytes)
	return base64Encoded
}

func createSingleSignOn(phrase []byte, clinicKey string) []byte {

	// STEPS 2 - 4: Compute the hash of the phrase concatenated with the key
	sha512Hash := sha512.New()
	sha512Hash.Write(phrase)
	sha512Hash.Write([]byte(clinicKey))
	hashedBytes := sha512Hash.Sum(nil)

	// STEP 5: Get a Base64String out of the hash that you created
	encodedBytes := make([]byte, base64.StdEncoding.EncodedLen(len(hashedBytes)))
	base64.StdEncoding.Encode(encodedBytes, hashedBytes)

	// STEP 6: If there are two = signs at the end, then remove them.
	base64Encoded := removeTwoEqualSignsIfPresent(encodedBytes)

	// STEP 7: Prepend the same random phrase from step 2 to your code.
	return append(phrase, base64Encoded...)
}

func removeTwoEqualSignsIfPresent(str []byte) []byte {
	if bytes.Compare(str[len(str)-2:len(str)], []byte("==")) == 0 {
		return str[:len(str)-2]
	}
	return str
}
