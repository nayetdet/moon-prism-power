# Moon Prism Power

Moon Prism Power is a command-line utility for comparing and migrating anime
and manga libraries from AniList to MyAnimeList.

## First-time setup

### 1. Create a MyAnimeList application

Create an application at <https://myanimelist.net/apiconfig> and set the
redirect URL to:

```text
http://127.0.0.1:3939/callback
```

### 2. Configure the project

Create your local environment file:

```sh
cp .env.example .env
```

Fill in the AniList username and MyAnimeList client ID:

```dotenv
ANILIST_USER=your_anilist_username
MAL_CLIENT_ID=your_mal_client_id
MAL_REFRESH_TOKEN=
```

### 3. Authorize MyAnimeList

Run:

```sh
make mpp-auth
```

The command opens the browser and prints a refresh token after authorization.
Copy the printed line into `.env`:

```dotenv
MAL_REFRESH_TOKEN=your_refresh_token
```

## Run the migration

```sh
make mpp
```

The migration shows a preview before making changes. Confirm the prompt to
continue, or press Enter to cancel.

The browser is only used by `mpp-auth`. `mpp` refreshes the access token in
memory when it starts, so no access token needs to be stored in `.env`.

If the refresh token becomes invalid, run `make mpp-auth` again and replace it
in `.env`.
