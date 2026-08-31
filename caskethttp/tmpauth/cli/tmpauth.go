package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println(usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "generate":
		privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			log.Fatalln("failed to generate private key:", err)
		}

		mPriv, err := x509.MarshalPKCS8PrivateKey(privKey)
		if err != nil {
			log.Fatalln("failed to marshal private key:", err)
		}

		buf := new(bytes.Buffer)
		err = pem.Encode(buf, &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: mPriv,
		})
		if err != nil {
			log.Fatalln("failed to encode pem:", err)
		}

		fmt.Println("Your private key in PEM is:")
		fmt.Println(buf.String())

		pub := elliptic.Marshal(elliptic.P256(), privKey.X, privKey.Y)
		priv := privKey.D.Bytes()

		fmt.Println("Your private key in tmpauth minified format is:")
		fmt.Println(base64.StdEncoding.EncodeToString(priv) + "." + base64.StdEncoding.EncodeToString(pub))
		fmt.Println("")
		fmt.Println("#########################################################################")
		fmt.Println("")
		fmt.Println("Your public key in PEM is:")

		mPub, err := x509.MarshalPKIXPublicKey(&(privKey.PublicKey))
		if err != nil {
			log.Fatalln("failed to marshal public key:", err)
		}

		buf.Reset()
		err = pem.Encode(buf, &pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: mPub,
		})

		fmt.Println(buf.String())

		fmt.Println("Your Casket plugin (tmpauth minified) compatible public key is:")
		fmt.Println(base64.StdEncoding.EncodeToString(pub))
	case "convert":
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}

		for len(data) > 4 {
			var block *pem.Block
			block, data = pem.Decode(data)
			switch block.Type {
			case "PRIVATE KEY":
				privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
				if err != nil {
					log.Fatalln("failed to parse private key:", err)
				}

				ecKey := privKey.(*ecdsa.PrivateKey)
				pub := elliptic.Marshal(elliptic.P256(), ecKey.X, ecKey.Y)
				priv := ecKey.D.Bytes()
				fmt.Println(base64.StdEncoding.EncodeToString(priv) + "." + base64.StdEncoding.EncodeToString(pub))
			case "PUBLIC KEY":
				pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
				if err != nil {
					log.Fatalln("failed to parse public key:", err)
				}

				ecKey := pubKey.(*ecdsa.PublicKey)
				pub := elliptic.Marshal(elliptic.P256(), ecKey.X, ecKey.Y)

				fmt.Println(base64.StdEncoding.EncodeToString(pub))
			default:
				log.Fatalln("unsupported PEM block type:", block.Type)
			}
		}
	case "create-secret":
		privateKey := os.Getenv("TMPAUTH_PRIVATE_KEY")
		if privateKey == "" {
			log.Fatalln("you must provide a tmpauth private key via env variable TMPAUTH_PRIVATE_KEY")
		}

		keyParts := strings.SplitN(privateKey, ".", 2)
		priv, err := base64.StdEncoding.DecodeString(keyParts[0])
		if err != nil {
			panic(err)
		}
		pub, err := base64.StdEncoding.DecodeString(keyParts[1])
		if err != nil {
			panic(err)
		}

		privNum := new(big.Int).SetBytes(priv)
		x, y := elliptic.Unmarshal(elliptic.P256(), pub)

		privKey := &ecdsa.PrivateKey{
			D: privNum,
			PublicKey: ecdsa.PublicKey{
				Curve: elliptic.P256(),
				X:     x,
				Y:     y,
			},
		}

		idBuf := make([]byte, 16)
		_, err = rand.Read(idBuf)
		if err != nil {
			panic(err)
		}

		secretBuf := make([]byte, 32)
		_, err = rand.Read(secretBuf)
		if err != nil {
			panic(err)
		}

		token, err := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"secret": base64.StdEncoding.EncodeToString(secretBuf),
			"sub":    hex.EncodeToString(idBuf),
			"iss":    "auth.tmpim.pw:central",
			"aud":    "auth.tmpim.pw:server:key:" + hex.EncodeToString(idBuf),
		}).SignedString(privKey)
		if err != nil {
			panic(err)
		}

		fmt.Println(token)
	}
}

var usage = `Usage: tmpauth <subcommand>
Subcommands:
	generate
		generates a PEM private key and a minified tmpauth public key
	convert
		stdins a PEM encoded public or private key, stdouts it to minified tmpauth key format
	create-secret
		creates a new server-secret token with a random ID and secret using the private key
		specified by environment variable TMPAUTH_PRIVATE_KEY.
		paste the token into https://jwt.io to inspect the claims/view the client ID and secret.
`
