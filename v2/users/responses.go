package users

type AuthenticationResponse struct {
	Token     string `json:"token"`
	User      User   `json:"user,omitempty"`
	Exception string `json:"exception"`
}

type UserResponse struct {
	User      User   `json:"user"`
	Exception string `json:"exception"`
}

type TokenRenewalResponse struct {
	Token     string `json:"token"`
	Exception string `json:"exception"`
}

type UsersResponse struct {
	Users     []User `json:"users"`
	Exception string `json:"exception"`
}

type ExceptionResponse struct {
	Exception string `json:"exception"`
}
