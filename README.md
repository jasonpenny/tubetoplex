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

### TODO

- [ ] Setup database to store videos
- [ ] Pull a page of posts from tumblr, extracting a link and tag
- [ ] Sync posts to database

- [ ] given a video url and tag, get info
- [ ] given a video url and tag, download video
- [ ] After downloading the video, rename to include a season and
  episode number
- [ ] Create episode NFO file

- [ ] Check for new posts on a schedule (like every hour), and sync to
  database
- [ ] Check database for new videos on a schedule, kick off download
  task
