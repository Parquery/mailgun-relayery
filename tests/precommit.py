#!/usr/bin/env python3
"""Run precommit checks on the Python files in the repository."""
import argparse
import concurrent.futures
import os
import pathlib
import subprocess
import sys
from typing import List, Union, Tuple  # pylint: disable=unused-import

import yapf.yapflib.yapf_api


def check(path: pathlib.Path, py_dir: pathlib.Path, overwrite: bool) -> Union[None, str]:
    """
    Run all the checks on the given file.

    :param path: to the source file
    :param py_dir: path to the source files
    :param overwrite: if True, overwrites the source file in place instead of reporting that it was not well-formatted.
    :return: None if all checks passed. Otherwise, an error message.
    """
    style_config = py_dir / 'style.yapf'

    report = []

    # yapf
    if not overwrite:
        formatted, _, changed = yapf.yapflib.yapf_api.FormatFile(
            filename=str(path), style_config=str(style_config), print_diff=True)

        if changed:
            report.append("Failed to yapf {}:\n{}".format(path, formatted))
    else:
        yapf.yapflib.yapf_api.FormatFile(filename=str(path), style_config=str(style_config), in_place=True)

    # mypy
    env = os.environ.copy()
    env['PYTHONPATH'] = ":".join([py_dir.as_posix(), env.get("PYTHONPATH", "")])

    proc = subprocess.Popen(
        ['mypy', str(path), '--ignore-missing-imports'],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        cwd=py_dir.as_posix(),
        env=env,
        universal_newlines=True)
    stdout, stderr = proc.communicate()
    if proc.returncode != 0:
        report.append("Failed to mypy {}:\nOutput:\n{}\n\nError:\n{}".format(path, stdout, stderr))

    # pylint
    proc = subprocess.Popen(
        ['pylint', str(path), '--rcfile={}'.format(py_dir / 'pylint.rc')],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        cwd=py_dir.as_posix(),
        env=env,
        universal_newlines=True)

    stdout, stderr = proc.communicate()
    if proc.returncode != 0:
        report.append("Failed to pylint {}:\nOutput:\n{}\n\nError:\n{}".format(path, stdout, stderr))

    # pydocstyle
    rel_pth = path.relative_to(py_dir)

    if rel_pth.parent.name != 'tests':
        proc = subprocess.Popen(
            ['pydocstyle', str(path)], stdout=subprocess.PIPE, stderr=subprocess.PIPE, universal_newlines=True)
        stdout, stderr = proc.communicate()
        if proc.returncode != 0:
            report.append("Failed to pydocstyle {}:\nOutput:\n{}\n\nError:\n{}".format(path, stdout, stderr))

    if len(report) > 0:
        return "\n".join(report)

    return None


def main() -> int:
    """Execute the main routine."""
    # pylint: disable=too-many-locals
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--overwrite",
        help="Overwrites the unformatted source files with the well-formatted code in place. "
             "If not set, an exception is raised if any of the files do not conform to the style guide.",
        action='store_true')
    parser.add_argument(
        "--max_workers",
        help="number of worker threads to run the tests on; "
             "if not specified, use the maximum number of threads available to ThreadPoolExecutor",
        type=int)

    args = parser.parse_args()

    overwrite = bool(args.overwrite)
    max_workers = None if args.max_workers is None else int(args.max_workers)

    repo_dir = pathlib.Path(os.path.realpath(__file__)).parent.parent

    # yapf: disable
    pths = sorted(
        list(repo_dir.glob("*.py")) +
        list((repo_dir / 'tests').glob("*.py")))
    # yapf: enable

    success = True

    futures_paths = []  # type: List[Tuple[concurrent.futures.Future, pathlib.Path]]
    with concurrent.futures.ThreadPoolExecutor(max_workers=max_workers) as executor:
        for pth in pths:
            if pth.stem == "precommit":
                continue
            future = executor.submit(fn=check, path=pth, py_dir=repo_dir, overwrite=overwrite)
            futures_paths.append((future, pth))

        for future, pth in futures_paths:
            report = future.result()
            if report is None:
                print("Passed all checks: {}".format(pth))
            else:
                print("One or more checks failed for {}:\n{}".format(pth, report))
                success = False

    if not success:
        print("One or more checks failed.")
        return 1

    return 0


if __name__ == "__main__":
    sys.exit(main())
