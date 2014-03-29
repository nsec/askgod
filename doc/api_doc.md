# Public functions
Public functions may be called by anyone over the XML-RPC interface.
Those will usually be restricted to very high level aggregate of the
data stored by askgod and may be cached along the way.

## config\_variables()
Returns a dict representing the server configuration, current keys include:
 - scores\_hide\_others (bool): Whether other team scores are hidden
 - scores\_overall\_progress (bool): Whether it's possible for teams to
   get their overall progress (percentage of the sum of the points of all
   achievable flags).
 - scores\_read\_only (bool): Whether the server accepts flag submission
 - scores\_writeups (bool): Whether writeups are enabled
 - teams\_setdetails (bool): Whether teams can fill in their own information

## scores\_progress(tags)
This function may be called in 3 different ways:
 - Without arguments, in which case it'll return the overall progress of
   the team if allowed by the server.
 - With a single tag as an argument (string) in which case, if the tag
   is valid and allowed for team listing, the progress percentage will be
   returned as a single integer.
 - With a list of tags in which case, a dict will be returned with the
   key being the tag string as provided and the value being an integer
   representing the percentage of progress for the given tag.

On failure, exceptions will be raised. Note that if any provided flag is
invalid, an exception will be raised, you will not get a partial dict in
that case.

## scores\_scoreboard()
Returns a list of dict each containing:
 - teamid (int): ID of the team
 - team\_name (string): Name of the team (may contain unicode)
 - team\_country (string): Two letters ISO code of the country
 - team\_website (string): URL of the team's website
 - score (int): Total score
 - score\_flags (int): Points made from flags
 - score\_writeups (int): Point made from writeups

The list is sorted by score in descending order.

If the scoreboard is disabled, all scores will be 0 except for the team
of the requestor.
If the writeups are disabled, score\_writeups will be 0.

## teams\_getdetails()
Returns a dict containing information related to the caller's team:
 - id (int): The team's ID
 - name (string): The team's name
 - country (string): Two letters ISO code
 - website (string): The team's website

The strings may be empty if unset.

## teams\_setdetails(properties)
One time setup of an unconfigured team.
This takes a dict of the various fields that need to be set and returns
True on success or raises the appropriate exception on failure.

The valid keys are:
 - name (string): Name of the team
 - country (string): Two capital letters (country ISO code)
 - website (string): URL of the team's website

Only fields that aren't already set may be configured this way, changing
an existing value requires admin privileges.

If this feature is disabled, an exception will be raised.
Support for the feature may be checked through config\_variables().

# Team functions
## scores\_list\_submitted()
Returns a list of dict each containing:
 - flagid (int): ID of the flag
 - value (int): Number of points earned
 - submit\_time (string): Submission time for the flag
 - return\_string (string): Message shown when the flag was sent
 - writeup\_value (int): Number of points earned for the writeup
 - writeup\_submit\_time (string): Submission time for the writeup
 - writeup\_string (string): Writeup identifier (WID + score entry ID)

The list is sorted by flagid in ascending order.

If the writeups are disabled, writeup\_string will be empty,
writeup\_time will be empty and writeup\_value will be 0.

## scores\_submit(flag)
Submits a flag (as a string).

On success, a list of dict is returned, each of those dicts can contain
any of those fields:
 - return\_string (string): String linked with the flag
 - trigger (bool): Set to true if the entry is the result of a trigger
 - value (int): Number of points scored
 - writeup\_string (string): ID to be used for writeup submission

All errors are returned as an XML-RPC exception with the exception
string set accordingly.

## scores\_submit\_special(code, flag)
Submits a special flag (as a string), identified by its code (also a string).

The behavior and return values is identical to scores\_submit above.

# Admin functions
Those functions are only accessible from admin subnets listed in the
server's configuration.

## class\_properties(class)
Returns a list of string representing all the properties of the provided
class.

The list is sorted alphabetically in ascending order.

All errors are returned as an XML-RPC exception with the exception
string set accordingly.

## monitor(loglevel=20)
This function is a bit of a hack. It won't return a valid XML-RPC
response but will instead keep the connection alive indefinitely and
hook it up to the logging module. This will basically mirror the
server's console to the client.

Loglevel must be an integer representing a valid logging level.
The default value is 20 (INFO).

## flags\_add
## flags\_delete
## flags\_list
## flags\_update
## scores\_add
## scores\_delete
## scores\_grant\_flag
## scores\_grant\_writeup
## scores\_list
## scores\_list\_timeline
## scores\_update
## teams\_add
## teams\_delete
## teams\_list
## teams\_update
## triggers\_add
## triggers\_delete
## triggers\_list
## triggers\_update
