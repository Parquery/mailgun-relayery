#!/usr/bin/env python3
"""Run a component test of Mailgun Relayery."""
import argparse
import contextlib
import http
import http.server
import json
import pathlib
import re
import socket
import subprocess
import sys
import threading
import uuid
from typing import Optional, Any, List  # pylint: disable=unused-import

import logthis
import requests
import temppathlib

import tests.control
import tests.database
import tests.proc
import tests.relay
import tests.siger

CORRECT_REQUESTS = []
WRONG_REQUESTS = []


class MockServerRequestHandler(http.server.BaseHTTPRequestHandler):
    """Mocks the Mailgun remote server."""

    MSG_PATTERN = re.compile(r'/messages')
    quiet = False

    # pylint: disable=redefined-builtin
    def log_message(self, format, *args):
        """Override the default logging fucntion."""
        if not self.quiet:
            sys.stderr.write("[MailGun Mock server] %s - - [%s] %s\n" % (self.address_string(),
                                                                         self.log_date_time_string(), format % args))

    # pylint: disable=invalid-name
    def do_POST(self):
        """Handle POST requests to the mock server."""
        if re.search(self.MSG_PATTERN, self.path):
            resp_json = {"message": "Queued. Thank you.", "id": "<20111114174239.25659.5817@samples.mailgun.org>"}
            # pylint: disable=no-member
            self.send_response(message="message successfully relayed.", code=requests.codes.ok)

            self.send_header('Content-Type', 'application/json; charset=utf-8')
            self.end_headers()

            response_content = json.dumps(resp_json)
            self.wfile.write(response_content.encode('utf-8'))

            CORRECT_REQUESTS.append(self.path)
        else:
            self.send_response(message="bad request: {}".format(self.path), code=400)
            WRONG_REQUESTS.append(self.path)


def start_mock_server(port: int, quiet: bool):
    """
    Start a mock MailGun server in a Daemon thread, listening to the target port.

    :param port: the port for the server to listen to
    :param quiet: if set, the server produces no output to STDOUT
    :return:
    """
    CORRECT_REQUESTS.clear()
    WRONG_REQUESTS.clear()

    handler = MockServerRequestHandler
    handler.quiet = quiet
    mock_server = http.server.HTTPServer(('localhost', port), handler)

    mock_server_thread = threading.Thread(target=mock_server.serve_forever)
    mock_server_thread.setDaemon(True)
    mock_server_thread.start()


