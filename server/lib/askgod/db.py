# -*- coding: utf-8 -*-
# Copyright 2013-2014 - Stéphane Graber <stgraber@nsec.io>

# This program is free software; you can redistribute it and/or modify
# it under the terms of the GNU General Public License version 2, as
# published by the Free Software Foundation.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License along
# with this program; if not, write to the Free Software Foundation, Inc.,
# 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

from askgod.utils import convert_properties, validate_properties
from askgod.exceptions import AskgodException

from storm.locals import DateTime, Int, Max, Reference, ReferenceSet
from storm.locals import Select, Store, Unicode
import datetime
import glob
import logging
import traceback


class DBFlag(object):
    """
        Storm class for table 'flag'.
    """

    __storm_table__ = "flag"
    __storm_primary__ = "id"
    id = Int()
    triggerid = Int()  # Foreign key to trigger
    teamid = Int()  # Foreign key to team
    code = Unicode()  # Codename
    flag = Unicode()  # The actual flag
    value = Int()  # Number of points to be awarded
    writeup_value = Int()  # Number of points to be awarde on write-up
    return_string = Unicode()  # String to return on success
    counter = Int()  # Number of times the flag can be submitted
    validator = Unicode()  # Path to validation script (for submit_special)
    description = Unicode()  # Description of the flag
    tags = Unicode()  # Comma separated list of tags (<ns>:<tag>)


class DBSchema(object):
    """
        Storm class for table 'schema'
    """

    __storm_table__ = "schema"
    __storm_primary__ = "version"
    version = Int()  # Schema version number


class DBScore(object):
    """
        Storm class for table 'score'
    """

    __storm_table__ = "score"
    __storm_primary__ = "id"
    id = Int()
    teamid = Int()  # Foreign key to team
    flagid = Int()  # Foreign key to flag
    value = Int()  # Number of points awarded
    writeup_value = Int()  # Number of writeup points awarded
    submit_time = DateTime()  # Time at which the flag was submitted
    writeup_time = DateTime()  # Time at which the writeup was submitted


class DBTeam(object):
    """
        Storm class for table 'team'
    """

    __storm_table__ = "team"
    __storm_primary__ = "id"
    id = Int()
    name = Unicode()  # Team name
    country = Unicode()  # Country (2-letter ISO)
    website = Unicode()  # Team website
    subnets = Unicode()  # Comma separated list of subnets
    notes = Unicode()  # Notes for admins


class DBTrigger(object):
    """
        Storm class for table 'trigger'
    """

    __storm_table__ = "trigger"
    __storm_primary__ = "id"
    id = Int()
    flagid = Int()  # Foreign key to flag
    count = Unicode()  # Number of points or percentage required amongst all
                       # flags linked to the trigger
    description = Unicode()  # Description of the trigger


# DB references
## Relationships for 'flag'
DBFlag.scores = ReferenceSet(DBFlag.id, DBScore.flagid)
DBFlag.team = Reference(DBFlag.teamid, DBTeam.id)
DBFlag.trigger = Reference(DBFlag.triggerid, DBTrigger.id)
DBFlag.triggers = ReferenceSet(DBFlag.id, DBTrigger.flagid)

## Relationships for 'score'
DBScore.flag = Reference(DBScore.flagid, DBFlag.id)
DBScore.team = Reference(DBScore.teamid, DBTeam.id)

## Relationships for 'team'
DBTeam.flags = ReferenceSet(DBTeam.id, DBFlag.teamid)
DBTeam.scores = ReferenceSet(DBTeam.id, DBScore.teamid)

## Relationships for 'trigger'
DBTrigger.flag = Reference(DBTrigger.flagid, DBFlag.id)
DBTrigger.flags = ReferenceSet(DBTrigger.id, DBFlag.triggerid)


def db_commit(db_store):
    """
        Attempt to commit and flush to the store.
        On failure, rollback and flush the store.
    """

    try:
        db_store.commit()
        db_store.flush()
        return True
    except:
        logging.critical(traceback.format_exc())
        db_store.rollback()
        db_store.flush()
        return False


