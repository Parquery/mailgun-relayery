"""Handle subprocesses."""
import subprocess
from typing import Optional

import time

import tests.siger


class terminating:  # pylint: disable=invalid-name
    """Terminate the process at the end of the block if the process has not finished already."""

    def __init__(self, proc: subprocess.Popen, timeout: Optional[int] = None) -> None:
        """
        Initialize the terminating context.

        :param proc: process to terminate
        :param timeout: in seconds; if time out is over, first we send SIGTERM, wait, then we send SIGKILL.
        """
        self.proc = proc
        self.timeout = timeout

    def __enter__(self) -> subprocess.Popen:
        return self.proc

    def __exit__(self, exc_type, exc_val, exc_tb):
        terminate_or_kill(proc=self.proc, timeout=self.timeout)


def sleep_while_process(proc: subprocess.Popen, seconds: float) -> None:
    """
    Sleep for the given number of seconds. If the process finished in the meanwhile, aborts the sleep.

    :param proc: process that needs to run while we are sleeping.
    :param seconds: number of seconds to sleep in total
    :return:

    """
    elapsed = 0.0
    while proc.poll() is None and elapsed < seconds:
        sleep_for(seconds=0.1)
        elapsed += 0.1


def terminate_or_kill(proc: subprocess.Popen, timeout: Optional[int] = None) -> None:
    """
    Terminate the given process. If the process did not terminate after the timeout, kills it.

    :param proc: process to terminate
    :param timeout: in seconds
    :return:

    """
    if proc.poll() is None:
        proc.terminate()
        try:
            proc.wait(timeout=timeout)
        except subprocess.TimeoutExpired:
            proc.kill()
            proc.communicate()


def sleep_for(seconds: float) -> None:
    """
    Sleep for the given period. Breaks if siger is done.

    :param seconds: to sleep; if 0 or negative, does not sleep at all.
    :return: None

    """
    if seconds <= 0.0:
        return

    rest = seconds
    while rest > 0.0:
        if rest > 1.0:
            sleep = 1.0
        else:
            sleep = rest

        rest -= sleep

        time.sleep(sleep)
        if tests.siger.Siger.done():
            break
