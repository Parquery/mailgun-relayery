#!/usr/bin/env python3
"""Run a test to verify that Mailgun Relayery check authentication data immediately upon receiving a request."""
import argparse
import pathlib
import socket
import subprocess
import sys
import time
import uuid

import logthis
import temppathlib

import tests.component_test
import tests.control
import tests.database
import tests.proc
import tests.relay
import tests.siger


def run_test(release_dir: pathlib.Path, operation_dir: pathlib.Path, quiet: bool) -> None:
    """
    Test that the relayry server checks the credentials before the request body.

    :param release_dir: path to the release directory
    :param operation_dir: path to where operation (temporary) files are stored
    :param quiet: if set, should produce as little output as possible
    :return:

    """
    # pylint: disable=too-many-locals, too-many-statements
    if not quiet:
        logthis.say("Starting the request size test for the relay server.")

    # create empty channel and timestamp databases
    database_dir = operation_dir / 'database'
    database_dir.mkdir(exist_ok=True, parents=True)
    tests.database.initialize_environment(database_dir=database_dir)

    # create and store a mock API key
    api_key = uuid.uuid4()
    api_key_pth = operation_dir / 'key'
    with api_key_pth.open(mode='w') as fid:
        fid.write(str(api_key))

    port_ctl = tests.component_test.find_free_port()

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
            domain="requests.test.com",
            min_period=0.001,
            max_size=100000000)
        client.put_channel(channel=channel)

    # start the relayery server
    port_relay = tests.component_test.find_free_port()

    # yapf: disable
    cmd = [str(release_dir / 'bin' / 'mailgun-relayery'),
           '-database_dir', database_dir.as_posix(),
           '-api_key_path', api_key_pth.as_posix(),
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
        wrong_token = token + "-wrong"

        # inspired by: https://stackoverflow.com/questions/48613006/python-sendall-not-raising-connection-closed-error
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
            sock.connect(('127.0.0.1', port_relay))
            request = 'POST {url} HTTP/1.1\r\nHost: 127.0.0.1:{port}\r\n' \
                      'Content-type: application/json\r\nAccept: application/json\r\n' \
                      'Content-length: 200\r\nConnection: close\r\n' \
                      'X-Descriptor: {desc}\r\nX-Token: {token}\r\n\r\n'.format(
                url="/api/message",
                port=port_relay,
                desc=desc, token=wrong_token)

            sock.sendall(request.encode("utf-8"))
            received = str(sock.recv(256), "utf-8")

            # as the raw response string contains variables like the timestamp of reception, we can't directly compare
            # it to a golden error string. Instead, we check that it contains the correct error status and message.
            assert "The request token for descriptor is invalid: some-channel-name.\n" in received
            assert "403 Forbidden" in received

            # a second message should be ignored
            time.sleep(0.5)
            sock.send(b"another msg")
            received = str(sock.recv(256), "utf-8")
            assert received == ""

            # a third message throws error
            time.sleep(0.5)
            broken_pipe = None  # Type: BrokenPipeError
            try:
                sock.send(b"a third, error-triggering message")
            except BrokenPipeError as pipe_err:
                broken_pipe = pipe_err

            assert broken_pipe is not None


def main() -> int:
    """Execute the main routine."""
    parser = argparse.ArgumentParser()

    parser.add_argument("-r", "--release_dir", help="path to the release directory", required=True)

    parser.add_argument(
        "--quiet", help="if specified, produces as little log messages as possible", action="store_true")

    args = parser.parse_args()
    assert isinstance(args.release_dir, str)
    assert isinstance(args.quiet, bool)

    release_dir = pathlib.Path(args.release_dir)

    tests.siger.Siger.initialize_signal_handlers()

    with temppathlib.TmpDirIfNecessary(path=None) as operation_dir:
        run_test(release_dir=release_dir, operation_dir=operation_dir.path, quiet=args.quiet)

    if not args.quiet:
        logthis.say("Test passed.")

    return 0


if __name__ == "__main__":
    sys.exit(main())
