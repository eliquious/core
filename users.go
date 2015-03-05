package core

const (
	DEFAULT_USER = "default.user"
)

type BaseUser struct {
	Username       string
	RequestId      int
	PrimaryEmail   string
	SaltedPassword string
	Salt           string
	CookieBlock    string
}
