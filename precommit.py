#!/usr/bin/env python3
"""Run all pre-commit checks."""
import argparse
import os
import pathlib
import subprocess
import sys
import tempfile


# pylint: disable=too-many-locals, too-many-statements, too-many-branches
def main() -> int:
    """Execute the main routine."""
    parser = argparse.ArgumentParser()
    parser.add_argument("--overwrite", help="if set, overwrites the malformatted files", action='store_true')
    args = parser.parse_args()
    overwrite = bool(args.overwrite)

    here = pathlib.Path(os.path.abspath(__file__)).parent
    src_dir = here.parent.parent.parent.parent

    env = os.environ.copy()
    if 'GOPATH' in env:
        env['GOPATH'] = '{}:{}'.format(src_dir, env['GOPATH'])
    else:
        env['GOPATH'] = (here / "mailgun-relayery").as_posix()

    pths = sorted(here.glob("*.go"))

    for subpth in here.iterdir():
        if subpth.is_dir() and not subpth.name.startswith('.'):
            pths.extend(subpth.glob("**/*.go"))

    pths = [x for x in pths if 'vendor' not in x.as_posix()]

    # dep ensure
    subprocess.check_call(['dep', 'ensure'], env=env, cwd=here.as_posix())

    # gofmt
    for pth in pths:
        if overwrite:
            retcode = subprocess.call(['gofmt', '-s', '-w', pth.as_posix()], env=env)
            if retcode != 0:
                raise RuntimeError("Failed to gofmt: {}".format(pth))
        else:
            out = subprocess.check_output(["gofmt", "-s", "-l", pth.as_posix()], env=env)
            if len(out) != 0:
                print("Code was not formatted; gofmt -s -l complains: {}".format(pth))
                return 1

    # gocontracts
    for pth in pths:
        if overwrite:
            retcode = subprocess.call(['gocontracts', '-w', pth.as_posix()])
            if retcode != 0:
                raise RuntimeError("Failed to gocontracts: {}".format(pth))
        else:
            with tempfile.NamedTemporaryFile() as tmpfile:
                file = tmpfile.file # type: ignore
                retcode = subprocess.call(
                    ['gocontracts', pth.as_posix()], universal_newlines=True, env=env, stdout=file)
                if retcode != 0:
                    raise RuntimeError("Failed to gocontracts: {}".format(pth))

                file.flush()
                proc = subprocess.Popen(
                    ['diff', tmpfile.name, pth.as_posix()],
                    stderr=subprocess.STDOUT,
                    stdout=subprocess.PIPE,
                    universal_newlines=True)

                out, _ = proc.communicate()
                if proc.returncode != 0:
                    raise RuntimeError("gocontracts did not match for {}:\n{}".format(pth, out))

    packages = [
        pathlib.Path(line)
        for line in subprocess.check_output(["go", "list", "./..."], universal_newlines=True, env=env).splitlines()
        if line.strip()
    ]

    packages = [pkg for pkg in packages if 'vendor' not in pkg.parents]

    subprocess.check_call(['go', 'vet'] + [pkg.as_posix() for pkg in packages], env=env, cwd=here.as_posix())

    subprocess.check_call(['golint'] + [pkg.as_posix() for pkg in packages], env=env, cwd=here.as_posix())

    subprocess.check_call(['errcheck'] + [pkg.as_posix() for pkg in packages], env=env, cwd=here.as_posix())

    for pkg in packages:
        subprocess.check_call(['go', 'test', pkg.as_posix()], env=env, cwd=here.as_posix())

    subprocess.check_call(['go', 'build', './...'], env=env, cwd=here.as_posix())

    # check that the servers return the same -version
    version_rel = subprocess.check_output(
        ['go', 'run', 'mailgun-relayery/main.go', '-version'], cwd=here.as_posix(), env=env,
        universal_newlines=True).strip()
    version_ctl = subprocess.check_output(
        ['go', 'run', 'mailgun-relay-controlery/main.go', '-version'],
        cwd=here.as_posix(),
        env=env,
        universal_newlines=True).strip()
    if version_ctl != version_rel:
        raise RuntimeError("Expected the relay server version {} and control server "
                           "version {} to coincide.".format(version_rel, version_ctl))

    # Check that the CHANGELOG.md is consistent with -version
    changelog_pth = here / "CHANGELOG.md"
    with changelog_pth.open('rt') as fid:
        changelog_lines = fid.read().splitlines()

    if len(changelog_lines) < 1:
        raise RuntimeError(
            "Expected at least a line in {}, but got: {} line(s)".format(changelog_pth, len(changelog_lines)))

    # Expect the first line to always refer to the latest version
    changelog_version = changelog_lines[0]

    if version_ctl != changelog_version:
        raise AssertionError(("The version in changelog file {} (parsed as its first line), {!r}, "
                              "does not match the output of -version: {!r}").format(changelog_pth, changelog_version,
                                                                                    version_ctl))
    return 0


if __name__ == "__main__":
    sys.exit(main())
