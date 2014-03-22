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


from askgod.config import config_get_bool
from askgod.db import db_commit, generic_add, generic_delete, generic_list, \
    generic_update, process_triggers, DBFlag, DBScore, DBTeam, DBTrigger
from askgod.decorators import admin_only, team_only
from askgod.log import monitor_add_client

from storm.locals import And, Reference, ReferenceSet

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
            raise KeyError("No such class name: %s" % classname)

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
        config['scores_read_only'] = config_get_bool("scores_read_only", False)
        config['scores_writeups'] = config_get_bool("scores_writeups", False)
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
            raise Exception("Team already has flag: %s" % flagid)

        score = DBScore()
        score.teamid = teamid
        score.flagid = flagid
        score.submit_time = datetime.datetime.now()

        if value is not None:
            score.value = value
        else:
            flags = db_store.find(DBFlag, id=flagid)
            if flags.count() != 1:
                raise IndexError("Couldn't find flagid=%s" % flagid)
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
            raise IndexError("Couldn't find scoreid=%s" % scoreid)
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
            result['value'] = entry.value
            result['submit_time'] = entry.submit_time
            if config_get_bool("server", "show_writeup", False):
                result['writeup_value'] = entry.writeup_value
            else:
                result['writeup_value'] = 0
            result['writeup_submit_time'] = entry.writeup_time
            if entry.flag.writeup_value:
                result['writeup_string'] = "WID%s" % entry.id
            else:
                result['writeup_string'] = ""
            result['return_string'] = entry.flag.return_string
            results.append(result)

        return sorted(results.values(), key=lambda result: result['flagid'])

    @admin_only
    def scores_list_timeline(self, client):
        """ Returns the timeline """
        db_store = client['db_store']

        result = []

        for score in db_store.find(DBScore).order_by(DBScore.submit_time):
            if not score.team.name:
                continue

            result.append({'teamid': score.teamid,
                           'submit_time': score.submit_time,
                           'value': score.value})

        return result

    def scores_scoreboard(self, client):
        """ Returns the scoreboard """
        db_store = client['db_store']

        teams = {}
        for team in db_store.find(DBTeam):
            if not team.name:
                continue

            teams[team.id] = {'teamid': team.id,
                              'team_name': team.name,
                              'team_country': team.country,
                              'team_website': team.website,
                              'score': 0,
                              'score_flags': 0,
                              'score_writeups': 0}

        for score in db_store.find(DBScore):
            if not score.teamid in teams:
                continue

            if score.value:
                teams[score.teamid]['score'] += score.value
                teams[score.teamid]['score_flags'] += score.value
            if (config_get_bool("server", "show_writeup", False)
                    and score.writeup_value):
                teams[score.teamid]['score'] += score.writeup_value
                teams[score.teamid]['score_writeups'] += score.writeup_value

        return sorted(teams.values(), key=lambda team: team['score'],
                      reverse=True)

    @team_only
    def scores_submit(self, client, flag):
        """ Submits/validates a flag """
        if config_get_bool("server", "scores_read_only", False):
            raise Exception("Server is read-only.")

        db_store = client['db_store']

        if not flag or (not isinstance(flag, str) and
                        not isinstance(flag, unicode)):
            if not flag:
                logging.debug("[team %02d] No flag provided" %
                              client['team'])
                raise Exception("No flag provided.")
            else:
                logging.debug("[team %02d] Invalid flag type: %s (%s)" %
                              (client['team'], flag, type(flag)))
                raise Exception("Invalid type for flag.")

        if isinstance(flag, str):
            flag = flag.decode('utf-8')

        logging.info("[team %02d] Submits flag: %s" % (client['team'], flag))

        results = db_store.find(DBFlag, DBFlag.flag.lower() == flag.lower())
        if results.count() == 0:
            logging.debug("[team %02d] Flag '%s' doesn't exist." %
                          (client['team'], flag))
            raise Exception("Flag isn't valid.")

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
                    raise Exception("Too late, the flag has been exhausted.")

            # Check that it wasn't already submitted
            if db_store.find(DBScore, flagid=entry.id,
                             teamid=client['team']).count() > 0:
                logging.debug("[team %02d] Flag '%s' was already submitted." %
                              (client['team'], flag))
                raise Exception("The flag has already been submitted.")

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
        raise Exception("Unknown error with your flag, please report this.")

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
                raise Exception("No code provided.")
            else:
                logging.debug("[team %02d] Invalid code type: %s (%s)" %
                              (client['team'], code, type(code)))
                raise Exception("Invalid type for code.")

        if isinstance(code, str):
            code = code.decode('utf-8')

        if not flag or (not isinstance(flag, str) and
                        not isinstance(flag, unicode)):
            if not flag:
                logging.debug("[team %02d] No flag provided" % client['team'])
                raise Exception("No flag provided.")
            else:
                logging.debug("[team %02d] Invalid flag type: %s (%s)" %
                              (client['team'], flag, type(flag)))
                raise Exception("Invalid type for flag.")

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
            raise Exception("Invalid code.")

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
                    raise Exception("Too late, the flag has been exhausted.")

            # Check that it wasn't already submitted
            if db_store.find(DBScore, flagid=entry.id,
                             teamid=client['team']).count() > 0:
                logging.debug("[team %02d] Flag '%s' was already submitted." %
                              (client['team'], code))
                raise Exception("The flag has already been submitted.")

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
        raise Exception("Unknown error with your flag, please report this.")

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

    @admin_only
    def teams_list(self, client):
        """ List all the teams in the database """
        return generic_list(client, DBTeam, sort=DBTeam.id)

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
