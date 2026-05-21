span

<h3 align=center>A Homemade Discord Bot for Golang practice and development...also for funsies.</a></h3>

<div align="center">
  <a href="http://www.harleyroper.com/" target="_blank">
    <img src="https://img.shields.io/badge/Check%20out-My%20Website!-brightgreen" alt="shield.png">
  </a>
  <a href="https://golang.org" target="_blank">
    <img src="https://img.shields.io/badge/Made%20with-%20GO-blue" alt="shield.png">
  </a>
  <a href="https://github.com/Beamer64/BuddieBot/blob/master/LICENSE" target="_blank">
    <img src="https://img.shields.io/github/license/beamer64/BuddieBot" alt="shield.png">
  </a>
</div>

<div align="center">
  <a target="_blank">
    <img src="https://img.shields.io/badge/Total%20Lines-9812-maroon.svg" alt="shield.png">
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
  <a href="#release-notes">Release Notes</a>
  •
  <a href="#license">License</a>
  •
  <a href="#credits">Credits</a>
</p>

<div align="center">
<img src="https://github.com/Beamer64/BuddieBot/blob/master/res/repo_imgs/under-construction-tape-png-program-under-construction-removebg-preview.png" width="770" height="250" alt="">
</div>

## About

BuddieBot is an open source Discord bot written in Golang that I initially created to develop and maintain new Golang technologies and practices. What originally started as a fun
idea and grown into a full personal pet project. I love working on BuddieBot to improve it with new features and utilities all the time. I enjoy using BuddieBot in my personal
servers and plan to continue growing this bot for the foreseeable future!

If you liked this repository, feel free to leave a star ⭐ to help promote BuddieBot!

---

## Features

130+ commands and counting!

BuddieBot also comes packed with a variety of features, such as:

* Play/Stop/Queue music from **YouTube** links
* Receive your daily **Horoscope**
* **Insult** your friends
* **Slash Commands!**
* **Moderator only** commands
* Manipulate **Text**
* **GAMES!**
* And much more!

---

## Installation

**You can invite BuddieBot to your server with the** [Temporarily Removed]
link. 🤖😁 Alternatively, you can clone this repo and host the bot yourself.

