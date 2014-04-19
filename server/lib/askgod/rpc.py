#!/usr/bin/python
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


from askgod.config import config_get_bool, config_get_list
from askgod.db import db_commit, convert_properties, generic_add, \
    generic_delete, generic_list, generic_update, process_triggers, \
    validate_properties, DBFlag, DBScore, DBTeam, DBTrigger
from askgod.decorators import admin_only, team_only, team_or_guest
from askgod.exceptions import AskgodException
from askgod.log import monitor_add_client
from askgod.notify import notify_flag

from storm.locals import And, Reference, ReferenceSet

from operator import itemgetter

import datetime
import logging
import subprocess


class AskGod:
    """
        The exported XMLRPC functions.
    """

    @admin_only
    def class_properties(self, client, classname):
        classes = {'flag': DBFlag,
                   'score': DBScore,
                   'team': DBTeam,
                   'trigger': DBTrigger}

        if not classname in classes:
            raise AskgodException("No such class name: %s" % classname)

        fields = []
        for field in dir(classes[classname]):
            if field.startswith("_"):
                continue

            # We don't recurse over the API, so skip references
            classfield = getattr(classes[classname], field)
            if (isinstance(classfield, Reference) or
                    isinstance(classfield, ReferenceSet)):
                continue

            fields.append(field)

        return sorted(fields)

    def config_variables(self, client):
        config = {}
        config['scores_hide_others'] = config_get_bool("server",
                                                       "scores_hide_others",
                                                       False)
        config['scores_progress_overall'] = config_get_bool(
            "server", "scores_progress_overall", False)
        config['scores_read_only'] = config_get_bool("server",
                                                     "scores_read_only", False)
        config['scores_writeups'] = config_get_bool("server",
                                                    "scores_writeups", False)
        config['teams_setdetails'] = config_get_bool("server",
                                                     "teams_setdetails", False)
        return config

    @admin_only
    def monitor(self, client, loglevel=20):
        client['db_store'].close()
        monitor_add_client(client['request'], loglevel)

    # Flags
    @admin_only
    def flags_add(self, client, entry):
        """ Adds a flag in the database"""
        return generic_add(client, DBFlag, entry)

    @admin_only
    def flags_delete(self, client, entryid):
        """ Removes a flag from the database """
        return generic_delete(client, DBFlag, entryid)

    @admin_only
    def flags_list(self, client):
        """ List all the flags in the database """
        return generic_list(client, DBFlag, sort=DBFlag.id)

    @admin_only
    def flags_update(self, client, entryid, entry):
        """ Update an existing flag in the database """

        return generic_update(client, DBFlag, entryid, entry)

    # Scores
    @admin_only
    def scores_add(self, client, entry):
        """ Adds a score in the database"""
        return generic_add(client, DBScore, entry)

    @admin_only
    def scores_delete(self, client, entryid):
        """ Removes a score from the database """
        return generic_delete(client, DBScore, entryid)

    @admin_only
    def scores_grant_flag(self, client, teamid, flagid, value):
        """ Gives the given flag to the team """
        db_store = client['db_store']

        # Check that it wasn't already submitted
        if db_store.find(DBScore, flagid=flagid,
                         teamid=teamid).count() > 0:
            raise AskgodException("Team already has flag: %s" % flagid)

        score = DBScore()
        score.teamid = teamid
        score.flagid = flagid
        score.submit_time = datetime.datetime.now()

        if value is not None:
            score.value = value
        else:
            flags = db_store.find(DBFlag, id=flagid)
            if flags.count() != 1:
                raise AskgodException("Couldn't find flagid=%s" % flagid)
            score.value = flags[0].value

        db_store.add(score)
        db_commit(db_store)

        logging.info("[team %02d] Scores %s points with flagid=%s (admin)" %
                     (score.teamid, score.value, score.flagid))

        client['team'] = teamid
        process_triggers(client)
        return True

    @admin_only
    def scores_grant_writeup(self, client, scoreid, value):
        """ Gives the writeup points to the team """
        db_store = client['db_store']

        scores = db_store.find(DBScore, id=scoreid)
        if scores.count() != 1:
            raise AskgodException("Couldn't find scoreid=%s" % scoreid)
        score = scores[0]

        score.writeup_time = datetime.datetime.now()

        if value is not None:
            score.writeup_value = value
        else:
            score.writeup_value = score.flag.writeup_value

        db_commit(db_store)

        logging.info("[team %02d] Scores %s points with a "
                     "writeup for flagid=%s (admin)" %
                     (score.teamid, score.writeup_value, score.flagid))

        return True

    @admin_only
    def scores_list(self, client):
        """ List all the scores in the database """
        return generic_list(client, DBScore)

    @team_only
    def scores_list_submitted(self, client):
        """ Lists all the flags the team found and any related hints """
        db_store = client['db_store']

        results = []

        for entry in db_store.find(DBScore, teamid=client['team']):
            result = {}
            result['flagid'] = entry.flagid
            result['description'] = (entry.flag.description
                                     if entry.flag.description else "")
            result['value'] = entry.value
            result['submit_time'] = entry.submit_time
            if config_get_bool("server", "scores_writeups", False):
                result['writeup_value'] = entry.writeup_value
            else:
                result['writeup_value'] = 0
            result['writeup_submit_time'] = entry.writeup_time
            if entry.flag.writeup_value:
                result['writeup_string'] = "WID%s" % entry.id
            else:
                result['writeup_string'] = ""
            result['return_string'] = (entry.flag.return_string
                                       if entry.flag.return_string else "")
            results.append(result)

        return sorted(results, key=lambda result: result['flagid'])

    @team_only
    def scores_progress(self, client, tags=None):
        """ Returns the progress percentage """
        db_store = client['db_store']

        if not tags:
            # Overall progress
            if not config_get_bool("server", "scores_progress_overall", False):
                raise AskgodException("Overall progress is disabled.")

            total = 0.0
            obtained = 0.0
            for entry in db_store.find(DBFlag):
                if entry.teamid and entry.teamid != client['team']:
                    continue

                total += entry.value

            for entry in db_store.find(DBScore, teamid=client['team']):
                obtained += entry.value

            return int(obtained / total * 100)

        ret = {}
        if not isinstance(tags, list):
            ret = 0.0
            tags = [tags]

        for tag in tags:
            namespace = tag.split(":")[0]
            if namespace not in config_get_list("server",
                                                "scores_progress_tags",
                                                []):
                raise AskgodException("Disallowed tag namespaced.")

            total = 0.0
            obtained = 0.0

            for entry in db_store.find(DBFlag):
                if entry.teamid and entry.teamid != client['team']:
                    continue
                entry_tags = entry.tags.split(",")

                if tag in entry_tags:
                    total += entry.value

            for entry in db_store.find(DBScore, teamid=client['team']):
                entry_tags = entry.flag.tags.split(",")
                if tag in entry_tags:
                    obtained += entry.value

            if isinstance(ret, dict):
                if not total:
                    ret[tag] = 0
                    continue
                ret[tag] = int(obtained / total * 100)
            else:
                if not total:
                    return 0
                return int(obtained / total * 100)

        return ret

    @team_or_guest
    def scores_scoreboard(self, client):
        """ Returns the scoreboard """
        db_store = client['db_store']

        hide_others = config_get_bool("server", "scores_hide_others", False)

        teams = {}
        for team in db_store.find(DBTeam):
            # Skip teams without a name (likely unconfigured)
            if not team.name:
                continue

            teams[team.id] = {'teamid': team.id,
                              'team_name': (team.name
                                            if team.name else ""),
                              'team_country': (team.country
                                               if team.country else ""),
                              'team_website': (team.website
                                               if team.website else ""),
                              'score': 0,
                              'score_flags': 0,
                              'score_writeups': 0}

        # The public scoreboard only shows 0 when hide_others is set
        if hide_others and 'team' not in client:
            return sorted(teams.values(), key=lambda team: team['teamid'])

        for score in db_store.find(DBScore):
            # Skip score entries without a matching team
            if not score.teamid in teams:
                continue

            # Skip teams other than the requestor when hide_others is set
            if hide_others and score.teamid != client['team']:
                continue

            if score.value:
                teams[score.teamid]['score'] += score.value
                teams[score.teamid]['score_flags'] += score.value
            if (config_get_bool("server", "scores_writeups", False)
                    and score.writeup_value):
                teams[score.teamid]['score'] += score.writeup_value
                teams[score.teamid]['score_writeups'] += score.writeup_value

        return sorted(teams.values(), key=lambda team: team['score'],
                      reverse=True)

    @team_only
    def scores_submit(self, client, flag):
        """ Submits/validates a flag """
        if config_get_bool("server", "scores_read_only", False):
            raise AskgodException("Server is read-only.")

        db_store = client['db_store']

        if not flag or (not isinstance(flag, str) and
                        not isinstance(flag, unicode)):
            if not flag:
                logging.debug("[team %02d] No flag provided" %
                              client['team'])
                raise AskgodException("No flag provided.")
            else:
                logging.debug("[team %02d] Invalid flag type: %s (%s)" %
                              (client['team'], flag, type(flag)))
                raise AskgodException("Invalid type for flag.")

        if isinstance(flag, str):
            flag = flag.decode('utf-8')

        logging.info("[team %02d] Submits flag: %s" % (client['team'], flag))

        results = db_store.find(DBFlag, DBFlag.flag.lower() == flag.lower())
        if results.count() == 0:
            logging.debug("[team %02d] Flag '%s' doesn't exist." %
                          (client['team'], flag))
            notify_flag(client['team'], "", 0, "")
            raise AskgodException("Flag isn't valid.")

        for entry in results:
            # Deal with per-team flags
            if (entry.teamid is not None and
                    entry.teamid != client['team']):
                continue

            # Deal with counter-limited flags
            if entry.counter:
                count = db_store.find(DBScore, flagid=entry.id).count()
                if count >= entry.counter:
                    logging.debug("[team %02d] Flag '%s' has been exhausted." %
                                  (client['team'], flag))
                    raise AskgodException("Too late, the flag has "
                                          "been exhausted.")

            # Check that it wasn't already submitted
            if db_store.find(DBScore, flagid=entry.id,
                             teamid=client['team']).count() > 0:
                logging.debug("[team %02d] Flag '%s' was already submitted." %
                              (client['team'], flag))
                raise AskgodException("The flag has already been submitted.")

            # Add to score
            score = DBScore()
            score.teamid = client['team']
            score.flagid = entry.id
            score.value = entry.value
            score.submit_time = datetime.datetime.now()

            db_store.add(score)
            db_commit(db_store)

            logging.info("[team %02d] Scores %s points with flagid=%s" %
                         (client['team'], entry.value, entry.id))
            notify_flag(client['team'], entry.code, entry.value, entry.tags)

            retval = []

            # Generate response
            response = {}
            response['value'] = score.value
            if entry.return_string:
                response['return_string'] = entry.return_string

            if entry.writeup_value:
                response['writeup_string'] = "WID%s" % score.id

            retval.append(response)

            # Process triggers
            retval += process_triggers(client)

            return retval

        logging.debug("[team %02d] Flag '%s' exists but can't be used." %
                      (client['team'], flag))
        raise AskgodException("Unknown error with your flag, "
                              "please report this.")

    @team_only
    def scores_submit_special(self, client, code, flag):
        """ Submits/validates a special flag (external validator) """
        if config_get_bool("server", "scores_read_only", False):
            return -6

        db_store = client['db_store']

        if not code or (not isinstance(code, str) and
                        not isinstance(code, unicode)):
            if not code:
                logging.debug("[team %02d] No code provided" % client['team'])
                raise AskgodException("No code provided.")
            else:
                logging.debug("[team %02d] Invalid code type: %s (%s)" %
                              (client['team'], code, type(code)))
                raise AskgodException("Invalid type for code.")

        if isinstance(code, str):
            code = code.decode('utf-8')

        if not flag or (not isinstance(flag, str) and
                        not isinstance(flag, unicode)):
            if not flag:
                logging.debug("[team %02d] No flag provided" % client['team'])
                raise AskgodException("No flag provided.")
            else:
                logging.debug("[team %02d] Invalid flag type: %s (%s)" %
                              (client['team'], flag, type(flag)))
                raise AskgodException("Invalid type for flag.")

        if isinstance(code, str):
            flag = flag.decode('utf-8')

        logging.info("[team %02d] Submits special flag for code: %s => %s" %
                     (client['team'], code, flag))

        # NOTE: Intentional, "DBFlag.flag == None" != "DBFlag.flag is not None"
        results = db_store.find(DBFlag, And(DBFlag.code == code,
                                            DBFlag.flag == None,
                                            DBFlag.validator != None))
        if results.count() == 0:
            logging.debug("[team %02d] Code '%s' doesn't exist." %
                          (client['team'], code))
            notify_flag(client['team'], "", 0, "")
            raise AskgodException("Invalid code.")

        for entry in results:
            # Deal with per-team flags
            if entry.teamid and entry.teamid != client['team']:
                continue

            # Deal with counter-limited flags
            if entry.counter:
                count = db_store.find(DBScore, flagid=entry.id).count()
                if count >= entry.counter:
                    logging.debug("[team %02d] Flag '%s' has been exhausted." %
                                  (client['team'], code))
                    raise AskgodException("Too late, the flag has "
                                          "been exhausted.")

            # Check that it wasn't already submitted
            if db_store.find(DBScore, flagid=entry.id,
                             teamid=client['team']).count() > 0:
                logging.debug("[team %02d] Flag '%s' was already submitted." %
                              (client['team'], code))
                raise AskgodException("The flag has already been submitted.")

            # Call validator
            if subprocess.call(["validator/%s" % entry.validator,
                                str(client['team']),
                                str(code),
                                str(flag)]) != 0:
                continue

            # Add to score
            score = DBScore()
            score.teamid = client['team']
            score.flagid = entry.id
            score.value = entry.value
            score.submit_time = datetime.datetime.now()

            db_store.add(score)
            db_commit(db_store)

            logging.info("[team %02d] Scores %s points with flagid=%s" %
                         (client['team'], entry.value, entry.id))
            notify_flag(client['team'], entry.code, entry.value, entry.tags)

            retval = []

            # Generate response
            response = {}
            response['value'] = score.value
            if entry.return_string:
                response['return_string'] = entry.return_string

            if entry.writeup_value:
                response['writeupid'] = "WID%s" % score.id

            retval.append(response)

            # Process triggers
            retval += process_triggers(client)

            return retval

        logging.debug("[team %02d] Flag '%s' exists but won't validate." %
                      (client['team'], flag))
        raise AskgodException("Unknown error with your flag, "
                              "please report this.")

    @team_or_guest
    def scores_timeline(self, client):
        """ Returns the timeline """
        db_store = client['db_store']

        hide_others = config_get_bool("server", "scores_hide_others", False)

        result = []

        # Guests don't get to see anything when hide_others is set
        if hide_others and "team" not in client:
            return result

        for score in db_store.find(DBScore).order_by(DBScore.submit_time):
            if not score.team.name:
                continue

            # Skip teams other than the requestor when hide_others is set
            if hide_others and score.teamid != client['team']:
                continue

            result.append({'teamid': score.teamid,
                           'submit_time': score.submit_time,
                           'value': score.value})

        return sorted(result, key=itemgetter('teamid', 'submit_time'))

    @admin_only
    def scores_update(self, client, entryid, entry):
        """ Update an existing score in the database """

        return generic_update(client, DBScore, entryid, entry)

    # Teams
    @admin_only
    def teams_add(self, client, entry):
        """ Adds a team in the database"""
        return generic_add(client, DBTeam, entry)

    @admin_only
    def teams_delete(self, client, entryid):
        """ Removes a team from the database """
        return generic_delete(client, DBTeam, entryid)

    @team_only
    def teams_getdetails(self, client):
        """ Return the details for the caller's team """

        db_store = client['db_store']

        results = db_store.find(DBTeam, id=client['team'])
        if results.count() != 1:
            raise AskgodException("Can't find a match for id=%s" %
                                  client['team'])

        dbentry = results[0]

        details = {'id': dbentry.id,
                   'name': dbentry.name if dbentry.name else "",
                   'country': dbentry.country if dbentry.country else "",
                   'website': dbentry.website if dbentry.website else ""}

        return details

    @admin_only
    def teams_list(self, client):
        """ List all the teams in the database """
        return generic_list(client, DBTeam, sort=DBTeam.id)

    @team_only
    def teams_setdetails(self, client, fields):
        """ Set the details for the caller's team """

        if not config_get_bool("server", "teams_setdetails", False):
            raise AskgodException("Setting team details isn't allowed.")

        db_store = client['db_store']

        convert_properties(fields)
        validate_properties(DBTeam, fields)

        results = db_store.find(DBTeam, id=client['team'])
        if results.count() != 1:
            raise AskgodException("Can't find a match for id=%s"
                                  % client['team'])

        dbentry = results[0]

        if set(fields.keys()) - set(['name', 'country', 'website']):
            raise AskgodException("Invalid field provided.")

        if "country" in fields:
            if len(fields['country']) != 2 or not fields['country'].isalpha():
                raise AskgodException("Invalid country ISO code.")
            fields['country'] = fields['country'].upper()

        for key, value in fields.items():
            if not value:
                continue

            if getattr(dbentry, key):
                raise AskgodException("Field is already set: %s" % key)

            setattr(dbentry, key, value)

        return db_commit(db_store)

    @admin_only
    def teams_update(self, client, entryid, entry):
        """ Update an existing team in the database """

        return generic_update(client, DBTeam, entryid, entry)

    # Triggers
    @admin_only
    def triggers_add(self, client, entry):
        """ Adds a trigger in the database"""
        return generic_add(client, DBTrigger, entry)

    @admin_only
    def triggers_delete(self, client, entryid):
        """ Removes a trigger from the database """
        return generic_delete(client, DBTrigger, entryid)

    @admin_only
    def triggers_list(self, client):
        """ List all the triggers in the database """
        return generic_list(client, DBTrigger)

    @admin_only
    def triggers_update(self, client, entryid, entry):
        """ Update an existing trigger in the database """

        return generic_update(client, DBTrigger, entryid, entry)