def run_test_control(release_dir: pathlib.Path, operation_dir: pathlib.Path, quiet: bool) -> None:
    """
    Test that the mailgun relayery control server works correctly.

    :param release_dir: path to the release directory
    :param operation_dir: path to where operation (temporary) files are stored
    :param quiet: if set, should produce as little output as possible
    :return:

    """
    # pylint: disable=too-many-locals, too-many-statements
    if not quiet:
        logthis.say("Starting the component test for the control server.")

    # create empty database
    database_dir = operation_dir / 'database'
    database_dir.mkdir(exist_ok=True, parents=True)
    tests.database.initialize_environment(database_dir=database_dir)

    port = find_free_port()

    # yapf: disable
    cmd = [str(release_dir / 'bin' / 'mailgun-relay-controlery'),
           '-database_dir', database_dir.as_posix(),
           '-address', ':{}'.format(port)]
    # yapf: enable
    if quiet:
        cmd += '-quiet'

    stdout = subprocess.DEVNULL if quiet else None

    proc = subprocess.Popen(cmd, stdout=stdout)
    with tests.proc.terminating(proc=proc, timeout=5):
        # let the server initialize
        tests.proc.sleep_while_process(proc=proc, seconds=1)

        if proc.poll() is not None:
            raise AssertionError("Expected the server process to be alive, but it died.")

        url = 'http://127.0.0.1:{}'.format(port)
        client = tests.control.RemoteCaller(url)

        # get channels listing through the ctl server
        pages = client.list_channels()
        assert isinstance(pages, tests.control.ChannelsPage)
        expected = tests.control.ChannelsPage(page=1, per_page=100, page_count=0, channels=[])
        assert pages.to_jsonable() == expected.to_jsonable(), \
            "Expected empty page listing ({}), got {}.".format(expected.to_jsonable(), pages.to_jsonable())

        # add a channel and token
        desc = "some-channel-name"
        token = "leWq221234123423oiweoWPEOFIWPEOFKPDKlsdkepwsodPOR"
        sender = tests.control.Entity(email="someone@some-domain.com")
        recipients = [
            tests.control.Entity(name="client", email="client@another-domain.com"),
            tests.control.Entity(name="devop", email="devop@some-domain.com")
        ]
        cc_field = [tests.control.Entity(name="client-2", email="client-2@another-domain.com")]
        bcc_field = [tests.control.Entity(name="devop-2", email="devop-2@some-domain.com")]
        channel = tests.control.Channel(
            descriptor=desc,
            token=token,
            sender=sender,
            recipients=recipients,
            cc=cc_field,
            bcc=bcc_field,
            domain="component.test.com",
            min_period=0.001,
            max_size=1000000)
        client.put_channel(channel=channel)

        # get channels listing to check successful insertion
        pages = client.list_channels()
        assert isinstance(pages, tests.control.ChannelsPage)
        expected = tests.control.ChannelsPage(page=1, per_page=100, page_count=1, channels=[channel])
        assert pages.to_jsonable() == expected.to_jsonable(), \
            "Expected empty page listing ({}), got {}.".format(expected.to_jsonable(), pages.to_jsonable())

        # remove channel
        client.delete_channel(descriptor=desc)

        # get channels listing to check successful deletion
        pages = client.list_channels()
        assert isinstance(pages, tests.control.ChannelsPage)
        expected = tests.control.ChannelsPage(page=1, per_page=100, page_count=0, channels=[])
        assert pages.to_jsonable() == expected.to_jsonable(), \
            "Expected empty page listing ({}), got {}.".format(expected.to_jsonable(), pages.to_jsonable())

        # put and overwrite channel
        client.put_channel(channel=channel)
        client.put_channel(channel=channel)
        client.put_channel(channel=channel)

        # get channels listing to check successful overwrite
        pages = client.list_channels()
        assert isinstance(pages, tests.control.ChannelsPage)
        expected = tests.control.ChannelsPage(page=1, per_page=100, page_count=1, channels=[channel])
        assert pages.to_jsonable() == expected.to_jsonable(), \
            "Expected empty page listing ({}), got {}.".format(expected.to_jsonable(), pages.to_jsonable())

        # delete non-existing channel
        resp = client.delete_channel(descriptor=desc + "-suffix")
        expected_resp = b'No channel associated to the descriptor some-channel-name-suffix was found.'
        assert resp == expected_resp, "expected {}, got {}".format(resp, expected_resp)


