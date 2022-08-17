package auth

type Socks5Auth struct {
	Username string
	Password string
}

func NewSocksAuth(username string, password string) Socks5Auth {
	return Socks5Auth{
		Username: username,
		Password: password,
	}
}

func (a Socks5Auth) Auth(username string, password string) bool {
	return a.Username == username && a.Password == password
}
