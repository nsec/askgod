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

import logging
import sys
import time

# Global list of connected monitor clients
monitors = []


class MonitorHandler(logging.Handler):
    def __init__(self):
        logging.Handler.__init__(self)

    def emit(self, record):
        for monitor, level in monitors:
            if level > record.levelno:
                continue

            try:
                monitor.send(self.format(record).encode("utf-8"))
                monitor.sendall("\n")
            except:
                monitor.close()
                try:
                    monitors.remove(monitor)
                except:
                    pass


def monitor_add_client(request, loglevel):
    """
        Add the HTTP request to the monitor pool and start keeping it
        alive.
    """

    monitors.append((request, loglevel))
    request.setblocking(True)

    try:
        while 1:
            request.recv(1024)
            time.sleep(1)
    except:
        return False


def log_config(path=None, level=None):
    """
        Setup our custom logging.
        Logs to both standard log files and to any attached monitor
        clients.
    """

    if level is None:
        level = logging.INFO
    else:
        level = int(level)

    formatter = logging.Formatter(
        "%(asctime)s %(levelname)s %(message)s")

    stdoutlogger = logging.StreamHandler(sys.stdout)
    stdoutlogger.setFormatter(formatter)
    stdoutlogger.setLevel(level)
    logging.root.addHandler(stdoutlogger)

    monitorlogger = MonitorHandler()
    monitorlogger.setFormatter(formatter)
    logging.root.addHandler(monitorlogger)

    if path:
        filelogger = logging.FileHandler(path)
        filelogger.setFormatter(formatter)
        filelogger.setLevel(level)
        logging.root.addHandler(filelogger)

    logging.root.setLevel(1)
