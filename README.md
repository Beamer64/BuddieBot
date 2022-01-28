<h1 align="center">
  <br>
  <img src="https://images.unsplash.com/photo-1563207153-f403bf289096?ixid=MnwxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8&ixlib=rb-1.2.1&auto=format&fit=crop&w=1051&q=80" width="500" height="350" alt="">
  <br>
  BuddieBot
  <br>
</h1>

<h3 align=center>A Homemade Discord Bot for Golang practice and development...also for funsies.</a></h3>

<div align="center">
  <a href="http://www.harleyroper.com/" target="_blank">
    <img src="https://img.shields.io/badge/Check%20out-My%20Website!-brightgreen" alt="shield.png">
  </a>
  <a href="https://golang.org" target="_blank">
    <img src="https://img.shields.io/badge/Made%20with-%20GO-blue" alt="shield.png">
  </a>
  <a href="https://github.com/Beamer64/DiscordBot/blob/master/LICENSE" target="_blank">
    <img src="https://img.shields.io/github/license/beamer64/DiscordBot" alt="shield.png">
  </a>
</div>

<div align="center">
  <a target="_blank">
    <img src="https://img.shields.io/tokei/lines/github/Beamer64/DiscordBot?color=maroon" alt="shield.png">
  </a>
</div>

---

<p align="center">
  <a href="#about">About</a>
  •
  <a href="#features">Features</a>
  •
  <a href="#installation">Installation</a>
  •
  <a href="#setting-up">Setting Up</a>
  •
  <a href="#terms">Terms</a>
  •
  <a href="#license">License</a>
  •
  <a href="#credits">Credits</a>
</p>

<div align="center">
<img src="https://github.com/Beamer64/DiscordBot/blob/master/res/under-construction-tape-png-program-under-construction-removebg-preview.png" width="770" height="250" alt="">
</div>

## About

BuddieBot is an open source Discord bot created by two people that wanted to practice their Golang skills and develop something fun and useful for them. We enjoy growing and developing BuddieBot
every day and play to do so for the foreseeable future! The goal of this project is to incorporate as many cool and fun features as we can find. We use BuddieBot in out personal servers every day!

If you liked this repository, feel free to leave a star ⭐ to help promote BuddieBot!

---

## Features

10+ commands and counting!

BuddieBot also comes packed with a variety of features, such as:

* Play/Stop/Queue music from **YouTube** links
* **Starting/Stopping** a Minecraft Server
* Receive your daily **Horoscope**
* **Insult** your friends
* **Slash Commands!**
* **Moderator only** commands
* And much more!

---

## Installation

**You can invite BuddieBot to your server with the** [Temporarily Removed]
link. 🤖😁 Alternatively, you can clone this repo and host the bot yourself.

