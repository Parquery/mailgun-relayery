#!/usr/bin/env python3
"""
Send an email with mailgun-relayery. To be used for testing purposes.

WARNING: this script will use real MailGun credits to send the email!
"""
import argparse
import pathlib
import subprocess
import sys
import uuid
from typing import Any, Optional, List  # pylint: disable=unused-import

import temppathlib

import tests.component_test
import tests.control
import tests.database
import tests.proc
import tests.relay
import tests.siger


# pylint: disable=invalid-name, too-many-arguments, too-many-instance-attributes
def run(release_dir: pathlib.Path, database_dir: pathlib.Path, domain: str, api_key_path: pathlib.Path,
        recipients: List[str], cc: List[str], bcc: List[str], quiet: bool) -> None:
    """
    Test that mailgun-relayery works with a real MailGun API key and account.

    :param release_dir: path to the release directory
    :param database_dir: path to where LMDB files are stored
    :param domain: address of the MailGun API domain
    :param api_key_path: path to the MailGun API key
    :param recipients: addresses to which the test email should be sent
    :param cc: addresses for the cc field of the test email
    :param bcc: addresses for the bcc field of the test email
    :param quiet: if set, should produce as little output as possible
    :return:

    """
    # pylint: disable=too-many-locals, too-many-statements

    # create empty database
    database_dir.mkdir(exist_ok=True, parents=True)
    tests.database.initialize_environment(database_dir=database_dir)

    port_ctl = tests.component_test.find_free_port()

    # start the control server to store a channel in the empty database
    # yapf: disable
    cmd = [str(release_dir / 'bin' / 'mailgun-relay-controlery'),
           '-database_dir', database_dir.as_posix(),
           '-address', ':{}'.format(port_ctl)]
    # yapf: enable

    descriptor = "test-channel"
    token = str(uuid.uuid4())

    stdout = subprocess.DEVNULL if quiet else None

    proc_ctl = subprocess.Popen(cmd, stdout=stdout)
    with tests.proc.terminating(proc=proc_ctl, timeout=5):
        # let the server initialize
        tests.proc.sleep_while_process(proc=proc_ctl, seconds=1)

        if proc_ctl.poll() is not None:
            raise AssertionError("Expected the server process to be alive, but it died.")

        url = 'http://127.0.0.1:{}'.format(port_ctl)
        client_ctl = tests.control.RemoteCaller(url)

        # add the channel
        sender = tests.control.Entity(name="MailGun Relayery Test", email="test@mailgunrelayery.com")

        channel = tests.control.Channel(
            descriptor=descriptor,
            token=token,
            sender=sender,
            domain=domain,
            recipients=[
                tests.control.Entity(name="recipient-" + str(i + 1), email=addr) for i, addr in enumerate(recipients)
            ],
            cc=[tests.control.Entity(name="cc-" + str(i + 1), email=addr) for i, addr in enumerate(cc)],
            bcc=[tests.control.Entity(name="bcc-" + str(i + 1), email=addr) for i, addr in enumerate(bcc)],
            min_period=1000000,
            max_size=100000000)
        client_ctl.put_channel(channel=channel)

        # start the relay server
        port_relay = tests.component_test.find_free_port()

        # yapf: disable
        cmd = [str(release_dir / 'bin' / 'mailgun-relayery'),
               '-database_dir', database_dir.as_posix(),
               '-api_key_path', api_key_path.as_posix(),
               '-address', ':{}'.format(port_relay)]
        # yapf: enable

        proc = subprocess.Popen(cmd, stdout=stdout)
        with tests.proc.terminating(proc=proc, timeout=5):
            # let the server initialize
            tests.proc.sleep_while_process(proc=proc, seconds=1)

            if proc.poll() is not None:
                raise AssertionError("Expected the server process to be alive, but it died.")

            url_rel = 'http://127.0.0.1:{}'.format(port_relay)
            client_rel = tests.relay.RemoteCaller(url_rel)

            message = tests.relay.Message(
                subject="Test Message",
                content="This message was sent as part of a test in "
                "mailgun-relayery.Please ignore it."
                "\n\nSincerely,\nMailgun Test",
                html="<h1>This message</h1> <p>was sent as part of a <b>test in "
                "mailgun-relayery.</b></p>"
                "<p>Please ignore it.</p><p>Sincerely,</p><p>Mailgun "
                "Test</p>")

            client_rel.put_message(x_descriptor=descriptor, x_token=token, message=message)


class Params:
    """Contains the parameters of the example."""

    def __init__(self) -> None:
        """Initialize the example parameters."""
        self.release_dir = pathlib.Path()
        self.domain = ""
        self.api_key_path = pathlib.Path()
        self.recipients = []  # type: List[str]
        self.cc = []  # type: List[str]
        self.bcc = []  # type: List[str]
        self.database_dir = None  # type: Optional[pathlib.Path]
        self.quiet = False


def params_from_command_line(args: Any) -> Params:
    """
    Parse the parameters from the command line.

    :param args: the command line arguments
    :return: parsed parameters

    """
    assert isinstance(args.release_dir, str)
    assert isinstance(args.domain, str)
    assert isinstance(args.api_key_path, str)
    assert isinstance(args.recipients, str)

    if args.cc is not None:
        assert isinstance(args.cc, str)
    if args.bcc is not None:
        assert isinstance(args.bcc, str)
    if args.database_dir is not None:
        assert isinstance(args.database_dir, str)

    assert isinstance(args.quiet, bool)

    params = Params()
    params.release_dir = pathlib.Path(args.release_dir)
    params.api_key_path = pathlib.Path(args.api_key_path)
    params.recipients = args.recipients.split(",")
    params.domain = args.domain

    if args.database_dir is not None:
        params.database_dir = pathlib.Path(args.database_dir)
    if args.cc is not None:
        params.cc = args.cc.split(",")
    if args.bcc is not None:
        params.bcc = args.bcc.split(",")

    params.quiet = args.quiet

    return params


def main() -> int:
    """Execute the main routine."""
    parser = argparse.ArgumentParser()
    parser.description = "CAUTION: This test will use real MailGun credits to send one email to the addresses you " \
                         "provide!"
    parser.add_argument("-r", "--release_dir", help="path to the release directory", required=True)

    parser.add_argument("--domain", help="name of the MailGun domain", required=True)
    parser.add_argument("--api_key_path", help="path to the MailGun API key", required=True)
    parser.add_argument(
        "--recipients", help="email addresses of the email recipients, separated by commas", required=True)
    parser.add_argument("--cc", help="email addresses for the email's cc field, separated by commas")
    parser.add_argument("--bcc", help="email addresses for the email's bcc field, separated by commas")

    parser.add_argument(
        "--database_dir",
        help="path to the database directory. If specified, you have to delete it yourself. "
        "Otherwise, uses mkdtemp and removes it at the end.")
    parser.add_argument(
        "--quiet", help="if specified, produces as little log messages as possible", action="store_true")

    args = parser.parse_args()
    params = params_from_command_line(args)

    tests.siger.Siger.initialize_signal_handlers()

    with temppathlib.TmpDirIfNecessary(path=params.database_dir) as database_dir:
        run(release_dir=params.release_dir,
            database_dir=database_dir.path,
            domain=params.domain,
            api_key_path=params.api_key_path,
            recipients=params.recipients,
            cc=params.cc,
            bcc=params.bcc,
            quiet=params.quiet)

    return 0


if __name__ == "__main__":
    sys.exit(main())
