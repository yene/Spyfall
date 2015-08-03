# Spyfall
The Card Game Spyfall implemented with images. Backend is written in go as exercise.

## Run
I am using fswatch to restart the server when the code changes.
```
brew install fswatch
fswatch-run *.go ./server.sh 
```

## Todo
* Go online: put on my server
* Free memory: remove old rooms based on the timestamp
* Add different themes (maybe archer, james bond, original)
* Error Messages: notify user what went wrong, maybe little info box
* Mobile: figure should have 100% vh/vw, the coundown and status should overlap, no whitespace
* Test concurrency: write a http test and hammer into the server, i think there is a bug
* split code into multiple files
* move settings into cmd arguments

## Maybe Features
* let the user set the countdown time
* understand the template functions better
* add roles
* notify when player leaves lobby or game

## TIL
* backdrop filter does not seem to work with figure: http://iamvdo.me/en/blog/advanced-css-filters