<!-- Server Invite: [Invite to Server](https://discord.com/api/oauth2/authorize?client_id=866151939472883762&permissions=8&redirect_uri=https%3A%2F%2Fgithub.com%2FBeamer64%2FDiscordBot&response_type=code&scope=bot%20identify%20email%20connections%20applications.commands%20guilds%20guilds.join%20gdm.join%20messages.read) -->

```
git clone https://github.com/Beamer64/DiscordBot.git
```

After cloning, run an

```
go get ./...
```

to snag all the dependencies.

---

## Setting Up

You have to create a `config.yaml` file in order to run the bot (you can use the example file provided as a base). Your file should look something like this:

```
# Tokens/API Keys
keys:
  botToken: ""
  webHookToken: ""
  botPublicKey: ""
  tenorAPIkey: ""
  insultAPI: "https://evilinsult.com/generate_insult.php?lang=en&type=json"

# IDs relating to Discord or Bot
discordIDs:
  webHookID: ""
  errorLogChannelID: ""

# Custom Settings
settings:
  botPrefix: ""
  botAdminRole: ""

# VM Server Info
vm:
  sshKeyBody: ""
  machineIP: ""
  type: "service_account"
  project_id: ""
  private_key_id: ""
  private_key: ""
  client_email: ""
  client_id: ""
  auth_uri: "https://accounts.google.com/o/oauth2/auth"
  token_uri: "https://oauth2.googleapis.com/token"
  auth_provider_x509_cert_url: "https://www.googleapis.com/oauth2/v1/certs"
  client_x509_cert_url: ""
  zone: ""
```

*The 'vm' section is for the GCP VM hosting of the Minecraft server. This can be omitted if unused.*

Visit the Discord [developer portal](https://discordapp.com/developers/applications/) to create an app and use the client token you are given for the `token` option. To get keys for supported APIs,
visit:

* [Tenor API](https://tenor.com/gifapi/documentation)
* [Google APIs](https://console.developers.google.com/apis/)
* [Insult API](https://evilinsult.com/api/)

After your `config.yaml` file is built, you have to enable `Privileged Intents` on your Discord [developer portal](https://discordapp.com/developers/applications/). You can find these intents under
the "Bot" section, and there are two ticks you have to switch on. For more information on Gateway Intents, check out [this](https://discordpy.readthedocs.io/en/latest/intents.html) link.

Once done, feel free to launch BuddieBot using the command `go run cmd/discord-bot/main.go`.

---

## To-Do

BuddieBot is in a continuous state of development. New features/updates may come at any time. Some pending ideas are:

- [ ]  Games
- [ ]  Be Funnier
- [ ]  Skip songs
- [ ]  Multiple Music Sources
- [ ]  Rename Repo
- [X]  DM Your Mother
- [X]  Convert most commands to embeds
- [X]  Play/Queue Music
- [X]  Slash Commands
- [X]  Custom tag/reaction system

---

## Terms

- *Guild* - This is what Discord refers to your server as. Servers are 'Guilds'.
- *botToken* - Given when a new bot is created. Located in the [Bot section](https://discord.com/developers/applications/866151939472883762/bot) of the Discord Dev portal.
- *webHookToken* - Can be easily found in the Webhook URL.††
- *botPublicKey* - Given when a new bot is created. Located in the [Gen Info section](https://discord.com/developers/applications/866151939472883762/information) of the Discord Dev portal.
- *webHookID* - Can be easily found in the Webhook URL.††
- *errorLogChannelID* - The ID of the Channel you'd like the bot to update with any errors it encounters.†
- *botPrefix* - The prefix given when the bot recognizes a command. For example, the one I use is '$'.
- *botAdminRole* - The name of the Role you create to restrict users from certain commands or actions. E.g. Mine was 'Bot Admin Role'.

† **To get the ID's of things in Discord, you will need to [Enable Dev Mode](https://techswift.org/2020/09/17/how-to-enable-developer-mode-in-discord) in Discord.**

†† **You will need to [Create a Webhook](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks) in Discord first. Then the ID and Token respectively can be found in the Webhook
URL. E.g. `https://discord.com/api/webhooks/WEBHOOK_ID/WEBHOOK_TOKEN`**

---

## License

Released under the [GNU GPL v3](https://www.gnu.org/licenses/gpl-3.0.en.html) license.

---

## Credits

This is just a list of various credits to any person(s) whose work are contributed to this open source project.

### To give credit where credit is due 😁

* **Wyatt Shuler** - *Co-contributor* - [Github](https://github.com/Saberr43), [Website](http://www.shuler.io/)

---

<img src="https://www.gstatic.com/tenor/web/attribution/PB_tenor_logo_blue_horizontal.png" width="600" height="100"  alt=""/>

The Tenor API is used to deliver gifs from the DiscordBot. The website can be found [here](https://tenor.com/). The API site can be found [here](https://tenor.com/gifapi/documentation#quickstart).

---

<img src="https://www.horoscope.com/images-US/horoscope-logo.svg" width="500" height="150"  alt=""/>

The horoscope.com daily horoscopes are used to allow our DiscordBot to deliver a daily horoscope to our users. The website can be found [here](https://www.horoscope.com/us/index.aspx).

---

<img src="https://image.flaticon.com/icons/png/512/4698/4698787.png" width="128" height="128"  alt=""/>

The icon for BuddieBot was made by [wanicon](https://www.flaticon.com/authors/wanicon) from www.flaticon.com

---

# [bwmarrin/discordgo](https://github.com/bwmarrin/discordgo)

We utilize the convenience of bwmarrin's discordgo Golang package in a majority of the DiscordBot's code.

> DiscordGo is a Go package that provides low level bindings to the Discord chat client API. DiscordGo has nearly complete support for all of the Discord API endpoints, websocket interface, and voice interface.

---

# [gocolly/colly](https://github.com/gocolly/colly)

Gocolly's colly framework is used to simplify any crawlers/scrapers in our DiscordBot.

> Lightning Fast and Elegant Scraping Framework for Gophers Colly provides a clean interface to write any kind of crawler/scraper/spider. With Colly you can easily extract structured data from websites, which can be used for a wide range of applications, like data mining, data processing or archiving.

---
