# Introduction
The admin API is restricted to those in one of the admin subnets.

It allows near direct DB interaction with flags, scores and teams.  
Additionaly it also allows for server monitoring and config auditing.

Note that being in an admin subnet also gives access to extra messages  
in the /1.0/scoreboard/events API (part of the guest API).

# Error handling
Errors are returned as HTTP errors with custom error message.

The most frequently used ones are:
 - 200 on success
 - 400 for bad input (e.g. broken JSON)
 - 403 when accessing from a non-admin subnet
 - 404 for missing target
 - 500 for any server side error (DB failure, disk error, ...)

Unlike the guest and team APIs, the admin endpoints will usually return  
server side errors unfiltered.

# /1.0/config
## GET
This returns the current Askgod configuration with a few sensitive fields masked.

The response is a JSON encoded version of api.Config (see api/config.go).

# /1.0/flags
## GET
This returns all the flags from the database.

The response is a JSON encoded version of a list of api.AdminFlag (see api/flag.go).

## POST
This is used to create a new flag entry in the database.

The input is a JSON encoded version of api.AdminFlagPost (see api/flag.go).

There is no expected output for this endpoint.

## DELETE
This is used to clear all flag entries from the database.

There is no expected input for this endpoint.

There is no expected output for this endpoint.

An http parameter of ?empty=1 is required to prevent accidents.

# /1.0/flags/{id}
## GET
This returns a single flag record from the database.

The response is a JSON encoded version of api.AdminFlag (see api/flag.go).

## PUT
This updates an existing flag record in the database.

The input is a JSON encoded version of api.AdminFlagPut (see api/flag.go).

There is no expected output for this endpoint.

## DELETE
This deletes an existing flag record in the database.

There is no expected input for this endpoint.

There is no expected output for this endpoint.

# /1.0/scores
## GET
This returns all the scores from the database.

The response is a JSON encoded version of a list of api.AdminScore (see api/score.go).

## POST
This is used to create a new score entry in the database.

The input is a JSON encoded version of api.AdminScorePost (see api/score.go).

There is no expected output for this endpoint.

## DELETE
This is used to clear all score entries from the database.

There is no expected input for this endpoint.

There is no expected output for this endpoint.

An http parameter of ?empty=1 is required to prevent accidents.

# /1.0/score/{id}
## GET
This returns a single score record from the database.

The response is a JSON encoded version of api.AdminScore (see api/score.go).

## PUT
This updates an existing score record in the database.

The input is a JSON encoded version of api.AdminScorePut (see api/score.go).

There is no expected output for this endpoint.

## DELETE
This deletes an existing score record in the database.

There is no expected input for this endpoint.

There is no expected output for this endpoint.

# /1.0/teams
## GET
This returns all the teams from the database.

The response is a JSON encoded version of a list of api.AdminTeam (see api/team.go).

## POST
This is used to create a new team entry in the database.

The input is a JSON encoded version of api.AdminTeamPost (see api/team.go).

There is no expected output for this endpoint.

## DELETE
This is used to clear all team entries from the database.

There is no expected input for this endpoint.

There is no expected output for this endpoint.

An http parameter of ?empty=1 is required to prevent accidents.

# /1.0/teams/{id}
## GET
This returns a single team record from the database.

The response is a JSON encoded version of api.AdminTeam (see api/team.go).

## PUT
This updates an existing team record in the database.

The input is a JSON encoded version of api.AdminTeamPut (see api/team.go).

There is no expected output for this endpoint.

## DELETE
This deletes an existing team record in the database.

There is no expected input for this endpoint.

There is no expected output for this endpoint.
