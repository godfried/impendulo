package user

import (
	"labix.org/v2/mgo/bson"
)

const (
	NONE    = 0
	F_SUB   = 1
	T_SUB   = 2
	FT_SUB  = 3
	U_SUB   = 4
	UF_SUB  = 5
	UT_SUB  = 6
	ALL_SUB = 7
)

const (
	SINGLE  = "file_remote"
	ARCHIVE = "archive_remote"
	TEST    = "archive_test"
	UPDATE  = "update"
	ID      = "_id"
	PWORD   = "password"
	SALT    = "salt"
	ACCESS  = "access"
)

type User struct {
	Name     string "_id"
	Password string "password"
	Salt     string "salt"
	Access   int    "access"
}

func (u *User) hasAccess(access int) bool {
	switch access {
	case NONE:
		return u.Access == NONE
	case F_SUB:
		return EqualsOne(u.Access, F_SUB, FT_SUB, UF_SUB, ALL_SUB)
	case T_SUB:
		return EqualsOne(u.Access, T_SUB, FT_SUB, UT_SUB, ALL_SUB)
	case U_SUB:
		return EqualsOne(u.Access, U_SUB, UF_SUB, UT_SUB, ALL_SUB)
	}
	return false
}

func ReadUser(umap bson.M) *User {
	name := umap[ID].(string)
	pword := umap[PWORD].(string)
	salt := umap[SALT].(string)
	access := umap[ACCESS].(int)
	return &User{name, pword, salt, access}
}
func (u *User) CheckSubmit(mode string) bool {
	if mode == SINGLE || mode == ARCHIVE {
		return u.hasAccess(F_SUB)
	} else if mode == TEST {
		return u.hasAccess(T_SUB)
	} else if mode == UPDATE {
		return u.hasAccess(U_SUB)
	}
	return false
}

func NewUser(uname, pword, salt string) *User {
	return &User{uname, pword, salt, F_SUB}
}

func EqualsOne(test interface{}, args ...interface{}) bool {
	for _, arg := range args {
		if test == arg {
			return true
		}
	}
	return false
}
