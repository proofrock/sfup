# sfup - transfer service for commandline usage
## v0.0.2

Imagine to have two headless servers, and needing to transfer a file between them. But the servers cannot connect to each other. And of course, being on the commandline, you can't use WeTransfer, Dropbox, ecc...

Enter `sfup`. It's a simple storage server/transfer service that can be used with cUrl. Install it:

- You set it up with a public endpoint (I use docker+cloudflare tunnels)
- You supply a whitelist of allowed e-mails and a SMTP server to send emails (e.g. GMail)

Then, when you need to transfer a file:

- From a browser, visit `http://sfup.example.com/reserve/<your_email_address>`
- If the address is authorized, you'll receive an email with a one-time shell command to upload a file
- Execute the command on server A, upload the file, and the operation will output another one-time command to download it
- On server B, execute the command and download the file
- The file is deleted from `sfup`

# Install

Simple install, with docker:

`docker run --name sfup -p 8080:8080 -v sfup:/data -v config.yaml:/config.yaml germanorizzo/sfup:latest`

For the `config.yaml` file, see and adapt `config.yaml.template` in this repository.

# Features

- One-time everything: files are stored up to the first (and only) download
- Expiration of reservations and files
- Configuration to set the maximum file size
- Single binary written in Go
- Very compact codebase

# Issues

- Upload command doesn't show progress
- Download command doesn't display errors

# Next Steps

It's pretty basic for now. Future plans:

- Better Windows compatibility
- Quotas
- E2e encryption
- Checksumming
- Web UI to download files
