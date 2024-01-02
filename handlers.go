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
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"slices"

	"github.com/gofiber/fiber/v2"
)

var maxId = big.NewInt(2147483647)

const AESSize = 128 >> 3

const mailTemplate = `<p>Hi, and thanks for using SFUP!</p>
<p>&nbsp;&nbsp;Use this command to upload your file:</p>
<pre>
curl -s %s/bash/%d|sh -s -- &lt;filename&gt;
</pre>
<p>or, simpler (and windows-compatible):</p>
<pre>
curl -qF "file=@&lt;filename&gt;" %s/ul/%d
</pre>
<p>Have fun!</p>
<p>-- sfup</p>`

func reserve(db *sql.DB) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		mail := c.Params("mail")

		if !slices.Contains(config.AllowedEmails, mail) {
			return c.Status(fiber.StatusUnauthorized).SendString("Sorry, your email is not authorized")
		}

		bid, err := rand.Int(rand.Reader, maxId)
		if err != nil {
			panic(err)
		}

		id := int(bid.Int64())

		if _, err = db.Exec("INSERT INTO SFUP (id, name, last_upd) VALUES (?,  NULL, CURRENT_TIMESTAMP)", id); err != nil {
			panic(err)
		}

		sendEmail(mail, "Your SFUP reservation", fmt.Sprintf(mailTemplate, c.BaseURL(), id, c.BaseURL(), id))

		return c.SendString("\nOk, all set up. We sent you an email with instructions!\n")
	}
}

const template = `#!/bin/bash
curl -qF "file=@$1" %s/ul/%s`

func bash(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.SendString(fmt.Sprintf(template, c.BaseURL(), id))
}

const dlUrlTemplate = "%s/dl/%s"
const dlMsgTemplate = `
Upload done! To download the file, either use a browser:

  %s

or, from the commandline:

  curl -OJ %s

The file will be deleted from SFUP after the download.

Have fun!
`

func upload(db *sql.DB) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		file, err := c.FormFile("file")
		if err != nil {
			panic(err)
		}
		f, err := file.Open()
		if err != nil {
			panic(err)
		}

		defer f.Close()

		key := randBytes(AESSize)

		aesName, err := aes.NewCipher(key)
		if err != nil {
			panic(err.Error())
		}
		ivName := randBytes(aesName.BlockSize())
		fnBytes := []byte(file.Filename)
		encName := make([]byte, len(fnBytes))
		ctrName := cipher.NewCTR(aesName, ivName)
		ctrName.XORKeyStream(encName, fnBytes)

		aesFile, err := aes.NewCipher(key)
		if err != nil {
			panic(err.Error())
		}
		ivFile := randBytes(aesName.BlockSize())

		n, err := db.Exec("UPDATE SFUP SET iv_file = ?, iv_name = ?, name = ?, last_upd = CURRENT_TIMESTAMP WHERE id = ? AND NAME IS NULL", ivFile, ivName, encName, id)
		if err != nil {
			panic(err)
		}
		if nn, _ := n.RowsAffected(); nn == 0 {
			return c.Status(fiber.StatusNotFound).SendString("Invalid ID or already used")
		}

		outFile, err := os.Create(dataDir(id))
		if err != nil {
			panic(err)
		}
		defer outFile.Close()

		sw := &cipher.StreamWriter{
			S: cipher.NewCTR(aesFile, ivFile),
			W: outFile,
		}

		if _, err := io.Copy(sw, f); err != nil {
			panic(err)
		}

		url := fmt.Sprintf(dlUrlTemplate, c.BaseURL(), id)
		return c.SendString(fmt.Sprintf(dlMsgTemplate, url, url))
	}
}

func download(db *sql.DB) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		file := dataDir(id)

		row := db.QueryRow("SELECT name FROM SFUP WHERE id = ?", id)
		var name string
		err := row.Scan(&name)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(fiber.StatusNotFound).SendString("Invalid ID or already used")
			}
			log.Fatal(err)
		}

		defer func() {
			_, _ = db.Exec("DELETE FROM SFUP WHERE id = ?", id)
			os.Remove(file)
		}()

		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", name))
		return c.SendFile(file, true)
	}
}
