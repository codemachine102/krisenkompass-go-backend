package service

import (
	"bytes"
	"fmt"
	"net/smtp"
	"strconv"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/config"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/pkg/multitemplate"
)

type EmailService struct {
}

func NewEmailService() *EmailService {
	return &EmailService{}
}

func (s *EmailService) SendUserVerificatonLink(emailTo, token string) error {
	mail := config.Get().Mail
	auth := smtp.PlainAuth("", mail.SmtpUsername, mail.SmtpPassword, mail.SmtpServer)

	var body bytes.Buffer
	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "Bestätigen Sie Ihr KRISENKOMPASS®-Konto"
	body.Write([]byte(fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n%s\n\n", mail.FromEmail, emailTo, subject, mimeHeaders)))

	link := config.Get().App.Client + "verification/" + token
	err := multitemplate.Render(&body, "userVerification", link)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = smtp.SendMail(mail.SmtpServer+":"+mail.SmtpPort, auth, mail.FromEmail, []string{emailTo}, body.Bytes())
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (s *EmailService) SendUserVerificatonLinkWithInvite(emailTo, token string, organizationName string) error {
	mail := config.Get().Mail
	auth := smtp.PlainAuth("", mail.SmtpUsername, mail.SmtpPassword, mail.SmtpServer)

	var body bytes.Buffer
	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "KRISENKOMPASS®: Einladung " + organizationName
	body.Write([]byte(fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n%s\n\n", mail.FromEmail, emailTo, subject, mimeHeaders)))

	link := config.Get().App.Client + "verification/" + token
	err := multitemplate.Render(&body, "userVerificationWithInvite", map[string]interface{}{
		"link": link,
		"name": organizationName,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = smtp.SendMail(mail.SmtpServer+":"+mail.SmtpPort, auth, mail.FromEmail, []string{emailTo}, body.Bytes())
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (s *EmailService) SendPasswordResetLink(emailTo, token string) error {
	mail := config.Get().Mail
	auth := smtp.PlainAuth("", mail.SmtpUsername, mail.SmtpPassword, mail.SmtpServer)

	var body bytes.Buffer
	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "KRISENKOMPASS®: Passwort zurücksetzen"
	body.Write([]byte(fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n%s\n\n", mail.FromEmail, emailTo, subject, mimeHeaders)))

	link := config.Get().App.Client + "password-reset/" + token
	err := multitemplate.Render(&body, "userPasswordReset", link)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = smtp.SendMail(mail.SmtpServer+":"+mail.SmtpPort, auth, mail.FromEmail, []string{emailTo}, body.Bytes())
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (s *EmailService) SendUserInvite(emailTo string, organizationID int64, organizationName string) error {
	mail := config.Get().Mail
	auth := smtp.PlainAuth("", mail.SmtpUsername, mail.SmtpPassword, mail.SmtpServer)

	var body bytes.Buffer
	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "KRISENKOMPASS®: Einladung " + organizationName
	body.Write([]byte(fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n%s\n\n", mail.FromEmail, emailTo, subject, mimeHeaders)))

	link := config.Get().App.Client + "organization/" + strconv.FormatInt(organizationID, 10)
	err := multitemplate.Render(&body, "userInvite", map[string]interface{}{
		"link": link,
		"name": organizationName,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = smtp.SendMail(mail.SmtpServer+":"+mail.SmtpPort, auth, mail.FromEmail, []string{emailTo}, body.Bytes())
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (s *EmailService) SendAdminNewOrganization(organizationID int64, plan, email, name, organizationRole, organizationName, website, city string, population int, phone, address, invoiceAddress, notes string) error {
	mail := config.Get().Mail
	auth := smtp.PlainAuth("", mail.SmtpUsername, mail.SmtpPassword, mail.SmtpServer)

	var body bytes.Buffer
	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "KRISENKOMPASS®: Neue Organisation - \"" + organizationName + "\""
	body.Write([]byte(fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n%s\n\n", mail.FromEmail, mail.AdminEmail, subject, mimeHeaders)))

	err := multitemplate.Render(&body, "adminNewOrganization", map[string]interface{}{
		"organizationID":   organizationID,
		"plan":             plan,
		"email":            email,
		"name":             name,
		"organizationRole": organizationRole,
		"organizationName": organizationName,
		"website":          website,
		"city":             city,
		"population":       population,
		"phone":            phone,
		"address":          address,
		"invoiceAddress":   invoiceAddress,
		"notes":            notes,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = smtp.SendMail(mail.SmtpServer+":"+mail.SmtpPort, auth, mail.FromEmail, []string{mail.AdminEmail}, body.Bytes())
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (s *EmailService) SendUserNewOrganization(organizationID int64, plan, email, name, organizationRole, organizationName, website, city string, population int, phone, address, invoiceAddress, notes string) error {
	mail := config.Get().Mail
	auth := smtp.PlainAuth("", mail.SmtpUsername, mail.SmtpPassword, mail.SmtpServer)

	var body bytes.Buffer
	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "KRISENKOMPASS®: Bestellung & neue Organisation - \"" + organizationName + "\""
	body.Write([]byte(fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n%s\n\n", mail.FromEmail, email, subject, mimeHeaders)))

	err := multitemplate.Render(&body, "userNewOrganization", map[string]interface{}{
		"organizationID":   organizationID,
		"plan":             plan,
		"email":            email,
		"name":             name,
		"organizationRole": organizationRole,
		"organizationName": organizationName,
		"website":          website,
		"city":             city,
		"population":       population,
		"phone":            phone,
		"address":          address,
		"invoiceAddress":   invoiceAddress,
		"notes":            notes,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = smtp.SendMail(mail.SmtpServer+":"+mail.SmtpPort, auth, mail.FromEmail, []string{email}, body.Bytes())
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
