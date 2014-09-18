// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/

package main

import (
	//"crypto/hmac"
	//"crypto/rand"
	//"crypto/sha1"
	"database/sql"
	//"encoding/hex"
	_ "github.com/lib/pq"
)

type User struct {
	Uid             string
	Email           string
	Generation      int
	ClientState     string
	OldClientStates []string
}

func (u *User) IsOldClientState(clientState string) bool {
	for _, oldClientState := range u.OldClientStates {
		if clientState == oldClientState {
			return true
		}
	}
	return false
}

type DatabaseSession struct {
	url string
	db  *sql.DB
}

func NewDatabaseSession(url string) (*DatabaseSession, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &DatabaseSession{url: url, db: db}, nil
}

func (session *DatabaseSession) Close() {
	session.db.Close()
}

func (ds *DatabaseSession) GetUser(email string) (*User, error) {
	var user User
	err := ds.db.QueryRow("select Uid,Email,Generation,Clientstate from Users where Email = $1", email).
		Scan(&user.Uid, &user.Email, &user.Generation, &user.ClientState)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return &user, nil
}

func (ds *DatabaseSession) AllocateUser(email string, generation int, clientState string) (*User, error) {
	_, err := ds.db.Exec("insert into Users (Email, Generation, ClientState) values ($1,$2,$3)", email, generation, clientState)
	if err != nil {
		return nil, err
	}
	return ds.GetUser(email)
}

func (ds *DatabaseSession) UpdateUser(email string, newGeneration int, newClientState string) (*User, error) {
	// TODO: Maybe this should run together in a transaction?
	if newGeneration != 0 {
		_, err := ds.db.Exec("update Users set Generation = $1 where Email = $2", newGeneration, email)
		if err != nil {
			return nil, err
		}
	}
	if len(newClientState) != 0 {
		_, err := ds.db.Exec("update Users set ClientState = $1 where Email = $2", newClientState, email)
		if err != nil {
			return nil, err
		}
	}
	return ds.GetUser(email)
}
