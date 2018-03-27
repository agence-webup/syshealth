package jwttools

import (
	"crypto/sha1"
	"fmt"
	"math/rand"
	"time"
	"webup/syshealth"

	jwt "github.com/dgrijalva/jwt-go"
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

// GetToken returns a JWT token based on server data and a unique ID
// the ID is stored with server and allows to revoke the token
func GetToken(server syshealth.Server) (string, string, error) {
	// generate the ID
	id := fmt.Sprintf("%d%s%s", generateID(), server.Name, server.IP)
	h := sha1.New()
	h.Write([]byte(id))
	hashedID := fmt.Sprintf("%x", h.Sum(nil))

	// Create the Claims
	claims := &jwt.StandardClaims{
		Issuer: "syshealth-server",
		Id:     hashedID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte("truite"))

	return signedToken, hashedID, err
}

func generateID() int {
	return seededRand.Intn(100000)
}
