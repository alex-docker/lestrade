# Lestrade

Lestrade is a tool to be used with Docker.
It injects a unix socket into your containers which can be used for
introspection.

For insance, have you ever wanted to know the name or ID of a container while
inside the container?
Or perhaps the host port for a given exposed port within the container?

Lastrade provides access to this for each of your containers without leaking
information about other containers.
Each container has it's own scoped access to its own information and only its
own information.

## Usage
```bash
git clone git@github.com/cpuguy83/lestrade
cd lestrade
docker build -t lestrade .
docker run -d \
  -v /var/lib/docker:/docker \
  -v /var/run/docker.sock:/var/run/docker.sock \
  lestrade
```

The lestrade socket will be placed into /lestrade.sock within all of your
containers.

There are endpoinds for `/inspect`, `/name`, and `/id` currently.

### Example
```bash
# from within a container after starting lestrade
apt-get update && apt-get install -y socat curl
socat TCP-LISTEN:80,fork,reuseaddr unix:/lestrade.sock
curl localhost/inspect
```

## TODO
1. Watch the docker event stream to capture new containers
