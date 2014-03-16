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


def convert_properties(entry):
    for key, value in entry.items():
        if isinstance(value, str):
            entry[key] = value.decode('utf-8')


def validate_properties(table, entry):
    if not isinstance(entry, dict):
        raise TypeError("Expected a dict.")

    properties = [key for key in dir(table)
                  if not key.startswith("_")]

    for key in entry.keys():
        if key not in properties:
            raise IndexError("Invalid key: '%s'" % key)
