# sfup - transfer service for commandline usage
## v0.1.0

Imagine to have two headless servers, and needing to transfer a file between them. But the servers cannot connect to each other. And of course, being on the commandline, you can't use WeTransfer, Dropbox, ecc...

Enter `sfup`. It's a simple storage server/transfer service that can be used with cUrl. Install it:

- You set it up with a public endpoint (I use docker+cloudflare tunnels);
- You supply a whitelist of allowed e-mails and a SMTP server to send emails (e.g. GMail).

Then, when you need to transfer a file:

- From a browser, visit `http://sfup.example.com/reserve/<your_email_address>`;
- If the address is authorized, you'll receive an email with a one-time shell command to upload a file (using cURL);
- Execute the command on server A, upload the file, and the operation will output:
  - A link for downloading with a browser;
  - A command to download from CLI with cURL;
- On server B, execute the command and download the file;
- The file is then deleted from `sfup`.

## Note

When the file is downloaded, an integrity check is performed. If it fails, the file is downloaded but the call will return `599` as a status code. Depending on your setup, this may be difficult to catch; for example, cURL will print:
```
curl: (22) The requested URL returned error: 599
```

# Install

Simple install, with docker:

`docker run --name sfup -p 8080:8080 -v sfup:/data -v config.yaml:/config.yaml germanorizzo/sfup:latest`

For the `config.yaml` file, see and adapt `config.yaml.template` in this repository.

# Features

- One-time *everything*: files are stored up to the first (and only) download
- E2e encryption, the key is in the download link
- Integrity check
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
