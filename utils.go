// Copyright 2023- Germano Rizzo
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/smtp"
)

func sendEmail(to, subject, body string) error {
	smtpServer := config.SMTP.Server
	smtpPort := config.SMTP.Port

	auth := smtp.PlainAuth("", config.SMTP.User, config.SMTP.Password, smtpServer)

	mailBody := fmt.Sprintf(
		"From: %s\nTo: %s\nSubject: %s\nMIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n%s",
		config.SMTP.From, to, subject, body)

	// Send the email
	err := smtp.SendMail(fmt.Sprintf("%s:%d", smtpServer, smtpPort), auth, config.SMTP.From, []string{to}, []byte(mailBody))
	if err != nil {
		return err
	}

	return nil
}

func randBytes(l int) []byte {
	b := make([]byte, l)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic(err.Error())
	}
	return b
}
