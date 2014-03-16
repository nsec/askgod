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

from ConfigParser import ConfigParser

config = {}


def config_get(section, key, default=KeyError):
    """
        Look for the provided key in the config, return its value as a
        string if found or the default value if not. If the default is an
        exception, raise it.
    """

    if config and section in config and key in config[section]:
        return config[section][key]

    if isinstance(default, Exception):
        raise default()

    return str(default)


def config_get_bool(section, key, default=KeyError):
    """
        Similar to config_get but returns a boolean.
    """

    value = config_get(section, key, default)

    if value.lower() in ("yes", "true", "1", "on"):
        return True

    return False


def config_get_int(section, key, default=KeyError):
    """
        Similar to config_get but returns an integer.
    """

    value = config_get(section, key, default)

    if value.isdigit():
        return int(value)

    return value


def config_get_list(section, key, default=KeyError):
    """
        Similar to config_get_list but returns a list.
    """

    value = config_get(section, key, default)

    if ", " in value:
        return value.split(", ")

    return value


def config_load(path):
    """
        Generate a dictionary from the provided config and load it in
        memory.  Returns a reference to it.
    """

    configp = ConfigParser()
    try:
        configp.read(path)
    except:
        return config

    for section in configp.sections():
        config_section = {}
        for option in configp.options(section):
            value = configp.get(section, option)
            config_section[option] = value.strip('"')
        config[section] = config_section

    return config
