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
	"flag"
	"fmt"
	"log"
	"os"

	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gopkg.in/yaml.v2"

	_ "github.com/mattn/go-sqlite3"
)

type Smtp struct {
	Server   string `yaml:"server"`
	Port     int    `yaml:"port"`
	User     string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

type Conf struct {
	Quota         int      `yaml:"quota"`
	MaxFileSize   int      `yaml:"max_file_size"`
	SMTP          Smtp     `yaml:"smtp_server"`
	AllowedEmails []string `yaml:"allowed_emails"`
}

type Args struct {
	ConfigFile string
	Port       int
	DataDir    string
}

var config Conf
var args Args

func dataDir(fname string) string {
	return fmt.Sprintf("%s/%s", args.DataDir, fname)
}

func main() {
	_configFile := flag.String("config-file", "config.yaml", "Path to configuration file")
	_port := flag.Int("port", 8080, "Port")
	_dataDir := flag.String("data-dir", "files", "Path to data dir")

	flag.Parse()

	args = Args{
		ConfigFile: *_configFile,
		Port:       *_port,
		DataDir:    *_dataDir,
	}

	data, err := os.ReadFile(args.ConfigFile)
	if err != nil {
		println(err.Error())
		return
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		println(err.Error())
		return
	}

	db, err := sql.Open("sqlite3", dataDir("sfup.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS SFUP (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		log.Fatal(err)
	}

	// Fiber instance
	app := fiber.New(fiber.Config{
		BodyLimit: 1024 * 1024 * 1024,
		// AppName:   "SFUP",
		BodyLimit:             config.MaxFileSize,
		DisableStartupMessage: true,
	})

	fmt.Fprint(os.Stdout, "SFUP v0.0.1\n")

	app.Use(recover.New())

	app.Get("/reserve/:mail", reserve(db))
	app.Get("/bash/:id", bash)
	app.Get("/dl/:id", download(db))
	app.Post("/ul/:id", upload(db))

	log.Fatal(app.Listen(fmt.Sprintf(":%d", args.Port)))
}
