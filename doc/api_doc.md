# Public functions
Public functions may be called by anyone over the XML-RPC interface.
Those will usually be restricted to very high level aggregate of the
data stored by askgod and may be cached along the way.

## scores\_scoreboard()
Returns a list of dict each containing:
 - teamid (int): ID of the team
 - score (int): Total score
 - score\_flags (int): Points made from flags
 - score\_writeups (int): Point made from writeups

The list is sorted by score in descending order.

If the scoreboard is disabled, all scores will be 0.
If the writeups are disabled, score\_writeups will be 0.

## teams\_list()
Returns a list of dict each containing:
 - id (int): ID of the team
 - name (string): Name of the team (may contain unicode)
 - country (string): Two letters ISO code of the country
 - website (string): URL of the team's website

The list is sorted by id in ascending order.

# Team functions
## scores\_list\_submitted
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
## scores\_submit\_special(code, flag)

# Admin functions
## class\_fields
## monitor
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
## teams\_list\_admin
## teams\_update
## triggers\_add
## triggers\_delete
## triggers\_list
## triggers\_update
