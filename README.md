# hearsay
An authorship attribution and NLP bot for IRC. Built in Go and Python 3. Now with GPU support when using BERT word embeddings.

## Table of Contents
- [Features](#features)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Examples](#examples)
- [Limitations](#limitations)
- [License](#license)

## Features
- Store and track messages from IRC channels
- Attribute a given message to the most likely user
- Flesch-Kincaid readability scores
- Sentiment analysis of nicks and messages
- Stylistic neighbours based on confusion matrices
- Extensive opt and privacy features (opt-out by default)
- Go-based IRC handling. Python-based NLP processing

## Installation
hearsay uses the [GoIRC](https://github.com/fluffle/goirc) client for IRC connections and handling, [go-sqlite3](https://github.com/mattn/go-sqlite3) for persistent storage, and [scikit-learn](https://scikit-learn.org/) for machine learning tasks. A local Python HTTP [API](https://fastapi.tiangolo.com/) relays NLP information back to the Go client.

It is highly recommended to use the provided Docker configuration to run hearsay. To build and run the bot, run `sudo docker compose up --build --detach` in the base directory.

> [!IMPORTANT]
> The bot is automatically configured to connect to `irc.zoite.net`. Change this before running the Docker container. See [Configuration](#Configuration) for more.

This might take a while.

## Configuration
The configuration file is located in `bot/config.yaml`. Before anything else, the default values must be changed to match your and your server's expectations.

```yaml
bot:
  prefix: "+"
  mode: "+B"
  server: "irc.zoite.net:6697"
  channel: "#antisocial"

storage:
  message_pool_size: 20
  message_quota: 1000
  people_quota: 5

scheduler:
  deletion_days: 1

model:
  bert: true
  gpu: false
```

- `prefix`: The bot prefix. If a message starts with this symbol or string, it will be activated. In case of prefix conflict with existing bots, change this.
- `mode`: This are the positive or (exclusive) negative modes to be set on the bot. `+B` is a common mode for server bots.
- `server`: Server and port to connect to on start-up.
- `channel`: Channel to connect to on start-up.
- `message_pool_size`: By default, hearsay does not submit an incoming message to the database when received. Instead, it waits for a message pool to fill up before creating a transaction where all (in this case 20) messages are submitted. This prevents frequent I/O. Depending on server size, you might want to adjust this value, but 20 is a good middle ground.
- `message_quota`: This is an important setting. Before users can access NLP commands, they must fulfil a message quota. If the message quota is too low, the bot will make inaccurate assessments. One thousand is a good albeit high quota. Five-hundred messages will also work with the cost of lessened accuracy.
- `people_quota`: Before authorship attribution commands can be used, five people must fulfil the `message_quota`. With a lower `people_quota`, the author population becomes less diverse. Five is a good start for small to medium big servers.
- `deletion_days`: When a user issues the `forget` command, all their data will be purged. To prevent accidental deletions, their request is put on a schedule. After the set amount of days, their data will be purged. Note that `deletion_days` cannot be lower than one.
- `bert`: Enables text embeddings with Google's BERT language model.
- `gpu`: Enable GPU with BERT resulting in massive time reduction.
> [!NOTE]
> Using BERT is slow and the accuracy gain is minimal. If you wish to disable it, set this setting to false. However, BERT will still install during installation. This produces some overhead. Remove the line `sentence-transformers` from `api/requirements.txt` to disable it completely. Please note that BERT is not enabled by default if it is set to true. A separate `--bert` flag has to be passed to `+retrain` to use it.

## Usage

To get help on a command, use the `help` command. Available commands are attribute, opt, forget, unforget, help, readability, retrain, about, sentiment, me, and profile.

- `attribute`: Attribute a message to a chatter who is opted in and fulfils the message quota. To view the model's scope of view, use the --list flag. Usage: `+attribute (--list|<message>)`
- `opt`:  Opt in or out from data collection and model training. If no arguments are submitted, your current opt status will be returned. Usage: `+opt [in|out] (default: out)`
- `forget`: Permanently purge all your data. Usage: `+forget`
- `unforget`: Cancel a scheduled data deletion. Usage: `+unforget`
- `help`: Get information on a command. Usage: `+help [command]`
- `readability`: Calculate the Flesch-Kincaid readability score of your messages (10,000 limit). Usage: `+readability`
- `retrain`: Refit the classification model. This can be done every 2 hours. Add the --cm flag for evaluation statistics (heavy). To ignore inactive nicks, provide the --past flag with the number of days of inactivity before being cut off. To include BERT embeddings, append the --bert flag. NOTE: Using BERT is very slow with minimal accuracy gain. This is compounded when used in conjunction with --cm. Usage: `+retrain [--cm, --bert, --past <days>]`
- `about`: Information about hearsay. Usage: `+about`
- `sentiment`: Extract the sentiment (positive, neutral, or negative) from a message. Usage: `+sentiment <message>`
- `me`: Statistics about yourself. Usage: `+me`
- `profile`: Build author profiles that provide higher attribution accuracy. Usage: `+profile (attribute|create|destroy) <name> | append <name> <message> | list`

## Examples
### Retrain
```
katt> +retrain --cm --past 20 --bert
<hearsay> katt: The SVM model has been retrained. It took 40.17 seconds to fit. Confusion matrix: http://tmpfiles.org/11225615/cm.png | 5-fold CV: Accuracy 0.6321, F1 score 0.6304
```
![Confusion matrix](/misc/cm.png)

### Attribution
```
<katt> +attribute i hope you had a good weekend
<hearsay> katt: Predicted author: morph_. Confidence scores: morph_ (0.56), ack_ (0.09), tonitrus_ (-1.02)
```

### Readability
```
<katt> +readability
<hearsay> katt: You have a Flesch-Kincaid score of 82.06 (6th grade level. Easy to read. Conversational English for consumers.)
```

### Sentiment
```
<katt> +sentiment I hate my job
<hearsay> katt: Largely negative with a compound score of -0.57. (pos: 0.00, neu: 0.45, neg: 0.55)
```

### Me
```
katt: Message count: 9001/400 | Readability: 82.01 | Sentiment: 0.10 (Positive) | Neighbour: gothdaria_
```

## Limitations
The nature of IRC messages (short and often fruitless in information) pose a challenge to authorship attribution. hearsay uses an instance-based, per-message model, including short messages. This introduces a lot of noise in the training data. To circumvent this, profiles were created. With profiles, it is possible to append more than one message to be attributed. You can expect a 25%+ accuracy boost depending on the size of a profile and author population.

## License
This project is unlicensed. The bot may be used in any and all ways, with or without credit.
