package core

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"
	"log"

	"github.com/gorilla/securecookie"
)

// Global cookie hash
var CookieHashKey = securecookie.GenerateRandomKey(64)

// SaltSize is the size of the salt for encrypting passwords
const SaltSize = 16

// GenerateSalt creates a new salt and encodes the given password.
// It returns the new salt, the ecrypted password and a possible error
func GenerateSalt(secret []byte) ([]byte, []byte, error) {
	buf := make([]byte, SaltSize, SaltSize+sha256.Size)
	_, err := io.ReadFull(rand.Reader, buf)

	if err != nil {
		log.Printf("random read failed: %v", err)
		return nil, nil, err
	}

	hash := sha256.New()
	hash.Write(buf)
	hash.Write(secret)
	return buf, hash.Sum(nil), nil
}

// SecureCompare compares salted passwords in constant time
// http://stackoverflow.com/questions/20663468/secure-compare-of-strings-in-go
func SecureCompare(given, actual []byte) bool {
	if subtle.ConstantTimeEq(int32(len(given)), int32(len(actual))) == 1 {
		return subtle.ConstantTimeCompare(given, actual) == 1
	}

	/* Securely compare actual to itself to keep constant time, but always return false */
	return subtle.ConstantTimeCompare(actual, actual) == 1 && false
}

// UUID4 generates a random UUID according to RFC 4122
func UUID4() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}

	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80

	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

// createDefaultUser checks for the default user and if it does not exist it is
// created
func createDefaultUser(users Keyspace) error {

	// check if default user exists
	exists, err := users.Contains("default.user@example.com")
	if err != nil {

		// error accessing database
		log.Printf("could not perform contains check on database: %v\n", err)
		return err
	} else if exists {

		// user already exists
		return nil
	}

	// create default user
	user := BaseUser{}
	user.Username = "default.user@example.com"
	user.PrimaryEmail = "default.user@example.com"

	// password
	secret := []byte("password")
	salt, saltedpw, err := GenerateSalt(secret)
	if err != nil {
		log.Printf("could not generate salt user: %v\n", err)
		return err
	}

	// encode salt and salted password with Base64
	user.SaltedPassword = base64.StdEncoding.EncodeToString(saltedpw)
	user.Salt = base64.StdEncoding.EncodeToString(salt)

	// user specific cookie block
	user.CookieBlock = base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))

	// save user
	return SaveUser(&user, users)
}

// authenticate validates a user's password with the salted password that has been stored
func authenticate(user BaseUser, password string) bool {
	log.Printf("Authenticating user: %#v\n", user.Username)

	// base64 encoded salted password
	combined, err := base64.StdEncoding.DecodeString(user.SaltedPassword)
	if err != nil {
		// could not decode salted password
		log.Printf("Could not decode salted password: (%s) %v", user.Username, err)
		return false
	}

	// base64 encoded salt
	salt, err := base64.StdEncoding.DecodeString(user.Salt)
	if err != nil {
		// could not decode salt
		log.Printf("Could not decode salt: (%s) %v", user.Username, err)
		return false
	}

	// test hash
	hash := sha256.New()
	hash.Write(salt)
	hash.Write([]byte(password))

	// compare byte strings
	log.Println("Comparing hashes..")
	return SecureCompare(hash.Sum(nil), combined)
}
