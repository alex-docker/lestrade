# Lestrade

Lestrade is a tool to be used with Docker.
It injects a unix socket into your containers which can be used for
introspection.

For insance, have you ever wanted to know the name or ID of a container while
inside the container?
Or perhaps the host port for a given exposed port within the container?

Lestrade provides access to this for each of your containers without leaking
information about other containers.
Each container has it's own scoped access to its own information and only its
own information.

## Usage
```bash
# From the host
git clone git@github.com/cpuguy83/lestrade
cd lestrade
./make.sh
./lestrade -g /var/lib/docker -socket /var/run/docker.sock
```

The lestrade socket will be placed into /lestrade.sock within all of your
containers.

There are endpoinds for `/inspect`, `/name`, and `/id` currently.

### Example
```bash
# from within a container after starting lestrade
apt-get update && apt-get install -y socat curl
socat TCP-LISTEN:80,fork,reuseaddr unix:/lestrade.sock # Make this curlable
curl localhost/inspect
```

## Issues
Can't be run in a container.  This is because when bind-mounting in the
docker graph dir (e.g. /var/lib/docker) we are unable to pickup newly
created containers.

Doing things like `docker rm -f <container>` will always return an error
if lestrade is running.
This is because docker will be trying to remove the container before
lestrade can successfully close the introspection socket.

## TODO
Make this a PR to Docker
