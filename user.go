// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/

package main

type User struct {
	Email           string
	Uid             string
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

func GetUser(email string) (*User, error) {
	return nil, nil
}

func AllocateUser(email string, generation int, clientState string) (*User, error) {
	return nil, nil
}

func UpdateUser(email string, newGeneration int, newClientState string) (*User, error) {
	return nil, nil
}
