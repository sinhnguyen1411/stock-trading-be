package user

type verificationEmailPayload struct {
	Email   string `json:"email"`
	Token   string `json:"token"`
	Purpose string `json:"purpose"`
}
