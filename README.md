# Gol Observer

Gol Observer is a little tool I use to monitor my log files remotely using a stupidly reduced web user interface.<br>
I basically works by `tail -F`ing the given log files and sending them to the client via websockets.
<br><br>**Why is it called *Gol*?**<br>
Because its partly written in Go and spells *log* backwards :D
## Installation instructions

- Clone this repo

### Frontend

**Important note**

Before deploying:
- Change the `BASE_URL`in the index.js to your backend url.
- Change the `API_KEY` in the index.js to your desired api key.

You can deploy the frontend to your to your favorite online hosted webspace or e.g. a apache2, nginx site (Hardly recommended to use HTTPS and a htaccess file because I was so lazy that I hardcoded the api key directly into index.js). 

Just make sure to copy these files:

- fonts (directory)
- index.html
- index-dist.css
- index.js

### Backend

- `cd server`
- `go build` 
(On my Raspberry Pi 4 I had to do `env GOOS=linux GOARCH=arm64 go build gol-observer.go`)
- `sudo cp -p gol-observer /usr/local/bin/`

**Important note**
Before running the service you might want to change the `WorkingDirectory` in the service file where your config.json file is located.

- On Linux copy service file to `/etc/systemd/system/`
- Reload daemons `sudo systemctl daemon-reload`
- Start the service `sudo systemctl start gol-observer`

### Configuration
The `config.json` is structured as follows:
```json
{
  "server" : {
    "host": "127.0.0.1",
    "port": "8888",
    "external_api_url": "https://your_url_here.com",
    "api_key": "your-api-key-here"
  },
  "log_files_glob": [
      "/path/to/sample/*.log"
  ]
}
```
