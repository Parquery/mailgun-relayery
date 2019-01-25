#!/usr/bin/env python3
"""
Release the code to the given directory as a binary package and a debian package.

The architecture is assumed to be AMD64 (i.e. Linux x64). If you want to release the code for a different architecture,
then please do that manually.
"""

import argparse
import os
import pathlib
import shutil
import subprocess
import sys
import tempfile
import textwrap


# pylint: disable=too-many-locals
def main() -> int:
    """Execute the main routine."""
    parser = argparse.ArgumentParser()
    parser.add_argument("--release_dir", help="directory where to put the release", required=True)
    args = parser.parse_args()

    release_dir = pathlib.Path(args.release_dir)

    release_dir.mkdir(exist_ok=True, parents=True)

    bin_dir = release_dir / 'bin'
    bin_dir.mkdir(parents=True, exist_ok=True)

    pkg_dir = release_dir / 'pkg'
    pkg_dir.mkdir(parents=True, exist_ok=True)

    env = os.environ.copy()
    if 'GOPATH' in env:
        env['GOPATH'] = '{}:{}'.format(release_dir.as_posix(), env['GOPATH'])
    else:
        env['GOPATH'] = release_dir.as_posix()

    env['GOBIN'] = bin_dir.as_posix()
    env['GOPKG'] = pkg_dir.as_posix()

    # set the working directory to the script's directory
    script_dir = pathlib.Path(os.path.dirname(os.path.realpath(__file__)))
    subprocess.check_call(['go', 'install', '-i', './...'], env=env, cwd=script_dir.as_posix())

    go_bin_dir = release_dir / "bin"
    bin_path = go_bin_dir / "mailgun-relayery"

    # Get mailgun-relayery version
    version = subprocess.check_output([bin_path.as_posix(), "-version"], universal_newlines=True).strip()

    # Release the binary package
    with tempfile.TemporaryDirectory() as tmp_dir:
        bin_package_dir = pathlib.Path(tmp_dir) / "mailgun-relayery-{}-linux-x64".format(version)

        target = bin_package_dir / "bin/mailgun-relayery"
        target.parent.mkdir(parents=True)

        shutil.copy(bin_path.as_posix(), target.as_posix())

        tar_path = bin_package_dir.parent / "mailgun-relayery-{}-linux-x64.tar.gz".format(version)

        subprocess.check_call(
            ["tar", "-czf", tar_path.as_posix(), "mailgun-relayery-{}-linux-x64".format(version)],
            cwd=bin_package_dir.parent.as_posix())

        shutil.move(tar_path.as_posix(), (release_dir / tar_path.name).as_posix())

    # Release the debian package
    with tempfile.TemporaryDirectory() as tmp_dir:
        deb_package_dir = pathlib.Path(tmp_dir) / "mailgun-relayery_{}_amd64".format(version)

        target = deb_package_dir / "usr/bin/mailgun-relayery"
        target.parent.mkdir(parents=True)
        shutil.copy(bin_path.as_posix(), target.as_posix())

        control_pth = deb_package_dir / "DEBIAN/control"
        control_pth.parent.mkdir(parents=True)

        control_pth.write_text(
            textwrap.dedent('''\
            Package: mailgun-relayery
            Version: {version}
            Maintainer: Teodoro Filippini (teodoro.filippini@gmail.com)
            Architecture: amd64
            Description: mailgun-relayery is a tool wrapping the MailGun API with an addition security layer.
            '''.format(version=version)))

        subprocess.check_call(
            ["dpkg-deb", "--build", deb_package_dir.as_posix()],
            cwd=deb_package_dir.parent.as_posix(),
            stdout=subprocess.DEVNULL)

        deb_pth = deb_package_dir.parent / "mailgun-relayery_{}_amd64.deb".format(version)

        shutil.move(deb_pth.as_posix(), (release_dir / deb_pth.name).as_posix())

    print("Released to: {}".format(release_dir))

    return 0


if __name__ == "__main__":
    sys.exit(main())
