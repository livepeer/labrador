# LABRADOR :dog:

Labrador is an automated broadcast testing and monitoring tool.

## Services
* stream-tester `:3001`
* stream-sender `:3002`
* dashboard `:3003` 
* grafana `:3004`

## Running Labrador

```
docker-compose up
```

Environment variables can be set by prepending them to the command above or editing the `.env`file in the project root

```
# Default Config
KEYSTORE_DIR=~/.lpData/rinkeby/keystore
DASHBOARD_EXTERNAL_URL=localhost:3002
```

- `KEYSTORE_DIR` - Labrador will preload the broadcaster container with a keystore directory, as a non-interactive terminal does not support ethereum account creation for `go-livepeer`. For ease of setup it is advisable to have an account pre-funded with test ETH for the test network you're using available here.

- `STREAMSENDER_EXTERNAL_URL` - The dashboard are static files served on an HTTP server. This means the javascript is executed in the user's browser and therefore the `stream-sender` API URL needs to be configured to be the external URL (or localhost when running locally). A reference to a container name/IP will try to lookup the running container on the user's host system, which wouldn't exist when the setup is ran from a cloud VM. 

- `ETHEREUM_NETWORK` - Ethereum network to use, currently only `rinkeby` is supported.

- `ETHEREUM_PASSPHRASE` - Password to unlock the ethereum account in `KEYSTORE_DIR`. This value can be either a path to file or a password. For security reasons it is advisable to use a file path.

### Enabling Experimental Verification

Verification is done by a pre-trained [machine learning classifier](https://github.com/livepeer/verification-classifier) model and is ran as an addition service in the docker network.

To enable verification add `-verifierAddr http://verifier:5000/verify -verifierPath /root/stream` to `BROADCASTER_CFG` in `.env`.

## Stream-sender

### API

#### GET /stats/select

Retrieves the statistics for a specific stream 

```
curl <host>:3002/stats/select -X GET -H "Content-Type: application/json" -d '{"base_manifest_id": "<manifest id>"}'
```

#### GET /stats/all

Retrieves the statistics for all streams

```
curl <host>:3002/stats/all
```

#### POST /stream/start

Start a stream, takes in following parameters:

```
{
    "host": "localhost", // broadcaster HTTP address
    "file_name": "official_test_source_2s_keys_24pfs.mp4", // must be present in stream-tester root dir
    "rtmp": 1935, // broadcaster RTMP port
    "media": 8935, // broadcaster HTTP port
    "repeat": 1, // number of times to repeat the stream
    "simultaneous": 1, // concurrent streams
    "profiles_num": 2, // number of requested renditions
    "do_not_clear_stats": false // will be overwritten to 'false' by the server
}
```

On a succesful request returns 

```
{
    "success": true,
    "base_manifest_id": <base_manifest_id">
}
```
