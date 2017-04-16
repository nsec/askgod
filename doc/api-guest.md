# Introduction
The guest API is restricted to those in one of the guest subnets.

It allows a very limited set of queries, mostly focusing on rendering  
the current scoreboard.

# Error handling
Errors are returned as HTTP errors with custom error message.

The most frequently used ones are:
 - 200 on success
 - 400 for bad input (e.g. broken JSON)
 - 403 when accessing from a non-admin subnet
 - 404 for missing target
 - 500 for any server side error (DB failure, disk error, ...)

# /
## GET
This returns a list of valid API versions.

The response is a JSON encoded list of string.

# /1.0
## GET
This returns the current server status.

The response is a JSON encoded version of api.Status (see api/status.go).

# /1.0/events
## GET (?type=TYPE)
This is a websocket endpoint sending a stream of JSON encoded messages.  

The messages are made of an outer layer containing the nature of the  
message and a timestamp and an inner layer containing the actual message  
(format depending on type).

The outter layer is a JSON encoded version of api.Event (see api/event.go).

The inner layer is also JSON encoded but the struct depends on the type.

Multiple types can be passed as a comma separated list.

### "timeline" type
Inner layer is api.EventTimeline

This represents a change to the timeline (points granted or taken) and requires guest access.

### "logging" type
Inner layer is api.EventLogging

This represents a server log entry and requires admin access.

### "flags" type
Inner layer is api.EventFlag

This represents a detailed succesful flag submission and requires admin access.

# /1.0/scoreboard
## GET
This returns the current scoreboard.

The response is a JSON encoded version of a list of api.ScoreboardEntry (see api/scoreboard.go).

# /1.0/timeline
## GET
This returns a full timeline of all flag submissions.

The response is a JSON encoded version of a list of api.TimelineEntry (see api/timeline.go).
