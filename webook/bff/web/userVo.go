package web

type (
	LoginSMSReq struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}

	SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	LoginJWTReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	UserEditReq struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}

	SendSMSCodeReq struct {
		Phone string `json:"phone"`
	}
)
