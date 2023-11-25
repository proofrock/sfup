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
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"slices"

	"github.com/gofiber/fiber/v2"
)

const mailTemplate = `<p>Hi!</p><p>&nbsp;&nbsp;Use this command to upload your file:</p><pre>
curl -s %s/bash/%d|sh -s -- &lt;filename&gt;
</pre><p>Have fun!</p><p>-- sfup</p>`

func reserve(db *sql.DB) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		mail := c.Params("mail")

		if !slices.Contains(config.AllowedEmails, mail) {
			return c.Status(fiber.StatusUnauthorized).SendString("Sorry, your email is not authorized")
		}

		id := rand.Int31()

		_, err := db.Exec("INSERT INTO SFUP (id, name) VALUES (?,  NULL)", id)
		if err != nil {
			panic(err)
		}

		sendEmail(mail, "Your SFUP reservation", fmt.Sprintf(mailTemplate, c.BaseURL(), id))

		return c.SendString("\nWe sent you an email with instructions!\n")
	}
}

const template = `#!/bin/bash
curl -q -F "file=@$1" %s/ul/%s`

func bash(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.SendString(fmt.Sprintf(template, c.BaseURL(), id))
}

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

		f.Close()

		n, err := db.Exec("UPDATE SFUP SET name = ? WHERE id = ?", file.Filename, id)
		if err != nil {
			panic(err)
		}
		if nn, _ := n.RowsAffected(); nn == 0 {
			return c.Status(fiber.StatusNotFound).SendString("Invalid ID or already used")
		}

		c.SaveFile(file, dataDir(id))
		return c.SendString(fmt.Sprintf("\nNow, to download, use: curl -OJ %s/dl/%s\n", c.BaseURL(), id))
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
