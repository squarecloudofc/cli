package square

type ResponseUser struct {
	Applications []UserApplication `json:"applications"`

	User struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Locale string `json:"locale"`
		Email  string `json:"email"`
		Plan   struct {
			Name   string `json:"name"`
			Memory struct {
				Limit     int `json:"limit"`
				Available int `json:"available"`
				Used      int `json:"used"`
			} `json:"memory"`
			Duration int64 `json:"duration"`
		} `json:"plan"`
	} `json:"user"`
}

type UserApplication struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Lang    string `json:"lang"`
	Cluster string `json:"cluster"`
	Avatar  string `json:"avatar"`
	RAM     int    `json:"ram"`
}
