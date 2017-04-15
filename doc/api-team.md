# Introduction
The team API is restricted to those in one of the team subnets.

It allows configuring team information and the submission of flags.

# Error handling
Errors are returned as HTTP errors with custom error message.

The most frequently used ones are:
 - 200 on success
 - 400 for bad input (e.g. broken JSON)
 - 403 when accessing from a non-admin subnet
 - 404 for missing target
 - 500 for any server side error (DB failure, disk error, ...)

# /1.0/team
## GET
This returns the current team information.

The response is a JSON encoded version of api.Team (see api/team.go).

## PUT
This updates the team information.

Note that whether this is allowed depends on askgod configuration.  
If the team\_self\_register flag (see GET /1.0) is set to true, then this is available.  
If the team\_self\_update flag is set to false, then only an initial change will be allowed.

The input is a JSON encoded version of api.TeamPut (see api/team.go).

# /1.0/team/flags
## GET
This returns a list of all valid flags the team submitted

The response is a JSON encoded version of a list of api.Flag (see api/flag.go).

## POST
This submits a flag to askgod.

The input is a JSON encoded version of api.FlagPost (see api/flag.go).

On success, the response is a JSON encoded version of api.Flag (see api/flag.go).

# /1.0/team/flags/{id}
## GET
This returns a single flag entry.

The response is a JSON encoded version of api.Flag (see api/flag.go).

## PUT
This updates a flag entry.

The input is a JSON encoded version of api.FlagPut (see api/flag.go).
