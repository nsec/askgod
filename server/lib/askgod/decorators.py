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

from askgod.config import config_get_list
from askgod.db import generic_list, DBTeam

import ipaddr
import logging


def admin_only(fn):
    """
        Decorator used to check that the query comes from the admin network.
    """

    def wrapped(*args, **kwargs):
        client = args[1]
        admin = False
        for net in config_get_list("server", "admin_net"):
            if ipaddr.IPAddress(client['client_address']) \
                    in ipaddr.IPNetwork(net):
                admin = True
                break

        if not admin:
            logging.info("Unauthorized admin request from '%s'" %
                         client['client_address'])
            return False

        return fn(*args, **kwargs)
    return wrapped


def team_only(fn):
    """
        Decorator used to check that the query comes from a valid team network.
    """

    def wrapped(*args, **kwargs):
        client = args[1]

        teamlist = generic_list(client, DBTeam,
                                ('id', 'subnets'),
                                DBTeam.id)

        for team in teamlist:
            subnets = team['subnets']
            if not subnets:
                continue

            if "," in subnets:
                subnets = [subnet.strip()
                           for subnet in team['subnets'].split(",")]

            for subnet in subnets:
                if ipaddr.IPAddress(client['client_address']) \
                        in ipaddr.IPNetwork(subnet):
                    client['team'] = team['id']
                    return fn(*args, **kwargs)

        logging.error("Team not found for '%s'" % client['client_address'])
        return False

    return wrapped
