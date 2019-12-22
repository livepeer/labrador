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

## Stream-sender

### API

#### GET /stats/select

Retrieves the statistics for a specific stream 

```
curl <host>:5000/stats/select -X GET -H "Content-Type: application/json" -d '{"base_manifest_id": "<manifest id>"}'
```

#### GET /stats/all

Retrieves the statistics for all streams

```
curl <host>:5000/stats/all
```

#### POST /stream/start

Start a stream 