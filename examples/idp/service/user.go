package idp

type User struct {
	Ident string
	Name  string
	Pass  string
}

func (u User) ID() string {
	return u.Ident
}

func (u User) String() string {
	return u.Name
}
