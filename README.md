<br/>
<p align="center">
  <a href="https://github.com/losuler/futaba">
    <img src="img/futaba.png" alt="logo" width="150" height="150">
  </a>

  <p align="center">
    <h3 align="center">Futaba</h3>
    <p align="center">
      A silly Discord bot for a friend.
    </p>
  </p>
</p>

## Commands

- Show the current time of a user on your server.

    ```
    t.user
    time.user
    ```

## Configuration

The configuration file `config.yaml` has two main sections (see `config.yaml.example`).

The `token` is the token for the bot (see [create a bot](#create-a-bot)).

```yaml
discord:
    token: 1234567890
```

Each list entry refers to a user on the server.

- On Linux you can use `timedatectl --list-timezones` to find the correct timezone.

```yaml
users:
    - username: name
      userid: 1234567890
      timezone: America/Los_Angeles
      nicknames: nick
```

## Create a bot

1. Browse to the [Discord Developer Portal](https://discordapp.com/developers/applications).

2. Click `New Application`.

3. Provide a name (can be different to the name of the bot itself).

4. Click `Bot` on the left side menu, then click `Add Bot`.

5. Under `Token`, click `Click to Reveal Token` to reveal the bot's token (used in `config.yaml`).

## Add a bot to a server

1. Replace `CLIENT_ID` with the client ID of the application (navigate to `General Information` 
on the left side menu).

    ```
    https://discordapp.com/oauth2/authorize?&client_id=CLIENT_ID&scope=bot&permissions=0
    ```

2. Select your server from the drop down menu.