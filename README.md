# WhatsNew

A simple tool to collect and process quite a few web news from multiple sources.

![](https://raw.githubusercontent.com/SpecializedGeneralist/whatsnew/main/assets/architecture.png)

## Requirements

* [Go 1.16](https://golang.org/dl/)
* [RabbitMQ](https://www.rabbitmq.com/)
* [PostgreSQL](https://www.postgresql.org/)

## How it works

`whatsnew` allows you to define custom pipelines for fetching, processing and search 
resources from the web - most notably, news articles.

It comes in the form of both a library and a standalone program:

* as a library, it provides the basic tools to implement your own custom
  functionalities: you can expand the existing codebase, or simply import it as
  a Go package and use what you need;
* this project also provides a ready-to-use implementation of a typical pipeline
  to fetch and analyze some resources from the web: you can run different tasks
  from a single tiny executable program.

`whatsnew` is designed with modularity, flexibility and scalability in mind.

The building blocks of `whatsnew` are **tasks**. A task is just a process which
performs some operations and usually can read/write data from/to a centralized
PostgreSQL database. The most common tasks generally fall in three categories:

* fetching, downloading or crawling documents (or articles) from the web and
  storing them on the database;
* performing some sort of analysis on one or more documents (from the database),
  eventually storing the results.
* searching, extracting or filtering the stored results using _semantic_ technology 
  by means of query written in natural language.

Certain tasks might not require any specific input or event to be triggered;
this is the case, among others, for tasks running some operations
periodically. For example, a task might fetch the list of articles from an
RSS Feed every 10 minutes and store the results to the database.

Other tasks might instead require prior operations to be complete before
proceeding, or might need to be triggered by certain events. For example,
a task might need to know when an article is crawled from the web, so that it
can be analyzed with NLP (Natural Language Processing) tools.

Tasks can _indirectly_ communicate each other via RabbitMQ messages. Once a
task is done with its own operations, it usually writes some data to the
database _and_ also publishes one or more messages to a RabbitMQ exchange,
commonly using custom routing-keys and referring to the IDs of the newly
processed records. One or more other tasks subscribed to the exchange will
then intercept the message and carry on their own computation.
In this way, each task might only need to know the RabbitMQ routing-keys for a
subscription and/or for publishing.

So, by simply configuring a bunch of routing keys, you can actually set up
an entire processing **pipeline**.

Following is an example of configuration in which you can see the use of routing keys (publish/subscribe) for each task:

```console
log_level: 'info'
db:
  dsn: 'host=localhost port=5432 user=postgres password=postgres dbname=whatsnew sslmode=disable statement_cache_mode=describe'
rabbitmq:
  uri: 'amqp://guest:guest@localhost:5672/'
  exchange_name: 'whatsnew'
feeds_fetching:
  num_workers: 50
  max_allowed_failures: 15
  sleeping_time: 10m
  omit_feed_items_published_before_enabled: true
  omit_feed_items_published_before: '2020-12-01T00:00:00Z'
  new_web_resource_routing_key: 'new_web_resource'
  new_feed_item_routing_key: 'new_feed_item'
gdelt_fetching:
  sleeping_time: 5m
  new_web_resource_routing_key: 'new_web_resource'
  new_gdelt_event_routing_key: 'new_gdelt_event'
  top_level_cameo_event_code_whitelist: [ ]
tweets_fetching:
  num_workers: 50
  sleeping_time: 5m
  omit_tweets_published_before_enabled: true
  omit_tweets_published_before: '2020-12-01T00:00:00Z'
  max_tweets_number: 3200
  new_web_resource_routing_key: ''
  new_tweet_routing_key: 'new_tweet'
  new_web_article_routing_key: 'new_web_article'
web_scraping:
  num_workers: 40
  sub_queue_name: 'whatsnew.web_scraping'
  sub_new_web_resource_routing_key: 'new_web_resource'
  pub_new_web_article_routing_key: 'new_web_article'
duplicate_detector:
  timeframe_hours: 72
  similarity_threshold: 0.7
  sub_queue_name: "whatsnew.duplicate_detector"
  sub_routing_key: "new_vectorized_web_article"
  pub_new_event_routing_key: "new_event"
  pub_new_related_routing_key: "new_related"
vectorizer:
  num_workers: 4
  sub_queue_name: "whatsnew.vectorizer"
  sub_new_web_article_routing_key: "web_article_classified"
  pub_new_vectorized_web_article_routing_key: "new_vectorized_web_article"
  labse_grpc_address: "localhost:1976"
  labse_tls_disable: false
zero_shot_classification:
  num_workers: 4
  sub_queue_name: 'whatsnew.zero_shot_classification'
  sub_routing_key: 'new_web_article'
  pub_routing_key: 'web_article_classified'
  payload_key: 'zero_shot_classification'
  zero_shot_grpc_address: "localhost:4001"
  grpc_tls_disable: true
  hypothesis_template: 'This text is about {}.'
  possible_labels: ['sport', 'economy', 'science']
  multi_class: true
supported_languages: [ 'en', 'es' ]
```

Each task being part of a pipeline can still be a completely independent
program. It's easy to make it use system resources or be scaled according to
specific needs, especially in cloud-computing environments. Each task/program
can also handle various failures without directly affecting the other tasks.
Moreover, some RabbitMQ messages can trigger multiple tasks, which can
therefore run in parallel.

## Built-in models and tasks

`whatsnew` provides a set of built-in tasks, available as packages under
`pkg/tasks` and also associated to specific commands on the main executable.
These tasks communicate with PostgreSQL database via a set of
[Gorm](https://gorm.io/) models, defined under `pkg/models`.

Here is a quick description of the tasks, along with references to the CLI
commands and source code packages:

* `fetch-feeds` | `pkg/tasks/feedsfetching`

  Periodically loop through the list of RSS/Atom Feed URL and add new
  Feed Items (articles) to the database.

* `fetch-gdelt` | `pkg/tasks/gdeltfetching`

  Periodically fetch the latest news (articles URLs) from the
  [GDELT Project](https://www.gdeltproject.org/).

* `fetch-tweets` | `pkg/tasks/tweetsfetching`

  Crawl tweet contents for the configured sources (users and search terms).

* `scrape-web` | `pkg/tasks/webscraping`

  Crawl web pages content, for example for feed items and GDELT articles.

* `translate` | `pkg/tasks/translator`

  Translate the web articles and tweets in a desired language via
  [Translator](https://github.com/SpecializedGeneralist/spago), a simple 
  self-hostable Machine Translation service, powered by spaGO.
  
* `zero-shot-classification` | `pkg/tasks/zeroshotclassification`

  Classify scraped news articles and tweets with a [spaGO](https://github.com/nlpodyssey/spago)
  zero-shot classification service (it must run separately and be configured
  appropriately).
  The categories are arbitrary, and the software does not require training. 
  For instance, categories like `music`, `sport`, `science`, `technology`, `economy and finance` should work out of the box.
  The zero-shot classification can also be used for sentiment analysis, i.e., `positive`, `negative`, or `neutral` categories.

* `vectorize` | `pkg/tasks/vectorizer`

  Create a vector representation of web articles and tweets, using
  [spaGO](https://github.com/nlpodyssey/spago) LaBSE encoding (external
  service).

* `duplicate-detector` | `pkg/tasks/duplicatedetector`

  Perform near-duplicate news/tweets detection via cosine similarity of
  vectors.

Additionally, here are some utility CLI commands:

* `create-db` - create the database with given name, without creating the tables
* `migrate-db` - initializes an existing database with the proper tables, or
  perform automatic migrations if you had an older version of the software
* `add-feeds` - insert RSS/Atom Feed URLs from a list
* `add-twitter-sources` - insert new Twitter sources from a TSV file
  (columns: `[type, value]`, corresponding to the fileds of `TwitterSource`
  model)

## Go library

`whatsnew` can be used as Go library from your own project:

```shell
go get -u github.com/SpecializedGeneralist/whatsnew
```

## CLI program

You can clone the project and build the CLI program:

```shell
git clone https://github.com/SpecializedGeneralist/whatsnew.git
cd whatsnew
go mod download
go build -o whatsnew cmd/whatsnew.go
./whatsnew -h
```

### Docker

The CLI program can run from a Docker container. To build the image:

```console
docker build -t whatsnew:latest .
```

Pre-built images are available on Docker Hub, at [specializedgeneralist/whatsnew](https://hub.docker.com/r/specializedgeneralist/whatsnew).
For example:

```shell
docker pull specializedgeneralist/whatsnew:0.3.3
```

### Usage example

Make sure RabbitMQ and PostgreSQL are up and running on your machine.

You may find the following commands useful to launch them as docker containers:
```console
docker run -d -p 5672:5672 -p 15672:15672 --name rabbitmq rabbitmq:3.8.6-management-alpine
docker run -d -p 5432:5432 --name postgres -e POSTGRES_PASSWORD=postgres postgres:12.3-alpine
```

Create a new PostgreSQL DB with name `whatsnew` (actually the name is your choice).

Take your time to create the configuration file, starting from
[sample-configuration.yml](https://github.com/SpecializedGeneralist/whatsnew/blob/main/sample-configuration.yml)
in the project folder. If we did well, you should understand all the settings
as you go.

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
