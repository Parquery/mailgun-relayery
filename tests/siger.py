#!/usr/bin/env python3
"""Provide singleton for managing SIGINT and SIGTERM signals so that programs can exit gracioussly."""

import signal

import logthis


class Siger:
    """Handle termination and interrupt signals."""

    __done = False

    @staticmethod
    def initialize_signal_handlers() -> None:
        """Initialize the signal handlers for signals SIGINT and SIGTERM."""
        signal.signal(signal.SIGINT, Siger.signal_handler)
        signal.signal(signal.SIGTERM, Siger.signal_handler)

    @staticmethod
    def signal_handler(signalnum: int, frame: object) -> None:
        """Handle stop signals."""
        # pylint: disable=unused-argument
        # we need signalnum and frame in order to pass this handler to the system.

        signalstr = signal.Signals(signalnum).name  # pylint: disable=no-member
        logthis.say("Received signal: {} {}".format(signalnum, signalstr))
        Siger.__done = True

    @staticmethod
    def done() -> bool:
        """
        Indicate when one of the handled signals was received.

        :return: True on signal

        """
        return Siger.__done
