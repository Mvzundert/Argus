# :microphone: Twitch-Go: Stream Activity & Now Playing Widget

This is a versatile Go application designed for live streamers. It connects to Twitch to display chat messages, subscriptions, and other activity in your terminal, while also spinning up a real-time "Now Playing" web widget that you can use as a Browser Source in streaming software like OBS.

## Key Features

* **Twitch CLI:** See live chat messages, cheers, and channel point redemptions directly in your terminal.
* **Now Playing Widget:** A browser-source overlay that automatically displays the current song from your media player.
* **Cross-Platform:** The "Now Playing" functionality supports both Linux (`playerctl`) and macOS (`nowplaying-cli`).
* **Simple Setup:** Configuration is handled through a single `.env` file.

---

## :rocket: Getting Started

### Prerequisites
Before running the application, ensure you have the following installed:

* **Go:** The application is written in Go. You can get it from the official [GoLang website](https://golang.org).
* **Media Player CLI:**
    * **Linux**: You need **`playerctl`** to fetch song data. Install it via your distribution's package manager (e.g., `sudo apt install playerctl`).
    * **macOS**: You need **`nowplaying-cli`**. Install it with `brew install nowplaying-cli`.

The Go application will automatically download its other dependencies when you run it.

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
https://id.twitch.tv/oauth2/authorize?response_type=token&client_id=YOUR_CLIENT_ID&redirect_uri=http://localhost&scope=chat%3Aread%20channel%3Aread%3Asubscriptions%20bits%3Aread%20channel%3Aread%3Aredemptions
```

After you authorize the application, your browser will be redirected to http://localhost. The token will be in the address bar's URL fragment. Copy the entire token string and paste it into the TWITCH_TOKEN variable in your .env file. Do not include the oauth: prefix.

## 2. Getting Your App Access Token
This token is for your application to make API calls to create EventSub subscriptions. You will need your Client ID and Client Secret for this.

Find your Client ID and Client Secret in the Twitch Developer Console for the application you registered.

Open your terminal and make the following curl request, replacing the placeholders with your own credentials:

```Bash
curl -X POST 'https://id.twitch.tv/oauth2/token' \
-H 'Content-Type: application/x-www-form-urlencoded' \
-d 'client_id=YOUR_CLIENT_ID&client_secret=YOUR_CLIENT_SECRET&grant_type=client_credentials'
```

The response will be a JSON object containing your access_token. Copy this token and paste it into the TWITCH_APP_ACCESS_TOKEN variable in your .env file.

## 3. Getting Your Twitch Channel ID
The EventSub API requires your Twitch Channel ID (a number), not your channel's name.

If you're already a developer and have access to the Twitch API, you can use the Get Users endpoint to find your Channel ID programmatically. 
This method requires a Client ID and an access token.

### Obtain an Access Token:
You'll need an app access token from the Twitch Developer Console.

### Make an API Call:
Use curl or a similar tool to make a GET request to the Get Users endpoint, passing your username as a query parameter.

```Bash

curl -X GET 'https://api.twitch.tv/helix/users?login=your_twitch_username' \
-H 'Authorization: Bearer <your-access-token>' \
-H 'Client-Id: <your-client-id>'
```

### Parse the Response:
The API will return a JSON response with an id field. This field contains the Channel ID. The response will look something like this:

```JSON
{
  "data": [
    {
      "id": "123456789",
      "login": "your_twitch_username",
      "display_name": "Your Twitch Username",
      ...
    }
  ]
}
```

The id field in the response is your Channel ID. This method is the most reliable and is what you would use in a custom application.

# Running the Application
Once you have created your .env file and filled in all the credentials, you can run the application by navigating to the project directory in your terminal and executing:

```Bash
go run .
```
>The application will start, display a live feed of your Twitch chat in the terminal, and launch a web server on http://localhost:8080 for the "Now Playing" widget.

# Using in OBS-Studio
For the "Now Playing" Widget:

- In OBS Studio, add a new Browser source.
- Set the URL to http://localhost:8080.
- Adjust the width and height to fit your desired overlay.
