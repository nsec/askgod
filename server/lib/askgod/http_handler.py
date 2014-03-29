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

from askgod.exceptions import AskgodException
from storm.locals import Store

import SimpleXMLRPCServer
import SocketServer
import logging
import socket
import traceback
import xmlrpclib


class CustomXMLRPCServer(SocketServer.ThreadingMixIn,
                         SimpleXMLRPCServer.SimpleXMLRPCServer):
    address_family = socket.AF_INET6


class CustomRequestHandler(SimpleXMLRPCServer.SimpleXMLRPCRequestHandler):
    def _dispatch(self, method, params):
        if self.server.instance is None:
            logging.error("Client without a server instance!")
            raise AskgodException("Internal server error.")

        # call instance method directly
        func = None
        try:
            func = SimpleXMLRPCServer.resolve_dotted_attribute(
                self.server.instance,
                method,
                self.server.allow_dotted_names)
        except Exception as e:
            logging.info("Failed to resolv '%s': %s" % (method, e))
            raise AskgodException("Unable to resolve method name.")

        if not func:
            logging.info("Function '%s' doesn't exist." % func)
            raise AskgodException("Invalid method name.")

        # Per connection data (address, DB connection, request)
        client = {}
        client['client_address'] = self.client_address[0]
        client['db_store'] = Store(self.server.database)
        client['request'] = self.request

        # Actually call the function
        try:
            retval = func(client, *params)
        except not AskgodException:
            logging.error(traceback.format_exc())
            raise AskgodException("Internal server error.")

        # Attempt to close the DB connection (if still there)
        try:
            client['db_store'].commit()
            client['db_store'].close()
        except:
            pass

        return retval

    def do_OPTIONS(self):
        self.send_response(200)
        self.validate_origin()
        self.send_header("Cache-Control",
                         "no-store, no-cache, must-revalidate")
        self.end_headers()
        self.wfile.write("OK")

    def do_POST(self):
        """Handles the HTTP POST request.

        Attempts to interpret all HTTP POST requests as XML-RPC calls,
        which are forwarded to the server's _dispatch method for handling.
        """

        # Check that the path is legal
        if not self.is_rpc_path_valid():
            self.report_404()
            return

        try:
            # Get arguments by reading body of request.
            # We read this in chunks to avoid straining
            # socket.read(); around the 10 or 15Mb mark, some platforms
            # begin to have problems (bug #792570).
            max_chunk_size = 10 * 1024 * 1024
            size_remaining = int(self.headers["content-length"])
            L = []
            while size_remaining:
                chunk_size = min(size_remaining, max_chunk_size)
                chunk = self.rfile.read(chunk_size)
                if not chunk:
                    break
                L.append(chunk)
                size_remaining -= len(L[-1])
            data = ''.join(L)

            data = self.decode_request_content(data)
            if data is None:
                return  # response has been sent

            # In previous versions of SimpleXMLRPCServer, _dispatch
            # could be overridden in this class, instead of in
            # SimpleXMLRPCDispatcher. To maintain backwards compatibility,
            # check to see if a subclass implements _dispatch and dispatch
            # using that method if present.
            response = self.server._marshaled_dispatch(
                data, getattr(self, '_dispatch', None), self.path)
        except Exception, e:  # This should only happen if the module is buggy
            # internal error, report as HTTP server error
            self.send_response(500)

            # Send information about the exception if requested
            if hasattr(self.server, '_send_traceback_header') and \
                    self.server._send_traceback_header:
                self.send_header("X-exception", str(e))
                self.send_header("X-traceback", traceback.format_exc())

            self.send_header("Content-length", "0")
            self.end_headers()
        else:
            # got a valid XML RPC response
            self.send_response(200)
            self.validate_origin()
            self.send_header("Cache-Control",
                             "no-store, no-cache, must-revalidate")
            self.send_header("Content-type", "text/xml")
            if self.encode_threshold is not None:
                if len(response) > self.encode_threshold:
                    q = self.accept_encodings().get("gzip", 0)
                    if q:
                        try:
                            response = xmlrpclib.gzip_encode(response)
                            self.send_header("Content-Encoding", "gzip")
                        except NotImplementedError:
                            pass
            self.send_header("Content-length", str(len(response)))
            self.end_headers()
            self.wfile.write(response)

    def validate_origin(request):
        """
            Receives an HTTP request, checks it against the allowed HTTP
            origins, if allowed, sets the appropriate headers.
        """

        request.send_header("Access-Control-Allow-Origin", "http://www.nsec")
        request.send_header("Access-Control-Allow-Headers", "Content-Type")
