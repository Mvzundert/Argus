# Twitch Go
This is a go based command-line-interface (CLI) application that connects to Twitch's IRC server to display chat messages.

# Prerequisite
Before you can run this application, you need to have the following installed:

-  **Go:** The application is written in Go. You can download and install it from [GoLang](https://golang.org)
- **Go Libraries:** You need a few Go packages to handle enviroment variables.

```Bash
    go get [github.com/joho/godotenv](https://github.com/joho/godotenv)
```

# Configuration

``` Bash
# Your Twitch username
TWITCH_NICK="your_username"

# Your Twitch OAuth token.
TWITCH_TOKEN="oauth:your_oauth_token"

# The Twitch channel you want to join (e.g., #twitch)
TWITCH_CHANNEL="#your_channel_name"

# Your Twitch Channel ID.
TWITCH_CHANNEL_ID="your_channel_id"
```

# Getting your Credentials
You need to obtain a Twitch oAuth token and your Twitch Channel ID to configure the application.

1. Getting the Twitch oAuth token
For this application, you need a User Access Token with specific permissions (scopes).

**Do not use a third-party service that asks for your Client Secret!!** 

The most secure way is to use Twitch's official Implicit Grant Flow.

Register a new application in the **[Twitch Developer Console](https://dev.twitch.tv/console/)**.

2. Set the OAuth Redirect URL to ```http://localhost.```

3. Open the following URL in your web browser, replacing YOUR_CLIENT_ID with the ID of the application you just registered:

```Bash
https://id.twitch.tv/oauth2/authorize?response_type=token&client_id=YOUR_CLIENT_ID&redirect_uri=http://localhost&scope=chat%3Aread
```

4. After you authorize the application, your browser will be redirected to http://localhost, and the token will be in the address bar. Copy the entire token, including the oauth: prefix.

# Getting your Twitch Channel ID
The PubSub system requires your Twitch Channel ID (a number), not your channel's name.

1. Go to the application in the **[Twitch Developer Console](https://dev.twitch.tv/console/)**.
2. Click manage
3. Should find your client ID there.
