package domain

type Session struct {
	SignedIn       bool
	SessionUsable  bool
	Account        string
	MailConsented  bool
	TeamsConsented bool
}

func SignedOut() Session {
	return Session{}
}
