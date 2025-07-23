package users

type JWTClaim struct {
	UserID       string `json:"userid"`
	EmailAddress string `json:"emailAddress"`
}

type UserName struct {
	FirstName  string `json:"firstName"`
	MiddleName string `json:"middleName,omitempty"`
	LastName   string `json:"lastName"`
}
