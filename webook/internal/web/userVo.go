package web

type LoginSMSReq struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

type SignUpReq struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type LoginJWTReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserEditReq struct {
	Nickname string `json:"nickname"`
	Birthday string `json:"birthday"`
	AboutMe  string `json:"aboutMe"`
}

type SendSMSCodeReq struct {
	Phone string `json:"phone"`
}
