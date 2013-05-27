package db

import(
	"github.com/godfried/cabanga/user"
"fmt"
)

//GetUserById retrieves a user matching the given id from the active database. 
func GetUserById(id interface{}) (*user.User, error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(USERS)
	var ret *user.User
	err := c.FindId(id).One(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving user %q from db", err, id)
	}
	return ret, nil
}

//AddUser adds a new user to the active database.
func AddUser(u *user.User) error {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(USERS)
	err := c.Insert(u)
	if err != nil {
		return fmt.Errorf("Encountered error %q when adding user %q to db", err, u)
	}
	return nil
}

//AddUsers adds new users to the active database.
func AddUsers(users ...*user.User) error {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(USERS)
	err := c.Insert(users)
	if err != nil {
		return fmt.Errorf("Encountered error %q when adding users %q to db", err, users)
	}
	return nil
}