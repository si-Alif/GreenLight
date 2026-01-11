package mailer

import (
	"bytes"
	"embed"
	"time"

	"github.com/wneessen/go-mail"

	ht "html/template"
	tt "text/template"
)

// the <path> in the directive should be relative to the file where the directive is placed

//go:embed "templates"
var templateFS embed.FS

// central unit of mail service
type Mailer struct {
	client *mail.Client // client instance would be used to connect to a SMTP server
	sender string
}

// new creates a email provider instance using all these values and return the provider instance
func New(host string, port int, username, password, sender string) (*Mailer, error) {
	client, err := mail.NewClient(
		host,
		mail.WithSMTPAuth(mail.SMTPAuthLogin),
		mail.WithUsername(username),
		mail.WithPassword(password),
		mail.WithTimeout(5*time.Second),
	)

	if err != nil {
		return nil, err
	}

	mailer := &Mailer{
		client: client,
		sender: sender,
	}

	return mailer, nil

}

// Send() method takes the recipient address , template string and data to embed in it
func (ml *Mailer) Send(recipient string, templateFile string, data any) error {

	// templateFS is the embedded directory tree (like a virtual folder).
	// templateFile is the path/key you use to fetch a specific file from that tree.

	// refer to MailService.md for details if you get confused later
	txtTmpl, err := tt.New("").ParseFS(templateFS, "templates/"+templateFile)

	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = txtTmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)
	err = txtTmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// html parsing
	htmlTmpl, err := ht.New("").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)
	err = htmlTmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	msg := mail.NewMsg()
	err = msg.To(recipient)
	if err != nil {
		return err
	}

	err = msg.From(ml.sender)
	if err != nil {
		return err
	}

	msg.Subject(subject.String())
	msg.SetBodyString(mail.TypeTextPlain, plainBody.String())
	msg.AddAlternativeString(mail.TypeTextHTML, htmlBody.String())

	//DialAndSend() takes the message to send , opens activated connection with SMTP server , sends the message and closes the connection
	// return ml.client.DialAndSend(msg)

	// try for 3s
	for i := 1 ; i<= 3 ; i++{
		err = ml.client.DialAndSend(msg)
		if err == nil {
			return nil
		}

		if i !=3 {
			time.Sleep(800 * time. Millisecond)
		}

	}

  return err

}
