// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/

package tokenserver

import (
	"encoding/json"
	"github.com/boltdb/bolt"
)

type User struct {
	Uid             uint64
	Email           string
	Generation      uint64
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
	path string
	db   *bolt.DB
}

func NewDatabaseSession(path string) (*DatabaseSession, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Users"))
		return err
	})
	return &DatabaseSession{path: path, db: db}, nil
}

func (session *DatabaseSession) Close() {
	session.db.Close()
}

func (ds *DatabaseSession) GetUser(email string) (*User, error) {
	var user *User

	err := ds.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("Users"))
		encodedUser := bucket.Get([]byte(email))
		if encodedUser == nil {
			return nil
		}

		var u User
		err := json.Unmarshal(encodedUser, &u)
		if err != nil {
			return err
		}

		user = &u

		return err
	})

	return user, err
}

func (ds *DatabaseSession) AllocateUser(email string, generation uint64, clientState string) (*User, error) {
	var user *User

	err := ds.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("Users"))

		uid, err := bucket.NextSequence()
		if err != nil {
			return err
		}

		u := &User{
			Uid:         uid,
			Email:       email,
			Generation:  generation,
			ClientState: clientState,
		}

		encodedUser, err := json.Marshal(u)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(email), encodedUser)
		if err != nil {
			return err
		}

		user = u

		return nil
	})

	return user, err
}

func (ds *DatabaseSession) UpdateUser(email string, newGeneration uint64, newClientState string) (*User, error) {
	var user *User

	err := ds.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("Users"))

		// Get the existing user record

		encodedUser := bucket.Get([]byte(email))
		if encodedUser == nil {
			return nil
		}

		var u User
		err := json.Unmarshal(encodedUser, &u)
		if err != nil {
			return err
		}

		// Update the user with the fields that have changed

		if newGeneration != 0 {
			u.Generation = newGeneration
		}

		if newClientState != "" {
			u.ClientState = newClientState
		}

		// Write the user back to the database

		encodedUser, err = json.Marshal(u)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(email), encodedUser)
		if err != nil {
			return err
		}

		user = &u

		return nil
	})

	return user, err
}
