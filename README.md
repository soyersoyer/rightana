# K20a
[![Build Status](https://travis-ci.org/soyersoyer/k20a.svg?branch=master)](https://travis-ci.org/soyersoyer/k20a) 
[![Go Report Card](https://goreportcard.com/badge/github.com/soyersoyer/k20a?)](https://goreportcard.com/report/github.com/soyersoyer/k20a)

Carefree web analytics on your server

## What is K20a?

Its a self hosted web analytics software.

## How does it work?

From the user's perspective it is very similar to GA
- register
- add your site
- get a tracking code
- include it to your webpage
- View the reports

## Who uses K20a?

We use it to track our and our client's websites.

## Goals

- Easy to install
- Easy to use
- Easy to upgrade (guaranteed after version 1.0)
- One binary distribution
- Space efficient, fast, embedded database
- Visitor friendly
- You don't have to sell your visitor's data to a company
- Tracks sessions, not users


### Screenshots

### Simple overview
![chart](https://user-images.githubusercontent.com/5169997/34117162-1f82043a-e41b-11e7-9ff5-72a0d82f1bfb.png)

### Multiple resolution and interval
![resolution](https://user-images.githubusercontent.com/5169997/34116446-f7ae1018-e418-11e7-9b12-159160aef5f6.png)

### Basic informations
![basic-info](https://user-images.githubusercontent.com/5169997/34116575-5484cf84-e419-11e7-8423-d9c9c769def5.png)

### Multiple summaries
![pages](https://user-images.githubusercontent.com/5169997/34116643-81d16ae2-e419-11e7-9547-1bf1d1c25879.png)
![sums](https://user-images.githubusercontent.com/5169997/34116646-83392fc8-e419-11e7-84b0-2331a7d84eb9.png)

### GeoIP support
![geoip](https://user-images.githubusercontent.com/5169997/34117762-f5268006-e41c-11e7-8ea3-34722e057fea.png)

### You can add filters
![filter](https://user-images.githubusercontent.com/5169997/34116771-d6d3328c-e419-11e7-8631-98910fda9dcb.png)

### You can use the bars for selecting the intervals too
![bar-navigation](https://user-images.githubusercontent.com/5169997/34116997-8ee17c44-e41a-11e7-874b-b83719136cad.png)

### Watch sessions
![sessions](https://user-images.githubusercontent.com/5169997/34117093-e0f252d8-e41a-11e7-8811-5c90d73560b5.png)

### Multiple user and site support
![registration](https://user-images.githubusercontent.com/5169997/34117560-533d43ce-e41c-11e7-8254-bce5390ed326.png)
![multiple-site](https://user-images.githubusercontent.com/5169997/34117484-2461c25a-e41c-11e7-86f6-3280a5d46291.png)

### You can easily get a tracking code for your site in the settings
![tracking-code](https://user-images.githubusercontent.com/5169997/34116917-498f58d2-e41a-11e7-8d3c-80190269a1cc.png)

### Add teammates
![teammates](https://user-images.githubusercontent.com/5169997/34117250-6577d690-e41b-11e7-9931-2c3ccca01b91.png)

### Deleting old data is easy
![storage](https://user-images.githubusercontent.com/5169997/34117249-6558a39c-e41b-11e7-9fb1-5c184e52fbb9.png)

## Installation

### From binary

1. Download the latest version from the Releases section (currently only x64 linux versions)
1. Start it and/or add to your service starter

### From source

1. Get a working go environment
1. Get a working node environment (for building the Angular frontend)
1. go get -u github.com/soyersoyer/k20a
1. cd ~/go/src/github.com/soyersoyer/k20a
1. ./build.sh
1. Start it and/or add to your service starter

## Configuration
The configuration filename is k20a.yaml (or an another format what the viper library support)
### Options

|Option|Default|Description|
|---|---|---|
|Listening|:3000|Where should the server listen|
|GeoIPCityFile|/var/lib/GeoIP/GeoLite2-City.mmdb|GeoIP2/GeoLite2 City file|
|GeoIPASNFile|/var/lib/GeoIP/GeoLite2-ASN.mmdb|GeoIP2/GeoLite2 ASN file|
|DataDir|data|Where is the base data dir|
|EnableRegistration|true|Whether registration enabled or not|
|UseBundledWebApp|true|Whether the program should use the bundled webapp or use the frontend/dist folder|

## Limitations
This software is under initial development (0.x) and the database format may change in the future. In other words, it is not guaranteed that the next version of the software will be able to read the the data stored by the current version.

## Coming features
- Admin page
- Filters for pageviews
- Backup solution
- Realtime views
