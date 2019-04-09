Mailgun Relayery
================

![build status](https://travis-ci.com/Parquery/mailgun-relayery.svg?branch=master)
[![Coverage Status](https://coveralls.io/repos/github/Parquery/mailgun-relayery/badge.svg?branch=master)](https://coveralls.io/github/Parquery/mailgun-relayery?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/Parquery/mailgun-relayery)](https://goreportcard.com/report/github.com/Parquery/mailgun-relayery)
[![godoc](https://img.shields.io/badge/godoc-reference-5272B4.svg)](https://godoc.org/github.com/Parquery/mailgun-relayery)

Mailgun Relayery is a tool for relaying messages with different API keys through the single-key MailGun API.

**The problem**: MailGun gives out a single API key to each account. For entities dealing with multiple clients or 
applications, giving out their API key to each of them constitutes a security violation.

**The solution**: _Mailgun Relayery_ acts as a middle layer between message senders and the MailGun API.
Senders are issued descriptor strings (analogous to usernames) and authentication tokens (analogous to 
passwords), which are used to authenticate their requests to the middle layer and to relay their messages to MailGun.

The tool stores authentication data (tokens) and channeling data (mailing fields like recipients, cc, bcc, _etc_.) 
in a local database, guaranteeing persistency. Two servers have access to the database:

* the Control server (with read-write access to the database) manages authentication data and mailing channels; this 
    server should only be accessed via secure connection and managed by a trusted party.
* the Relay server (with read-only access to the database) receives and authenticates HTTP requests to relay 
    messages to the MailGun API; this server is open to the whole Internet.


All communication with the servers takes place via HTTP following the
[relay server API](https://github.com/Parquery/mailgun-relayery/swagger/relay/swagger.yaml) and the
[control server API](https://github.com/Parquery/mailgun-relayery/swagger/control/swagger.yaml).

An example [Python script](https://github.com/Parquery/mailgun-relayery/example.py) is available for a quick start. 


Usage
=====

Compilation
-----------
If you're running a linux x64 architecture, you can directly use the binaries available in 
the [Releases](https://github.com/Parquery/mailgun-relayery/releases) page.

*  Run the [release script](https://github.com/Parquery/mailgun-relayery/release.py) to compile and release the two 
server binaries to a target directory:
  
    ```bash
    ./release.py --release_dir your/release/directory
    ```
*  Run the database initialization binary to create an empty channel database in a target directory:
  
    ```bash
    ./your/release/directory/mailgun-relayery-init -database_dir /your/database/directory
    ```

Running the servers
-------------------

*  Run the Control Server:
  
    ```bash
    your/release/directory/bin/mailgun-relay-controlery -database_dir your/database/directory
    ```

*  Run the Relay Server:
  
    ```bash
    your/release/directory/bin/mailgun-relayery \
       -database_dir your/database/directory \
       -api_key_path path/to/mailgun/api/key.txt
    ```
    
Sending requests
----------------
* Use the [Control Server API](https://github.com/Parquery/mailgun-relayery/swagger/control/swagger.yaml) 
  to create a mailing channel and authorization token:
  
    ```bash
    curl -i \
        -X PUT \
        -H "Accept: application/json" \
        -H "Content-Type: application/json" \
        --data '{
            "domain": "marketing.domainname.com",
            "descriptor": "some-channel",
            "token": "oqiwdJKNsdK",
            "sender": {
              "email": "your-name@company.com",
              "name":"Some Sender"
            },
            "recipients":[
              {
                "email": "your-client@client.com",
                "name": "Recipient"
              }
            ],
            "cc": [],
            "bcc": [],
            "min_period":0.1,
            "max_size":10000000
          }' \
      "localhost:8300/api/channel"
    ```
     
* Use the [Relay Server API](https://github.com/Parquery/mailgun-relayery/swagger/relay/swagger.yaml) 
  to relay a message:
  
    ```bash
    curl -i \
        -X POST \
        -H "Accept: application/json" \
        -H "Content-Type: application/json" \
        -H "X-Descriptor: some-channel" \
        -H "X-Token: oqiwdJKNsdK" \
        --data '{
          "message": "hello there",
          "subject": "a message from your friend"
        }' \
        "localhost:8200/api/message
    ```
     
Development
===========
Please refer to [CONTRIBUTING.md](https://github.com/Parquery/mailgun-relayery/blob/master/CONTRIBUTING.md).

Versioning
==========
We follow [Semantic Versioning](http://semver.org/spec/v1.0.0.html).
The version X.Y.Z indicates:

* X is the major version (backward-incompatible),
* Y is the minor version (backward-compatible), and
* Z is the patch version (backward-compatible bug fix).