def run_test_relay(release_dir: pathlib.Path, operation_dir: pathlib.Path, quiet: bool) -> None:
    """
    Test that relayry server works correctly.

    :param release_dir: path to the release directory
    :param operation_dir: path to where operation (temporary) files are stored
    :param quiet: if set, should produce as little output as possible
    :return:

    """
    # pylint: disable=too-many-locals, too-many-statements
    if not quiet:
        logthis.say("Starting the component test for the relay server.")

    # create empty channel and timestamp databases
    database_dir = operation_dir / 'database'
    database_dir.mkdir(exist_ok=True, parents=True)
    tests.database.initialize_environment(database_dir=database_dir)

    # create and store a mock API key
    api_key = uuid.uuid4()
    api_key_pth = operation_dir / 'key'
    with api_key_pth.open(mode='w') as fid:
        fid.write(str(api_key))

    port_ctl = find_free_port()

    # start the control server to store a channel in the empty database
    # yapf: disable
    cmd = [str(release_dir / 'bin' / 'mailgun-relay-controlery'),
           '-database_dir', database_dir.as_posix(),
           '-address', ':{}'.format(port_ctl)]
    # yapf: enable
    if quiet:
        cmd += '-quiet'

    desc = "some-channel-name"
    token = "leWq221234123423oiweoWPEOFIWPEOFKPDKlsdkepwsodPOR"

    stdout = subprocess.DEVNULL if quiet else None

    proc = subprocess.Popen(cmd, stdout=stdout)
    with tests.proc.terminating(proc=proc, timeout=5):
        # let the server initialize
        tests.proc.sleep_while_process(proc=proc, seconds=1)

        if proc.poll() is not None:
            raise AssertionError("Expected the server process to be alive, but it died.")

        url = 'http://127.0.0.1:{}'.format(port_ctl)
        client = tests.control.RemoteCaller(url)

        # add a channel
        sender = tests.control.Entity(email="someone@some-domain.com")
        recipients = [
            tests.control.Entity(name="client", email="client@another-domain.com"),
            tests.control.Entity(name="devop", email="devop@some-domain.com")
        ]
        cc_field = [tests.control.Entity(name="client-2", email="client-2@another-domain.com")]
        bcc_field = [tests.control.Entity(name="devop-2", email="devop-2@some-domain.com")]
        channel = tests.control.Channel(
            descriptor=desc,
            token=token,
            sender=sender,
            recipients=recipients,
            cc=cc_field,
            bcc=bcc_field,
            domain="component.test.com",
            min_period=0.001,
            max_size=1000000)
        client.put_channel(channel=channel)

    # start and test the relayery server

    port_relay = find_free_port()
    port_mailgun = find_free_port()

    start_mock_server(port=port_mailgun, quiet=quiet)
    if not quiet:
        logthis.say("Mock MailGun server listening on :{}.".format(port_mailgun))

    # yapf: disable
    cmd = [str(release_dir / 'bin' / 'mailgun-relayery'),
           '-database_dir', database_dir.as_posix(),
           '-api_key_path', api_key_pth.as_posix(),
           '-mailgun_address', 'http://127.0.0.1:{}'.format(port_mailgun),
           '-address', ':{}'.format(port_relay)]
    # yapf: enable
    if quiet:
        cmd += '-quiet'

    proc = subprocess.Popen(cmd, stdout=stdout)
    with tests.proc.terminating(proc=proc, timeout=5):
        # let the server initialize
        tests.proc.sleep_while_process(proc=proc, seconds=1)

        if proc.poll() is not None:
            raise AssertionError("Expected the server process to be alive, but it died.")

        url_rel = 'http://127.0.0.1:{}'.format(port_relay)
        client_rel = tests.relay.RemoteCaller(url_rel)

        # relay a message to the mock MailGun server
        message = tests.relay.Message(subject="a message from your friend", content="dear friend, I hope all is good.")
        resp = client_rel.put_message(x_descriptor=desc, x_token=token, message=message)
        expected = b'The message has been correctly relayed.'
        assert resp == expected, "expected {}, got {}".format(expected, resp)
        assert len(CORRECT_REQUESTS) == 1
        assert len(WRONG_REQUESTS) == 0

        req_pth = "/component.test.com/messages"
        assert CORRECT_REQUESTS[0] == req_pth, "expected {}, got {}".format(req_pth, CORRECT_REQUESTS[0])

        # error 403: invalid token
        wrong_token = token + "_suffix"
        http_err = None  # type: Optional[requests.exceptions.HTTPError]
        try:
            _ = client_rel.put_message(x_descriptor=desc, x_token=wrong_token, message=message)
        except requests.exceptions.HTTPError as err:
            http_err = err

        expected_err = "403 Client Error: Forbidden for url: {}/api/message".format(url_rel)
        assert http_err.__str__() == expected_err, "expected {}, got {}".format(expected_err, http_err)

        # error 404: non-existing descriptor
        wrong_desc = desc + "_suffix"
        http_err = None
        try:
            _ = client_rel.put_message(x_descriptor=wrong_desc, x_token=token, message=message)
        except requests.exceptions.HTTPError as err:
            http_err = err

        expected_err = "404 Client Error: Not Found for url: {}/api/message".format(url_rel)
        assert http_err.__str__() == expected_err, "expected {}, got {}".format(expected_err, http_err)


