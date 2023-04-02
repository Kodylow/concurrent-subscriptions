package main

import (
	"sync"
	"time"

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

// sendMail sends the message as an email
func (m *Mail) sendMail(msg Message, errorChan chan error) {
	formattedMessage, plainTextMessage := m.buildMessages(msg, errorChan)
	email := m.buildEmail(msg, plainTextMessage, formattedMessage)
	server := m.setupMailServer()

	smtpClient, err := server.Connect()
	if err != nil {
		errorChan <- err
	}

	if err := email.Send(smtpClient); err != nil {
		errorChan <- err
	}
}

// buildEmail build the email object with the message and attachments
func (*Mail) buildEmail(msg Message, plainTextMessage string, formattedMessage string) *mail.Email {
	email := mail.NewMSG()
	email.SetFrom(msg.From).AddTo(msg.To).SetSubject(msg.Subject)
	email.SetBody(mail.TextPlain, plainTextMessage).AddAlternative(mail.TextHTML, formattedMessage)
	for _, attachment := range msg.Attachments {
		email.AddAttachment(attachment)
	}
	return email
}

// buildMessages builds the HTML and plain text messages for the email
func (m *Mail) buildMessages(msg Message, errorChan chan error) (string, string) {
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
func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	return "", nil
}

// buildPlainTextMessage builds the plain text message for the email
func (m *Mail) buildPlainTextMessage(msg Message) (string, error) {
	return "", nil
}

// getEncryption returns the encryption type for the mail server
func (m *Mail) getEncryption(e string) mail.Encryption {
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
