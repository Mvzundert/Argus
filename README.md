# Twitch Go
This is a Go-based command-line interface (CLI) application that connects to both Twitch's IRC chat server and the EventSub API to display chat messages and activity feed events (subscribers, cheers, and channel points redemptions).

# Prerequisites
Before you can run this application, you need to have the following installed:

**Go:** The application is written in Go. You can download and install it from [GoLang](https://golang.org)

**Go Libraries:** You need a few Go packages to handle environment variables and WebSocket connections. You can get them by running the following commands in your terminal:

go get [github.com/joho/godotenv](https://github.com/joho/godotenv)
go get [github.com/gorilla/websocket](https://github.com/gorilla/websocket)

# Configuration
You must create a .env file in the same directory as the Go application with the following environment variables.

```Bash
# Your Twitch username
TWITCH_NICK="your_username"

# Your Twitch User Access Token with required scopes.
TWITCH_TOKEN="your_user_token"

# Your Twitch Client ID. Found in the Twitch Dev Console.
TWITCH_CLIENT_ID="your_client_id"

# The Twitch channel you want to join (e.g., #twitch)
TWITCH_CHANNEL="#your_channel_name"

# The numeric ID for your Twitch channel.
TWITCH_CHANNEL_ID="your_channel_id"
```

# Getting Your Credentials
You need to obtain three pieces of information to configure the application: your User Access Token, your Client ID, and your Twitch Channel ID.

## 1. Getting Your User Access Token
This token is for reading chat messages and is specific to your user account. Do not share your Client Secret with any third-party service. The most secure way to get this token is by using Twitch's official Implicit Grant Flow.

Register a new application in the Twitch Developer Console.

Set the OAuth Redirect URL to ```http://localhost```

Open the following URL in your web browser, replacing YOUR_CLIENT_ID with the ID of the application you just registered:

```Bash
[https://id.twitch.tv/oauth2/authorize?response_type=token&client_id=YOUR_CLIENT_ID&redirect_uri=http://localhost&scope=chat%3Aread%20channel%3Aread%3Asubscriptions%20bits%3Aread%20channel%3Aread%3Aredemptions](https://id.twitch.tv/oauth2/authorize?response_type=token&client_id=YOUR_CLIENT_ID&redirect_uri=http://localhost&scope=chat%3Aread%20channel%3Aread%3Asubscriptions%20bits%3Aread%20channel%3Aread%3Aredemptions)
```

After you authorize the application, your browser will be redirected to http://localhost. The token will be in the address bar's URL fragment. Copy the entire token string and paste it into the TWITCH_TOKEN variable in your .env file. Do not include the oauth: prefix.

## 2. Getting Your App Access Token
This token is for your application to make API calls to create EventSub subscriptions. You will need your Client ID and Client Secret for this.

Find your Client ID and Client Secret in the Twitch Developer Console for the application you registered.

Open your terminal and make the following curl request, replacing the placeholders with your own credentials:

```Bash
curl -X POST '[https://id.twitch.tv/oauth2/token](https://id.twitch.tv/oauth2/token)' \
-H 'Content-Type: application/x-www-form-urlencoded' \
-d 'client_id=YOUR_CLIENT_ID&client_secret=YOUR_CLIENT_SECRET&grant_type=client_credentials'
```

The response will be a JSON object containing your access_token. Copy this token and paste it into the TWITCH_APP_ACCESS_TOKEN variable in your .env file.

## 3. Getting Your Twitch Channel ID
The EventSub API requires your Twitch Channel ID (a number), not your channel's name.

You can find your channel ID by visiting your Twitch profile URL and copying the numbers after ```bash https://www.twitch.tv/directory/game/```

Alternatively, you can use a third-party tool like Twitch User ID Converter to find your ID.

# Running the Application
Once you have created your .env file and filled in all the credentials, you can run the application by navigating to the project directory in your terminal and executing:

```Bash
go run .
```