<!-- Server Invite: [Invite to Server](https://discord.com/api/oauth2/authorize?client_id=866151939472883762&permissions=8&redirect_uri=https%3A%2F%2Fgithub.com%2FBeamer64%2FBuddieBot&response_type=code&scope=bot%20identify%20email%20connections%20applications.commands%20guilds%20guilds.join%20gdm.join%20messages.read) -->

```
git clone https://github.com/Beamer64/BuddieBot.git
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
  prodBotToken:   ""
  testBotToken:   ""
  webHookToken:   ""
  botPublicKey:   ""
  tenorAPIkey:    ""
  steamAPI:       "http://api.steampowered.com/ISteamApps/GetAppList/v0002/?key=STEAMKEY&format=json"
  affirmationAPI: "https://www.affirmations.dev/"
  kanyeAPI:       "https://api.kanye.rest/"
  adviceAPI:      "https://api.adviceslip.com/advice"
  doggoAPI:       "https://api.thedogapi.com/v1/images/search"
  albumPickerAPI: "http://recommended-album-api-dev.us-east-1.elasticbeanstalk.com/prediction/"
  wyrAPI:         "https://would-you-rather-api.abaanshanid.repl.co/"

# IDs relating to Discord or Bot
discordIDs:
  currentBotAppID:     "" # Don't set. This gets set later.
  prodBotAppID:        ""
  testBotAppID:        ""
  masterGuildID:       ""
  testGuildID:         ""
  webHookID:           ""
  errorLogChannelID:   ""
  eventNotifChannelID: ""

# Custom Settings
settings:
  botPrefix:     ""
  botAdminRole:  ""
  # Used for email sending
  email:         ""
  emailPassword: ""

database:
  tableName: ""
  region:    ""
  accessKey: ""
  secretKey: ""
```

*Sections like "database" and various "IDs" are used for specfic functions or features. This can be omitted if unused. That part of the bot just wont work until supplied.*

Visit the Discord [developer portal](https://discordapp.com/developers/applications/) to create an app and use the client token you are given for the `token` option. To get keys
for supported APIs,
visit:

* [Tenor API](https://tenor.com/gifapi/documentation)

After your `config.yaml` file is built, you have to enable `Privileged Intents` on your Discord [developer portal](https://discordapp.com/developers/applications/). You can find
these intents under
the "Bot" section, and there are two ticks you have to switch on. For more information on Gateway Intents, check out [this](https://discordpy.readthedocs.io/en/latest/intents.html)
link.

Once done, feel free to launch BuddieBot using the command `go run cmd/discord-bot/main.go`.

### Audio (Lavalink) — required for `$play` / `$stop`

Audio playback is handled by [Lavalink](https://lavalink.dev/), a Java-based audio service. The bot doesn't open its own
voice connection — Lavalink does, which is what gives the bot DAVE/E2EE voice support. In **debug mode** the bot spawns
Lavalink as a child process automatically; in **production** you run Lavalink as a separate service alongside the bot.

#### Dev setup (one-time, on your coding PC)

1. Install Java 17+:
   * **Windows:** `winget install Microsoft.OpenJDK.21`
   * **Linux:** `sudo apt install openjdk-21-jre-headless`
   * Verify with `java -version`.
2. Create a `lavalink/` directory at the repo root.
3. Download `Lavalink.jar` from the [latest 4.x release](https://github.com/lavalink-devs/Lavalink/releases) into `lavalink/`.
4. Create `lavalink/application.yml`:

```yaml
server:
  port: 2333
  address: 127.0.0.1   # localhost-only — never expose this port

plugins:
  youtube:
    enabled: true
    allowSearch: true
    allowDirectVideoIds: true
    allowDirectPlaylistIds: true
    clients:
      - MUSIC
      - ANDROID_VR
      - WEB
      - WEBEMBEDDED

lavalink:
  plugins:
    - dependency: "dev.lavalink.youtube:youtube-plugin:1.13.5"
      repository: "https://maven.lavalink.dev/releases"
  server:
    password: "youshallnotpass"
    sources:
      youtube: false   # built-in YouTube is deprecated; the plugin above replaces it
      soundcloud: true
      bandcamp: true
      twitch: true
      vimeo: true
      http: true
      local: false

logging:
  level:
    root: INFO
    lavalink: INFO
```

5. The first debug launch is slow (~30s) while Lavalink downloads the YouTube plugin from Maven. Subsequent launches are fast.

When you run the bot in your debugger, it auto-spawns Lavalink and reaps the Java child on shutdown (signal-based on
graceful stop; OS-level mechanisms — `Pdeathsig` on Linux, Job Objects on Windows — handle hard kills). If a stuck
Java process ever survives, kill it manually:

* **Windows:** `taskkill /IM java.exe /F`
* **Linux:** `pkill -f Lavalink.jar`

#### Production deployment (same host as the bot)

The bot and Lavalink run as two independent processes on the same machine. Run Lavalink as a systemd unit:

```ini
# /etc/systemd/system/lavalink.service
[Unit]
Description=Lavalink
After=network.target

[Service]
Type=simple
User=lavalink
WorkingDirectory=/opt/lavalink
ExecStart=/usr/bin/java -Xmx512M -jar Lavalink.jar
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Lavalink stays bound to `127.0.0.1`, so the port is never reachable from outside the host — no firewall rules needed.
Resource budget: ~256–512 MB RAM on top of whatever the bot uses. The bot's `lavalink:` config in `config.yaml` points at
`127.0.0.1:2333` (same as dev). Bot updates and Lavalink updates are independent — restarting one doesn't require
restarting the other.

---

## To-Do

BuddieBot is in a continuous state of development. New features/updates may come at any time. Some pending ideas are:

- [X]  Games
- [ ]  Be Funnier
- [ ]  Skip songs
- [ ]  Multiple Music Sources
- [X]  Rename Repo
- [X]  DM Your Mother
- [X]  Custom tag/reaction system
- [ ]  Server specific settings
- [ ]  BuddieBot Website
- [X]  Txt commands

---

## Terms

- *Guild* - This is what Discord refers to your server as. Servers are 'Guilds'.
- *botToken* - Given when a new bot is created. Located in the [Bot section](https://discord.com/developers/applications/866151939472883762/bot) of the Discord Dev portal.
- *webHookToken* - Can be easily found in the Webhook URL.††
- *botPublicKey* - Given when a new bot is created. Located in the [Gen Info section](https://discord.com/developers/applications/866151939472883762/information) of the Discord Dev
  portal.
- *webHookID* - Can be easily found in the Webhook URL.††
- *errorLogChannelID* - The ID of the Channel you'd like the bot to update with any errors it encounters.†
- *botPrefix* - The prefix given when the bot recognizes a command. For example, the one I use is '$'.
- *botAdminRole* - The name of the Role you create to restrict users from certain commands or actions. E.g. Mine was 'Bot Admin Role'.

† **To get the ID's of things in Discord, you will need to [Enable Dev Mode](https://techswift.org/2020/09/17/how-to-enable-developer-mode-in-discord) in Discord.**

†† **You will need to [Create a Webhook](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks) in Discord first. Then the ID and Token respectively can be
found in the Webhook
URL. E.g. `https://discord.com/api/webhooks/WEBHOOK_ID/WEBHOOK_TOKEN`**

---

## Release Notes

Release notes (when updated) can be found in the res folder of the project or at:
https://github.com/Beamer64/BuddieBot/blob/master/res/release.md

---

## License

Released under the [GNU GPL v3](https://www.gnu.org/licenses/gpl-3.0.en.html) license.

---

## TOS

The Terms of Service for BuddieBot can be found [Here](https://github.com/Beamer64/BuddieBot/blob/master/res/legal/TOS.md).

---

## Privacy Policy

The Privacy Policy for BuddieBot can be found [Here](https://github.com/Beamer64/BuddieBot/blob/master/res/legal/priv_policy.md).

---

## Credits

This is just a list of various credits to any person(s) whose work are contributed to this open source project.

### To give credit where credit is due 😁

<img src="https://www.gstatic.com/tenor/web/attribution/PB_tenor_logo_blue_horizontal.png" width="600" height="100"  alt=""/>

The Tenor API is used to deliver gifs from the BuddieBot. The website can be found [here](https://tenor.com/). The API site can be
found [here](https://tenor.com/gifapi/documentation#quickstart).

---

<img src="https://www.horoscope.com/images-US/horoscope-logo.svg" width="500" height="150"  alt=""/>

The horoscope.com daily horoscopes are used to allow our BuddieBot to deliver a daily horoscope to our users. The website can be
found [here](https://www.horoscope.com/us/index.aspx).

---
