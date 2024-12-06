# Gol Observer

Gol Observer is a little tool I use to monitor my log files remotely using a stupidly reduced web user interface. I basically works by `tail -f`ing the given log files and sending them to the client via websockets.



## Installation instructions

- Clone this repo



### Frontend

**Important note**

Before deploying change the `BASE_URL`in the index.js to your backend url.



You can deploy the frontend to your to your favorite online hosted webspace or e.g. a apache2, nginx site. 

Just make sure to copy these files:

- index.html

- index-dist.css

- index.js



### Backend

**Important note**

Before building change the Allowed origins in the `main`func:

```go
c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://127.0.0.1:5500",
			"https://logs.cr.codes",
		},
})
```

And if you want the port:

```go
log.Println("Listening on 127.0.0.1:8888")
err := http.ListenAndServe(":8888", handler)
```

- `cd server`

- `GOBIN=/usr/local/bin sudo go install gol-observer` compiles and installs the server in the path /usr/local/bin

- On Linux copy service file to `/etc/systemd/system/`

- Start the service `systemctl start gol-observer`
