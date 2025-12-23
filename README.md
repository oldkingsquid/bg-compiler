# Bytegolf Compiler

The Bytegolf Compiler is a sandbox container orchestration tool for
running code remotely over HTTP. Each code submission is run in a fresh
container that is destroyed after the code finishes executing.

## Usage

The compiler is made to run server side and expose an HTTP API. Based on 
configration seen in the next section the compiler will start X number of 
workers that run code submissions in a fresh container, returning the code 
output to the user.

### Prerequisites

- Docker - [installation instructions](https://docs.docker.com/engine/install/)

### Compiler Configuration and Startup

All Docker images are required on the host machine before code can be remotely
executed using the targeted image. **The compiler will not pull images from
Docker Hub.**

Once either the timeout or max output length is hit, the worker container will
kill the container and start any backlogged requests if present.

```sh
bg-compiler start [flags]
```

```
Start a compiler webserver that takes HTTP requests and 
runs them in a docker container.

Usage:
   start [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--backlog`| 2000 | Job backlog before API rejects requests. |
| `--memory` | 512 | Amount of memory that a container can use for a single run (in MB). |
| `--cpu` | 1024 | CPU shares for a container. |
| `--output` | 30 | Number of bytes that can be read from a container output before the container is killed (in KB). |
| `--timeout` | 30 | Container timeout in seconds. |
| `--workers` | 4 | Number of concurrent workers. |
| `--port` | 8080 | Port to run the compiler on. |
| `--gvisor` | false | Use [gVisor](https://gvisor.dev/) for increased container security (requires runsc). |

### HTTP Calls

Once the server is started and ready to accept HTTP requests, the following
endpoints are available:

#### Compile Request

##### Request

POST `/compile` HTTP/1.1
Content-Type: application/json

```sh
curl --request POST \
  --url {{address}}/compile \
  --header 'Content-Type: application/json' \
  --data '{
	"script": "print('\''hello, world!'\'')",
	"image": "python:3.11.1-alpine3.17",
	"count": 1,
	"cmd": "python3"
}'
```

| Field | Type | Description |
|---|---|---|
| `script` | string | The code to be executed. |
| `image` | string | The Docker image to run the code in. Image must be available on host. |
| `count` | int | The number of times to run the code. |
| `cmd` | string | The command to run the code with. Ex: `python3`, `go run` |
| `extension` | string | The file extension of the code. Ex: `py`, `go` |


##### Response

```json
[
	{
		"stdout": "hello, world!\n",
		"stderr": "",
		"duration_ms": 48,
		"timed_out": false
	}
]
```

| Field | Type | Description |
|---|---|---|
| `stdout` | string | Sandbox stdout. |
| `stderr` | string | Sandbox stderr. |
| `duration_ms` | int | Execution time in milliseconds. |
| `timed_out` | bool | Whether the container timed out or not. |


### Hello, World!

This section walks through setting up the webserver with a couple of workers
and running a simple "Hello, world!" program in Bun.

#### Step 1: Pull Docker Image

Pull the Docker image that will be used to run the code. In this example we
will use the `oven/bun:1` image to run Javascript using Bun.

```sh
docker pull oven/bun:1
```

#### Step 2: Start the Compiler

Start the compiler with 2 workers and a 30 second timeout. Each of the workers 
will run in a container with 512MB of memory and a 1KB output limit on port 8000.

```sh
bg-compiler start \
  --workers 2 \
  --timeout 30 \
  --memory 512 \
  --output 1 \
  --port 8000
```

#### Step 3: Send a Compile Request

Send a compile request to the server. The request will run the code in the
`oven/bun:1` image and print `Hello, world!` to stdout.

Using a [JSON escape tool](https://www.freeformatter.com/json-escape.html#before-output),
transform the following code from

```javascript
console.log("Hello, world!");
```

to

```txt
console.log(\"Hello, world!\");
```

Using cURL, send a compile count of 10 to the server:

```sh
curl --request POST \
  --url http://localhost:8000/compile \
  --header 'Content-Type: application/json' \
  --data '{
	"script": "console.log(\"Hello, world!\");",
	"image": "oven/bun:1",
	"count": 4,
	"cmd": "bun",
	"extension": "js"
}'
```

Output:

```json
[
	{
		"stdout": "hello, world!\n",
		"stderr": "",
		"duration_ms": 34,
		"timed_out": false
	},
	{
		"stdout": "hello, world!\n",
		"stderr": "",
		"duration_ms": 39,
		"timed_out": false
	},
	{
		"stdout": "hello, world!\n",
		"stderr": "",
		"duration_ms": 45,
		"timed_out": false
	},
	{
		"stdout": "hello, world!\n",
		"stderr": "",
		"duration_ms": 43,
		"timed_out": false
	}
]
```

## Installation

- Docker is a prerequisite for running the compiler. [Installation instructions](https://docs.docker.com/engine/install/)

### By Source

If installing by source, Go is required to build the compiler.
[Installation instructions](https://golang.org/doc/install)

Clone the repo

```sh
git clone https://github.com/oldkingsquid/bg-compiler.git
```

Install the compiler

```sh
cd bg-compiler
go install
```

```sh
bg-compiler start
```

### By Binary

Download the latest release from the [releases page](https://github.com/oldkingsquid/bg-compiler/releases)

Run the binary

```sh
bg-compiler start
```