def run_test_relay_errors(release_dir: pathlib.Path, operation_dir: pathlib.Path, quiet: bool) -> None:
    """
    Test that relayry server blocks wrong requests.

    :param release_dir: path to the release directory
    :param operation_dir: path to where operation (temporary) files are stored
    :param quiet: if set, should produce as little output as possible
    :return:

    """
    # pylint: disable=too-many-locals, too-many-statements
    if not quiet:
        logthis.say("Starting the component test for the relay server errors.")

    # create empty database
    database_dir = operation_dir / 'database'
    database_dir.mkdir(exist_ok=True, parents=True)
    tests.database.initialize_environment(database_dir=database_dir)

    # create and store a mock API key
    api_key = uuid.uuid4()
    api_key_pth = operation_dir / 'key'
    with api_key_pth.open(mode='w') as fid:
        fid.write(str(api_key))

    port_ctl = find_free_port()

    # start the control server to store a channel in the empty database
    # yapf: disable
    cmd = [str(release_dir / 'bin' / 'mailgun-relay-controlery'),
           '-database_dir', database_dir.as_posix(),
           '-address', ':{}'.format(port_ctl)]
    # yapf: enable
    if quiet:
        cmd += '-quiet'

    desc_large_min_period = "some-channel-large-period"
    desc_small_max_size = "some-channel-small-size"
    token = "leWq221234123423oiweoWPEOFIWPEOFKPDKlsdkepwsodPOR"

    stdout = subprocess.DEVNULL if quiet else None

    proc_ctl = subprocess.Popen(cmd, stdout=stdout)
    with tests.proc.terminating(proc=proc_ctl, timeout=5):
        # let the server initialize
        tests.proc.sleep_while_process(proc=proc_ctl, seconds=1)

        if proc_ctl.poll() is not None:
            raise AssertionError("Expected the server process to be alive, but it died.")

        url = 'http://127.0.0.1:{}'.format(port_ctl)
        client_ctl = tests.control.RemoteCaller(url)

        # add two channels: one with large minimum period, the other with small max size
        sender = tests.control.Entity(email="someone@some-domain.com")
        recipients = [
            tests.control.Entity(name="client", email="client@another-domain.com"),
            tests.control.Entity(name="devop", email="devop@some-domain.com")
        ]
        cc_field = [tests.control.Entity(name="client-2", email="client-2@another-domain.com")]
        bcc_field = [tests.control.Entity(name="devop-2", email="devop-2@some-domain.com")]
        channel = tests.control.Channel(
            descriptor=desc_large_min_period,
            token=token,
            sender=sender,
            recipients=recipients,
            cc=cc_field,
            bcc=bcc_field,
            domain="component.test.com",
            min_period=10,
            max_size=1000000)
        client_ctl.put_channel(channel=channel)

        channel = tests.control.Channel(
            descriptor=desc_small_max_size,
            token=token,
            sender=sender,
            recipients=recipients,
            cc=cc_field,
            bcc=bcc_field,
            domain="component.test.com",
            min_period=0.0001,
            max_size=1)
        client_ctl.put_channel(channel=channel)

        # start the relayery server
        port_relay = find_free_port()
        port_mailgun = find_free_port()

        start_mock_server(port=port_mailgun, quiet=quiet)
        if not quiet:
            logthis.say("Mock MailGun server listening on :{}.".format(port_mailgun))

        # yapf: disable
        cmd = [str(release_dir / 'bin' / 'mailgun-relayery'),
               '-database_dir', database_dir.as_posix(),
               '-api_key_path', api_key_pth.as_posix(),
               '-mailgun_address', 'http://127.0.0.1:{}'.format(port_mailgun),
               '-address', ':{}'.format(port_relay)]
        # yapf: enable
        if quiet:
            cmd += '-quiet'

        proc = subprocess.Popen(cmd, stdout=stdout)
        with tests.proc.terminating(proc=proc, timeout=5):
            # let the server initialize
            tests.proc.sleep_while_process(proc=proc, seconds=1)

            if proc.poll() is not None:
                raise AssertionError("Expected the server process to be alive, but it died.")

            url_rel = 'http://127.0.0.1:{}'.format(port_relay)
            client_rel = tests.relay.RemoteCaller(url_rel)

            message = tests.relay.Message(
                subject="a message from your friend", content="dear friend, I hope all is good.")

            # error 429: too many requests
            http_err = None  # type: Optional[requests.exceptions.HTTPError]
            try:
                _ = client_rel.put_message(x_descriptor=desc_large_min_period, x_token=token, message=message)
                _ = client_rel.put_message(x_descriptor=desc_large_min_period, x_token=token, message=message)
            except requests.exceptions.HTTPError as err:
                http_err = err

            expected_err = "429 Client Error: Too Many Requests for url: {}/api/message".format(url_rel)
            assert http_err.__str__() == expected_err, "expected {}, got {}".format(expected_err, http_err)

            # error 413: request entity too large
            http_err = None
            try:
                _ = client_rel.put_message(x_descriptor=desc_small_max_size, x_token=token, message=message)
            except requests.exceptions.HTTPError as err:
                http_err = err

            expected_err = "413 Client Error: Request Entity Too Large for url: {}/api/message".format(url_rel)
            assert http_err.__str__() == expected_err, "expected {}, got {}".format(expected_err, http_err)

            # overwrite a channel with a different (still very large) min_period
            sender = tests.control.Entity(email="someone@some-domain.com")
            recipients = [
                tests.control.Entity(name="client", email="client@another-domain.com"),
                tests.control.Entity(name="devop", email="devop@some-domain.com")
            ]
            cc_field = [tests.control.Entity(name="client-2", email="client-2@another-domain.com")]
            bcc_field = [tests.control.Entity(name="devop-2", email="devop-2@some-domain.com")]

            channel = tests.control.Channel(
                descriptor=desc_large_min_period,
                token=token,
                sender=sender,
                recipients=recipients,
                cc=cc_field,
                bcc=bcc_field,
                domain="component.test.com",
                min_period=50,
                max_size=1000000)
            client_ctl.put_channel(channel=channel)

            # verify that the server erased the last seen timestamp and allows us to relay a message
            message = tests.relay.Message(
                subject="a message from your friend", content="dear friend, I hope all is good.")
            resp = client_rel.put_message(x_descriptor=desc_large_min_period, x_token=token, message=message)
            expected = b'The message has been correctly relayed.'
            assert resp == expected, "expected {}, got {}".format(expected, resp)
            assert len(CORRECT_REQUESTS) == 2
            assert len(WRONG_REQUESTS) == 0


