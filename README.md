## Overview

### How does this work?

Well, its been established that we want to convert one url to different URLs on different streaming platforms.

When an incoming URL has been receieved (say a Spotify link), its metadata is first retrieved. How? By using the API endpoint of the incoming platform (in this case, Spotify). It gets the metadata by using the search functionality on the platform and returns the first result. There is very little possibility that the first returned result wouldn't be what we're expecting. I am personally yet to face this. I'd update this in a heartbeat if this turns out to be wrong.

After the metadata has been retrieved, its used to retrieve results from other platforms. In this case, after the metadata about the track (artiste, album, title, etc) has been returned, it searches on Deezer (and other platforms) to get the track on these platforms. These results are then returned altogether.

In the case of playlists, it does something similar except that when a playlist link has been pasted, it'll fetch all the tracks under the playlist, then use it to look for the tracks on other platforms.

### Okay, you seem to talk about making requests a lot and iterating over playlists. _What about the performance?_

In order to make things faster, it actually caches **ALL** tracks that have been searched. So in the case where one wants to search for a new track or playlist, it first check the cache (the cache used here is good ol redis) to see if the track has already been searched. It fetches if it has been. This makes things blazing for commonly shared/searched tracks.

Also, this codebase uses goroutines to run things. It **concurrently** searches for tracks on platforms using goroutines _AND_ returns the results using Go channels. For people unfamiliar with Go, this is pretty straightforward (dont be scared by the technical jargons). The aim of this project is to help understand Go (if you're not familiar with it) and to provide a (fun) project to work on (if you're already familiar with Go).

Then, it uses websockets to make sure things dont crawl to a stop. It uses websocket to allow for multi-client support. While its fast enough using a REST API (yes, there are also REST API options for use in things like twitter bots — coming soon), it can slow things down. I want it to be as fast as possible, So it uses websocket to get the URL which it passes to the backend which in turn uses goroutines and channels to run things.

So to recap, its fast enough. It can be faster though but right now, speed isnt exactly a problem. **An average uncached request takes about 3-5s for playlists with like 100 or more tracks and even less for single tracks AND WAAAYYY LESS FOR CACHED SEARCHES.** Its does this using the following:

- Caching using redis
- Goroutines and channels to run and send results back concurrently (this is a special Go perk, IMO. don't sleep on it 😜)
- Websockets

### Ayy dude, whats up with the prisma DB setup?

Well, I had something else in mind that time. I still have plans to do things relating to user accounts and stuff so I guess it remains as it is (for now at least).

### So how do I get started?

Create a .env file in the project root.

Please create a Deezer app [here](https://developers.deezer.com/myapps/create), then create a spotify app (here)[https://developer.spotify.com/dashboard/applications]. Copy the credentials provided and update the .env file accordingly. Before you get started, **please set your env to dev.** You can use

```bash
export ENV=dev
```

in your terminal if you use a \*NIX machine (Linux/MacOS). Follow Prisma setup [here](https://github.com/prisma/prisma-client-go) if you have issues with prisma. Do not hesitate to open an issue also

Build the whole thing and you should be good to go.

```bash
go build -o app && ./app
```

### And the roadmap?

More platforms to be supported. Then development of separate libraries/wrappers around the REST API of these platforms (esp Deezer). This is mostly for better developer experience. Also, finally cleanup the codebase where necessary. What I know is, I'd definitely try to work on this anytime I can.

A big shoutout to [Amaka](https://github.com/ammiecodes). All designs and assets/icons are courtesy of her. It wont look like "Zoove" without her magic 🌟.

_**Please if there are security issues, do not hesitate to contact me directly first. onigbindeayomide at gmail dott com**_
