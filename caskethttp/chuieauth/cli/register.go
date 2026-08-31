package main

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strconv"
	"time"
)

type databaseUser struct {
	Username string    `gorethink:"username"`
	Hash     []byte    `gorethink:"password,omitempty"`
	Expiry   time.Time `gorethink:"expiry"`
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: register username password expiry")
		return
	}

	username, password := os.Args[1], os.Args[2]
	expiryNum, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Println(err)
		return
	}

	expiry := time.Unix(int64(expiryNum), 0)
	if expiryNum == 0 {
		expiry = time.Time{}
	}

	session, err := r.Connect(r.ConnectOpts{
		Address:  "127.0.0.1:28015",
		AuthKey:  os.Getenv("DBAUTH"),
		Database: "chuieauth",
		MaxIdle:  1,
		MaxOpen:  1,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password),
		bcrypt.DefaultCost)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := r.Table("users").Insert(databaseUser{
		Username: username,
		Hash:     hash,
		Expiry:   expiry,
	}, r.InsertOpts{
		Conflict: "update",
	}).RunWrite(session)

	if err != nil {
		fmt.Println(err)
		return
	}

	if resp.Inserted == 1 {
		fmt.Println("Account created")
	} else if resp.Replaced == 1 || resp.Updated == 1 {
		fmt.Println("Account updated")
	} else {
		fmt.Println("Unknown changes made:")
		fmt.Println(resp)
	}
}
