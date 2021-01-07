# WhatsNew

A simple tool to collect and process quite a few web news from multiple sources.

## Requirements

* [Go 1.15](https://golang.org/dl/)
* [Go modules](https://blog.golang.org/using-go-modules)
* [RabbitMQ](https://www.rabbitmq.com/)
* [PostgreSQL](https://www.postgresql.org/)

## Usage

WhatsNew is a Go module, so it can be used as a library within your own software.

```console
go get -u github.com/SpecializedGeneralist/whatsnew
```

It also works out-of-the-box using its CLI mode:

```console
NAME:
   whatsnew - A simple tool to collect and process quite a few web news from multiple sources

USAGE:
   whatsnew [global options] command [command options] [arguments...]

COMMANDS:
   create-db    Perform automatic database creation
   migrate-db   Perform automatic database migration
   add-feeds    Add new feeds from a list
   fetch-feeds  Fetch all feeds and get new feed items
   fetch-gdelt  Fetch latest news from GDELT
   scrape-web   Scrape news articles from `Web Resource` URLs
   help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config FILE, -c FILE  load configuration from YAML FILE
   --help, -h              show help (default: false)
```

### Build

To build WhatsNew (CLI mode), move into the project directory, and run the following command:

```console
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-extldflags=-static" -a -o whatsnew cmd/whatsnew.go
``` 

If the command is successful you should find the `whatsnew` executable in the same folder.

> For now, it has not been tested on other operating systems and other architectures. It would be great if you could try it and let us know where it works and where it doesn't!

Alternatively, the [Docker](https://www.docker.com/) image can be built like this:

```console
docker build -t whatsnew:main . -f Dockerfile
```

### Run

Make sure the services of RabbitMQ and PostgreSQL are up and running on your machine.

You may find the following commands useful to launch them as docker containers.

```console
docker run -d -p 5672:5672 -p 15672:15672 --name rabbitmq rabbitmq:3.8.6-management-alpine
docker run -d -p 5432:5432 --name postgres -e POSTGRES_PASSWORD=postgres postgres:12.3-alpine
```

Create a new PostgreSQL DB with name `whatsnew` (actually the name is your choice).

Take your time to create the configuration file, starting from [sample-configuration.yml](https://github.com/SpecializedGeneralist/whatsnew/blob/main/sample-configuration.yml) in the project folder. If we did well you should understand all the settings as you go.

Set up the database running:

```console
./whatsnew --config your-config.yml migrate-db
```

Make a list of RSS feeds relevant to you, for example:

```console
cat <<EOT >> feeds.txt
https://rss.nytimes.com/services/xml/rss/nyt/World.xml
http://feeds.bbci.co.uk/news/rss.xml
EOT
```

Now load them on WhatsNew:

```console
./whatsnew --config your-config.yaml add-feeds -f feeds.txt
```

Once it's done, run multiple WhatsNew instances, each with a specific command:

1. ```./whatsnew --config your-config.yml scrape-web```
2. ```./whatsnew --config your-config.yml fetch-feeds```
3. ```./whatsnew --config your-config.yml fetch-gdelt```


Enjoy ...or tell us what went wrong!
