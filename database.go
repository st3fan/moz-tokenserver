// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/

package main

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type User struct {
	Uid             string
	Email           string
	Node            string
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
	return nil, nil
}

func (ds *DatabaseSession) AllocateUser(email string, generation int, clientState string) (*User, error) {
	return nil, nil
}

func (ds *DatabaseSession) UpdateUser(email string, newGeneration int, newClientState string) (*User, error) {
	return nil, nil
}
