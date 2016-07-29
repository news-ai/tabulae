package emails

type Email struct {
	// Email details
	Sender  string
	To      []string
	Subject string
	Body    string

	// User details
	FirstName string
}

const confirmMessage = `
Thank you for signing up on NewsAI Tabulae! 
`
