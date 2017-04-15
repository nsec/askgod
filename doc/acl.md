# Access levels
Askgod implements the following access levels:
 - guest
 - team
 - admin

Access control is done based on the subnet of the requestor, subnets for  
each access level may be defined in askgod.yaml.

# Access structure
Users of a higher access level have automatic access to the lower levels.

So the member of a team subnet will have access to the guest API and an  
admin subnet member will have access to the team and guest API.

Note that for the team API endpoints there are further access  
restrictions caused by team authentication. One needs to both have the  
team ACL and be in a subnet attached to a valid team in the database to  
have access.
