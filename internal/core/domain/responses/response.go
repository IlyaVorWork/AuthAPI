package responses

type Error struct {
	Error string `json:"error"`
}

type AddRolesError struct {
	Error       string            `json:"error"`
	RolesStatus map[string]string `json:"roles status"`
}

type AddRolesSuccess struct {
	Login       string            `json:"login"`
	RolesStatus map[string]string `json:"roles status"`
}

type LoginSuccess struct {
	Token string `json:"access_token"`
}

type GetUserSuccess struct {
	Id       string   `json:"id"`
	Login    string   `json:"login"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

type GetFileListSuccess struct {
	List []string `json:"files list"`
}