def find_free_port() -> int:
    """
    Find a next free port on the machine. Mind that this is not multi-process safe and can lead to race conditions.

    :return: a free port as a number

    """
    skt = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    with contextlib.closing(skt):
        skt.bind(('', 0))
        _, port = skt.getsockname()
        return port


class Params:
    """Parameter of the component test."""

    def __init__(self) -> None:
        """Initialize with default values."""
        self.release_dir = pathlib.Path()
        self.operation_dir = None  # type: Optional[pathlib.Path]
        self.quiet = False


def params_from_command_line(args: Any) -> Params:
    """
    Parse the parameters from the command line.

    :param args: the command line arguments
    :return: parsed parameters

    """
    assert isinstance(args.release_dir, str)

    if args.operation_dir is not None:
        assert isinstance(args.operation_dir, str)

    assert isinstance(args.quiet, bool)

    params = Params()
    params.release_dir = pathlib.Path(args.release_dir)

    if args.operation_dir is not None:
        params.operation_dir = pathlib.Path(args.operation_dir)

    params.quiet = args.quiet

    return params


def main() -> int:
    """Execute the main routine."""
    parser = argparse.ArgumentParser()

    parser.add_argument("-r", "--release_dir", help="path to the release directory", required=True)

    parser.add_argument(
        "--operation_dir",
        help="path to the operation directory. If specified, you have to delete it yourself. "
        "Otherwise, uses mkdtemp and removes it at the end.")
    parser.add_argument(
        "--quiet", help="if specified, produces as little log messages as possible", action="store_true")

    args = parser.parse_args()
    params = params_from_command_line(args)

    tests.siger.Siger.initialize_signal_handlers()

    with temppathlib.TmpDirIfNecessary(path=params.operation_dir) as operation_dir:
        test_dir = operation_dir.path / 'test_{}'.format(uuid.uuid4())
        test_dir.mkdir()

        run_test_control(release_dir=params.release_dir, operation_dir=test_dir / 'test_control', quiet=params.quiet)

        run_test_relay(
            release_dir=params.release_dir, operation_dir=operation_dir.path / 'test_relay', quiet=params.quiet)

        run_test_relay_errors(
            release_dir=params.release_dir, operation_dir=operation_dir.path / 'test_relay_errors', quiet=params.quiet)

    if not params.quiet:
        logthis.say("Test passed.")

    return 0


if __name__ == "__main__":
    sys.exit(main())
