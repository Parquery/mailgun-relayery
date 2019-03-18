#!/usr/bin/env python3
# Automatically generated file by swagger_to. DO NOT EDIT OR APPEND ANYTHING!
"""Implements the client for relay."""

# pylint: skip-file
# pydocstyle: add-ignore=D105,D107,D401

import contextlib
import json
from typing import Any, BinaryIO, Dict, List, MutableMapping, Optional

import requests
import requests.auth


def from_obj(obj: Any, expected: List[type], path: str = '') -> Any:
    """
    Checks and converts the given obj along the expected types.

    :param obj: to be converted
    :param expected: list of types representing the (nested) structure
    :param path: to the object used for debugging
    :return: the converted object
    """
    if not expected:
        raise ValueError("`expected` is empty, but at least one type needs to be specified.")

    exp = expected[0]

    if exp == float:
        if isinstance(obj, int):
            return float(obj)

        if isinstance(obj, float):
            return obj

        raise ValueError('Expected object of type int or float at {!r}, but got {}.'.format(path, type(obj)))

    if exp in [bool, int, str, list, dict]:
        if not isinstance(obj, exp):
            raise ValueError('Expected object of type {} at {!r}, but got {}.'.format(exp, path, type(obj)))

    if exp in [bool, int, float, str]:
        return obj

    if exp == list:
        lst = []  # type: List[Any]
        for i, value in enumerate(obj):
            lst.append(from_obj(value, expected=expected[1:], path='{}[{}]'.format(path, i)))

        return lst

    if exp == dict:
        adict = dict()  # type: Dict[str, Any]
        for key, value in obj.items():
            if not isinstance(key, str):
                raise ValueError('Expected a key of type str at path {!r}, got: {}'.format(path, type(key)))

            adict[key] = from_obj(value, expected=expected[1:], path='{}[{!r}]'.format(path, key))

        return adict

    if exp == Message:
        return message_from_obj(obj, path=path)

    raise ValueError("Unexpected `expected` type: {}".format(exp))


def to_jsonable(obj: Any, expected: List[type], path: str = "") -> Any:
    """
    Checks and converts the given object along the expected types to a JSON-able representation.

    :param obj: to be converted
    :param expected: list of types representing the (nested) structure
    :param path: path to the object used for debugging
    :return: JSON-able representation of the object
    """
    if not expected:
        raise ValueError("`expected` is empty, but at least one type needs to be specified.")

    exp = expected[0]
    if not isinstance(obj, exp):
        raise ValueError('Expected object of type {} at path {!r}, but got {}.'.format(exp, path, type(obj)))

    # Assert on primitive types to help type-hinting.
    if exp == bool:
        assert isinstance(obj, bool)
        return obj

    if exp == int:
        assert isinstance(obj, int)
        return obj

    if exp == float:
        assert isinstance(obj, float)
        return obj

    if exp == str:
        assert isinstance(obj, str)
        return obj

    if exp == list:
        assert isinstance(obj, list)

        lst = []  # type: List[Any]
        for i, value in enumerate(obj):
            lst.append(to_jsonable(value, expected=expected[1:], path='{}[{}]'.format(path, i)))

        return lst

    if exp == dict:
        assert isinstance(obj, dict)

        adict = dict()  # type: Dict[str, Any]
        for key, value in obj.items():
            if not isinstance(key, str):
                raise ValueError('Expected a key of type str at path {!r}, got: {}'.format(path, type(key)))

            adict[key] = to_jsonable(value, expected=expected[1:], path='{}[{!r}]'.format(path, key))

        return adict

    if exp == Message:
        assert isinstance(obj, Message)
        return message_to_jsonable(obj, path=path)

    raise ValueError("Unexpected `expected` type: {}".format(exp))


class Message:
    """Represents a message to be relayed."""

    def __init__(self, subject: str, content: str, html: Optional[str] = None) -> None:
        """Initializes with the given values."""
        # contains the text to be used as the email's subject.
        self.subject = subject

        # contains the text to be used as the email's content.
        self.content = content

        # contains the optional html text to be used as the email's content.
        #
        # If set, the "content" field of the Message is ignored.
        self.html = html

    def to_jsonable(self) -> MutableMapping[str, Any]:
        """
        Dispatches the conversion to message_to_jsonable.

        :return: JSON-able representation
        """
        return message_to_jsonable(self)


def new_message() -> Message:
    """Generates an instance of Message with default values."""
    return Message(subject='', content='')


def message_from_obj(obj: Any, path: str = "") -> Message:
    """
    Generates an instance of Message from a dictionary object.

    :param obj: a JSON-ed dictionary object representing an instance of Message
    :param path: path to the object used for debugging
    :return: parsed instance of Message
    """
    if not isinstance(obj, dict):
        raise ValueError('Expected a dict at path {}, but got: {}'.format(path, type(obj)))

    for key in obj:
        if not isinstance(key, str):
            raise ValueError('Expected a key of type str at path {}, but got: {}'.format(path, type(key)))

    subject_from_obj = from_obj(obj['subject'], expected=[str], path=path + '.subject')  # type: str

    content_from_obj = from_obj(obj['content'], expected=[str], path=path + '.content')  # type: str

    if 'html' in obj:
        html_from_obj = from_obj(obj['html'], expected=[str], path=path + '.html')  # type: Optional[str]
    else:
        html_from_obj = None

    return Message(subject=subject_from_obj, content=content_from_obj, html=html_from_obj)


def message_to_jsonable(message: Message, path: str = "") -> MutableMapping[str, Any]:
    """
    Generates a JSON-able mapping from an instance of Message.

    :param message: instance of Message to be JSON-ized
    :param path: path to the message used for debugging
    :return: a JSON-able representation
    """
    res = dict()  # type: Dict[str, Any]

    res['subject'] = message.subject

    res['content'] = message.content

    if message.html is not None:
        res['html'] = message.html

    return res


class RemoteCaller:
    """Executes the remote calls to the server."""

    def __init__(self, url_prefix: str, auth: Optional[requests.auth.AuthBase] = None) -> None:
        self.url_prefix = url_prefix
        self.auth = auth

    def put_message(self, x_descriptor: str, x_token: str, message: Message) -> bytes:
        """
        Sends a message to the server, which relays it to the MailGun API.

        The given (descriptor, token) pair are authenticated first.
        The message's metadata is determined by the channel information from the database.

        :param x_descriptor:
        :param x_token:
        :param message:

        :return: signals that the message was correctly relayed to MailGun.
        """
        url = self.url_prefix + '/api/message'

        headers = {}  # type: Dict[str, str]

        headers['X-Descriptor'] = x_descriptor

        headers['X-Token'] = x_token

        data = to_jsonable(message, expected=[Message])

        resp = requests.request(method='post', url=url, headers=headers, json=data, auth=self.auth)

        with contextlib.closing(resp):
            resp.raise_for_status()
            return resp.content


# Automatically generated file by swagger_to. DO NOT EDIT OR APPEND ANYTHING!
