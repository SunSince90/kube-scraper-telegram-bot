# Telegram BOT Listener

A telegram bot living inside Kubernetes that listens for messages to send to
users.

Part of a personal project, *kube-scraper* (link coming soon).

## Overview

The telegram bot listener is a server that lives inside a Kubernetes cluster,
waiting for messages to send to users.
It exists mostly as an exercise sharpen my skills in Kubernetes, go, gRPC and
other technologies and to escape boredom.

Here's how it is implemented:

* gRPC server with SendMessage method
* Firestore as a backend
* Kubernetes YAML files
* Made to run on a Raspberry PI 4

TODO: UPDATE this readme with better information

## Run

Create the docker file and push it to your preferred registry:

```bash
make docker-build docker-push IMG=<repository>
```

Provide the correct values in the yamls included in `deploy` or create them
by using `kubectl create secret generic <name>` and options
`--from-literal=...` or `--from-file` based on the enclosed secret.
On `06_deployment.yaml` specify the image name.

Then: `kubectl apply -f ./deploy`.

## Firestore chats format

TODO: explain it.

Take a look at the files inside `listenerserv` to learn how a chat is stored.

## Contacts

Feel free to contact and/create issues if you're using this in your project
and need help.
