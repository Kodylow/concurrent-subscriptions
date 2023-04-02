package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"sync"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Mail struct {
	Domain      string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
	Wait        *sync.WaitGroup
	MailerChan  chan Message
	ErrorChan   chan error
	DoneChan    chan bool
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []string
	Data        any
	DataMap     map[string]any
	Template    string
}

func (app *Config) listenForMail() {
	app.InfoLog.Println("Listening for mail")
	for {
		select {
		case msg := <-app.Mailer.MailerChan:
			app.InfoLog.Println("Sending email")
			go app.Mailer.sendMail(msg, app.Mailer.ErrorChan)
		case err := <-app.Mailer.ErrorChan:
			app.ErrorLog.Println("Error sending email: ", err)
		case <-app.Mailer.DoneChan:
			app.Mailer.Wait.Done()
			return
		}
	}
}

// sendMail sends the message as an email
func (m *Mail) sendMail(msg Message, errorChan chan error) {
	log.Println("In sendMail...")
	defer m.Wait.Done()
	formattedMessage, plainTextMessage := m.buildMessages(&msg, errorChan)
	email := m.buildEmail(&msg, plainTextMessage, formattedMessage)
	server := m.setupMailServer()

	log.Println("Connecting to mail server...")
	smtpClient, err := server.Connect()
	if err != nil {
		log.Println("Error connecting to mail server")
		errorChan <- err
	}
	log.Println("Connected to mail server...")

	if err := email.Send(smtpClient); err != nil {
		log.Println("Error sending email: ", err)
		errorChan <- err
	}
}

// buildEmail build the email object with the message and attachments
func (*Mail) buildEmail(msg *Message, plainTextMessage string, formattedMessage string) *mail.Email {
	log.Println("In buildEmail")
	email := mail.NewMSG()
	email.SetFrom(msg.From).AddTo(msg.To).SetSubject(msg.Subject)
	email.SetBody(mail.TextPlain, plainTextMessage).AddAlternative(mail.TextHTML, formattedMessage)
	for _, attachment := range msg.Attachments {
		email.AddAttachment(attachment)
	}
	return email
}

// buildMessages builds the HTML and plain text messages for the email
func (m *Mail) buildMessages(msg *Message, errorChan chan error) (string, string) {
	log.Println("In buildMessages")
	if msg.Template == "" {
		msg.Template = "default"
	}
	if msg.From == "" {
		msg.From = m.FromAddress
	}
	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	data := map[string]any{
		"message": msg.Data,
	}

	msg.DataMap = data
	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		errorChan <- err
	}

	plainTextMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		errorChan <- err
	}

	return formattedMessage, plainTextMessage
}

// setupMailServer sets up the mail server
func (m *Mail) setupMailServer() *mail.SMTPServer {
	log.Println("In setupMailServer")
	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second
	return server
}

// buildHTMLMessage builds the HTML message for the email
func (m *Mail) buildHTMLMessage(msg *Message) (string, error) {
	log.Println("In buildHTMLMessage")
	templateToRender := fmt.Sprintf("/Users/kodylow/Documents/github/golang/concurrent-subscriptions/cmd/web/templates/%s.html.gohtml", msg.Template)

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		log.Println("Error parsing template: ", err)
		return "", err
	}

	var tpl bytes.Buffer
	if err := t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		log.Println("Error executing template: ", err)
		return "", err
	}

	formattedMessage := tpl.String()
	formattedMessage, err = m.inlineCSS(formattedMessage)
	if err != nil {
		log.Println("Error inlining CSS: ", err)
		return "", err
	}

	return formattedMessage, nil
}

// buildPlainTextMessage builds the plain text message for the email
func (m *Mail) buildPlainTextMessage(msg *Message) (string, error) {
	log.Println("In buildPlainTextMessage")
	templateToRender := fmt.Sprintf("/Users/kodylow/Documents/github/golang/concurrent-subscriptions/cmd/web/templates/%s.plain.gohtml", msg.Template)

	t, err := template.New("email-plain").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err := t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	plainMessage := tpl.String()

	return plainMessage, nil
}

func (m *Mail) inlineCSS(s string) (string, error) {
	log.Println("In inlineCSS")
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}

// getEncryption returns the encryption type for the mail server
func (m *Mail) getEncryption(e string) mail.Encryption {
	log.Println("In getEncryption")
	switch e {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSLTLS
	case "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}
