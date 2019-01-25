#!/usr/bin/env python3
"""Initialize the mailgun-relayery channel database."""
import pathlib

import icontract
import lmdb

# Descriptor -> Channel database
DB_CHANNEL_KEY = 'channel'.encode()  # database name

# Descriptor -> Timestamp database
DB_TIMESTAMP_KEY = 'timestamp'.encode()  # database name


@icontract.require(lambda database_dir: database_dir.exists())
def initialize_environment(database_dir: pathlib.Path) -> None:
    """
    Initialize the database; the database directory is assumed to exist.

    :param database_dir: where the database is stored
    :return:

    """
    with lmdb.open(path=database_dir.as_posix(), map_size=32 * 1024 * 1024 * 1024, max_dbs=2, readonly=False) as env:
        env.open_db(DB_CHANNEL_KEY, create=True)
        env.open_db(DB_TIMESTAMP_KEY, create=True)
