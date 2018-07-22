# tubetoplex

A tool to automatically download youtube and other online videos and
import them as episodes into a Custom TV Show in Plex.

## Background

There are a lot of videos available for online streaming, but I want to
watch these videos during my commute, where I don't have internet
access.
I save links to these videos I find on a tumblr blog at
https://softwaredevvideos.tumblr.com/

This tool will automatically find new videos on the blog that it has not
processed yet, download them, move them to a directory of a Custom TV Show,
rename them to be picked up in a particular season, and create an NFO file with
information about the "episode".

## Development

Dependency management is handled by [dep](https://github.com/golang/dep)

Install on MacOS with `brew install dep`

Update dependencies with `dep ensure`

## Running

Copy `run.sh.sample` to `run.sh` and modify for your environment.

Build a docker container `docker build -t tubetoplex .`

Run in interactive mode to see logs as they happen.
`--privileged` is used to allow mounting a windows shared folder from
the network on a windows host.
`docker run --rm --privileged -it tubetoplex`
