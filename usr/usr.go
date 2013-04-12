package usr

const(
NONE = 0
F_SUB = 1
T_SUB = 2
FT_SUB = 3
U_SUB = 4
UF_SUB = 5
UT_SUB = 6
ALL_SUB = 7
)

type User struct {
	Name     string "_id,omitempty"
	Password string "password"
	Salt     string "salt"
	Access int "access"
}

func (u *User) hasAccess(access int)(ret bool){
switch access{
	case NONE: ret = u.Access == NONE 
	case F_SUB: ret = EqualsOne(u.Access, F_SUB, FT_SUB, UF_SUB, ALL_SUB)
	case T_SUB: ret = EqualsOne(u.Access, T_SUB, FT_SUB, UT_SUB, ALL_SUB)
	case U_SUB: ret = EqualsOne(u.Access, U_SUB, UF_SUB, UT_SUB, ALL_SUB)
	}
return ret
}

func (u *User) CheckSubmit(mode string)(ret bool){
	if mode == "SINGLE" || mode == "MULTIPLE"{
		ret = u.hasAccess(F_SUB)
	} else if mode == "TEST"{
		ret = u.hasAccess(T_SUB)
	} else if mode == "UPDATE"{
		ret = u.hasAccess(U_SUB)
	}
	return ret
}


func NewUser(uname, pword, salt string) *User {
	return &User{uname, pword, salt, F_SUB}
}


func EqualsOne(test interface{}, args... interface{})(eq bool){
	for _, arg := range args{
		if eq = test == arg; eq{
			break
		}
	}
	return eq
}