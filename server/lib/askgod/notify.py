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

import json
import logging
import os
import socket
import subprocess


def notify_flag(teamid, code, value, tags):
    notify_servers = config_get_list("server", "notify_servers", [])
    notify_scripts = config_get_list("server", "notify_scripts", [])

    data = {'teamid': teamid,
            'code': code,
            'value': value,
            'tags': tags}
    json_data = json.dumps(data)

    old_timeout = socket.getdefaulttimeout()
    socket.setdefaulttimeout(3)
    for server in notify_servers:
        try:
            address = socket.getaddrinfo(server, 5000)[0][4]

            sock = socket.socket(socket.AF_INET6, socket.SOCK_STREAM)
            sock.connect(address)

            sock.sendall(json_data)

            sock.shutdown(socket.SHUT_RDWR)
            sock.close()
        except:
            logging.error("Unable to reach the notify server: %s" % server)

    socket.setdefaulttimeout(old_timeout)

    for script in notify_scripts:
        cmd = [script, str(teamid), code.encode(), str(value), tags.encode()]
        with open(os.devnull, "a") as devnull:
            ret = subprocess.call(cmd, stdout=devnull, stderr=devnull)

        if ret != 0:
            logging.error("Notify script '%s' returned non-zero: %s" %
                          (script, ret))