def db_update_schema(database):
    """
        Check for pending database schema updates.
        If any are found, apply them and bump the version.
    """

    # Connect to the database
    db_store = Store(database)

    # Check if the DB schema has been loaded
    db_exists = False
    try:
        db_store.execute(Select(DBSchema.version))
        db_exists = True
    except:
        db_store.rollback()
        logging.debug("Failed to query schema table.")

    if not db_exists:
        logging.info("Creating database")
        schema_file = sorted(glob.glob("schema/schema-*.sql"))[-1]
        schema_version = schema_file.split(".")[0].split("-")[-1]
        logging.debug("Using '%s' to deploy schema '%s'" % (schema_file,
                                                            schema_version))
        with open(schema_file, "r") as fd:
            try:
                for line in fd.read().replace("\n", "").split(";"):
                    if not line:
                        continue
                    db_store.execute("%s;" % line)
                    db_commit(db_store)
                logging.info("Database created")
            except:
                logging.critical("Failed to initialize the database")
                return False

    # Get schema version
    version = db_store.execute(Select(Max(DBSchema.version))).get_one()[0]
    if not version:
        logging.critical("No schema version.")
        return False

    # Apply updates
    for update_file in sorted(glob.glob("schema/update-*.sql")):
        update_version = update_file.split(".")[0].split("-")[-1]
        if int(update_version) > version:
            logging.info("Using '%s' to deploy update '%s'" % (update_file,
                                                               update_version))
            with open(update_file, "r") as fd:
                try:
                    for line in fd.read().replace("\n", "").split(";"):
                        if not line:
                            continue
                        db_store.execute("%s;" % line)
                        db_commit(db_store)
                except:
                    logging.critical("Failed to load schema update")
                    return False

    # Get schema version
    new_version = db_store.execute(Select(Max(DBSchema.version))).get_one()[0]
    if new_version > version:
        logging.info("Database schema successfuly updated from '%s' to '%s'" %
                     (version, new_version))

    db_store.close()


def generic_add(client, dbclass, entry):
    db_store = client['db_store']

    if 'id' in entry:
        raise AskgodException("You can't pass the 'id' field in an entry.")

    convert_properties(entry)
    validate_properties(dbclass, entry)

    dbentry = dbclass()
    for key, value in entry.items():
        setattr(dbentry, key, value)

    db_store.add(dbentry)
    return db_commit(db_store)


def generic_delete(client, dbclass, entryid):
    db_store = client['db_store']

    if not isinstance(entryid, int):
        raise AskgodException("The ID must be an integer.")

    results = db_store.find(dbclass, id=entryid)
    if results.count() != 1:
        raise AskgodException("Can't find a match for id=%s" % entryid)

    dbentry = results[0]

    db_store.remove(dbentry)
    return db_commit(db_store)


def generic_list(client, table, fields=None, sort=None):
    db_store = client['db_store']
    result = []
    if sort:
        entries = db_store.find(table).order_by(sort)
    else:
        entries = db_store.find(table)

    for entry in entries:
        record = {}
        if fields:
            for field in fields:
                record[field] = getattr(entry, field)
        else:
            for field in dir(entry):
                if field.startswith("_"):
                    continue

                # We don't recurse over the API, so skip references
                classfield = getattr(table, field)
                if (isinstance(classfield, Reference) or
                        isinstance(classfield, ReferenceSet)):
                    continue

                record[field] = getattr(entry, field)
        result.append(record)

    return result


def generic_update(client, dbclass, entryid, entry):
    db_store = client['db_store']

    if not isinstance(entryid, int):
        raise AskgodException("The ID must be an integer.")

    if 'id' in entry:
        raise AskgodException("You can't pass the 'id' field in an entry.")

    results = db_store.find(dbclass, id=entryid)
    if results.count() != 1:
        raise AskgodException("Can't find a match for id=%s" % entryid)

    convert_properties(entry)
    validate_properties(dbclass, entry)

    flag = results[0]
    for key, value in entry.items():
        setattr(flag, key, value)

    return db_commit(db_store)


def process_triggers(client):
    db_store = client['db_store']
    team = client['team']

    retval = []
    team_flags = [entry.flagid
                  for entry in db_store.find(DBScore, teamid=team)]

    for entry in db_store.find(DBTrigger):
        # We already have that one, moving on
        if entry.flagid in team_flags:
            continue

        # Figure out how many points are needed
        flags = []
        total = 0
        for sub_entry in entry.flags:
            total += sub_entry.value
            flags.append(sub_entry.id)

        if entry.count.isdigit():
            count = int(entry.count)
        elif '%' in entry.count:
            count = (total / 100.0 * int(entry.count.replace('%', '')))
        else:
            logging.error("Invalid 'count' field for trigger: %s" % entry.id)

        # Check how many points we have
        team_count = 0
        for sub_entry in db_store.find(DBScore,
                                       (DBScore.teamid == team) &
                                       DBScore.flagid.is_in(flags)):
            team_count += sub_entry.value

        if team_count < count:
            continue

        # Add to score
        score = DBScore()
        score.teamid = team
        score.flagid = entry.flag.id
        score.value = entry.flag.value
        score.submit_time = datetime.datetime.now()

        db_store.add(score)
        db_commit(db_store)

        logging.info("[team %02d] Scores %s points with flagid=%s (trigger)" %
                     (team, entry.flag.value, entry.flag.id))

        # Generate response
        response = {}
        response['value'] = score.value
        response['trigger'] = True
        if entry.flag.return_string:
            response['return_string'] = entry.flag.return_string

        if entry.flag.writeup_value:
            response['writeup_string'] = "WID%s" % score.id
        retval.append(response)
    return retval
