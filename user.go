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
