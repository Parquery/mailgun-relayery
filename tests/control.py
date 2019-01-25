#!/usr/bin/env python3
# Automatically generated file by swagger_to. DO NOT EDIT OR APPEND ANYTHING!
"""Implements the client for control."""

# pylint: skip-file
# pydocstyle: add-ignore=D105,D107,D401

import contextlib
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

    if exp == Entity:
        return entity_from_obj(obj, path=path)

    if exp == Channel:
        return channel_from_obj(obj, path=path)

    if exp == ChannelsPage:
        return channels_page_from_obj(obj, path=path)

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

    if exp == Entity:
        assert isinstance(obj, Entity)
        return entity_to_jsonable(obj, path=path)

    if exp == Channel:
        assert isinstance(obj, Channel)
        return channel_to_jsonable(obj, path=path)

    if exp == ChannelsPage:
        assert isinstance(obj, ChannelsPage)
        return channels_page_to_jsonable(obj, path=path)

    raise ValueError("Unexpected `expected` type: {}".format(exp))


class Entity:
    """Contains the email address and optionally the name of an entity."""

    def __init__(self, email: str, name: Optional[str] = None) -> None:
        """Initializes with the given values."""
        self.email = email

        self.name = name

    def to_jsonable(self) -> MutableMapping[str, Any]:
        """
        Dispatches the conversion to entity_to_jsonable.

        :return: JSON-able representation
        """
        return entity_to_jsonable(self)


def new_entity() -> Entity:
    """Generates an instance of Entity with default values."""
    return Entity(email='')


def entity_from_obj(obj: Any, path: str = "") -> Entity:
    """
    Generates an instance of Entity from a dictionary object.

    :param obj: a JSON-ed dictionary object representing an instance of Entity
    :param path: path to the object used for debugging
    :return: parsed instance of Entity
    """
    if not isinstance(obj, dict):
        raise ValueError('Expected a dict at path {}, but got: {}'.format(path, type(obj)))

    for key in obj:
        if not isinstance(key, str):
            raise ValueError('Expected a key of type str at path {}, but got: {}'.format(path, type(key)))

    email_from_obj = from_obj(obj['email'], expected=[str], path=path + '.email')  # type: str

    if 'name' in obj:
        name_from_obj = from_obj(obj['name'], expected=[str], path=path + '.name')  # type: Optional[str]
    else:
        name_from_obj = None

    return Entity(email=email_from_obj, name=name_from_obj)


def entity_to_jsonable(entity: Entity, path: str = "") -> MutableMapping[str, Any]:
    """
    Generates a JSON-able mapping from an instance of Entity.

    :param entity: instance of Entity to be JSON-ized
    :param path: path to the entity used for debugging
    :return: a JSON-able representation
    """
    res = dict()  # type: Dict[str, Any]

    res['email'] = entity.email

    if entity.name is not None:
        res['name'] = entity.name

    return res


class Channel:
    """Defines the messaging channel."""

    def __init__(self,
                 descriptor: str,
                 token: str,
                 sender: Entity,
                 recipients: List[Entity],
                 domain: str,
                 min_period: float,
                 max_size: int,
                 cc: Optional[List[Entity]] = None,
                 bcc: Optional[List[Entity]] = None) -> None:
        """Initializes with the given values."""
        self.descriptor = descriptor

        self.token = token

        self.sender = sender

        self.recipients = recipients

        # indicates the MailGun domain for the channel.
        self.domain = domain

        # is the minimum push period frequency for a channel, in seconds.
        self.min_period = min_period

        # indicates the maximum allowed size of the request, in bytes.
        self.max_size = max_size

        self.cc = cc

        self.bcc = bcc

    def to_jsonable(self) -> MutableMapping[str, Any]:
        """
        Dispatches the conversion to channel_to_jsonable.

        :return: JSON-able representation
        """
        return channel_to_jsonable(self)


def new_channel() -> Channel:
    """Generates an instance of Channel with default values."""
    return Channel(descriptor='', token='', sender=new_entity(), recipients=[], domain='', min_period=0.0, max_size=0)


def channel_from_obj(obj: Any, path: str = "") -> Channel:
    """
    Generates an instance of Channel from a dictionary object.

    :param obj: a JSON-ed dictionary object representing an instance of Channel
    :param path: path to the object used for debugging
    :return: parsed instance of Channel
    """
    if not isinstance(obj, dict):
        raise ValueError('Expected a dict at path {}, but got: {}'.format(path, type(obj)))

    for key in obj:
        if not isinstance(key, str):
            raise ValueError('Expected a key of type str at path {}, but got: {}'.format(path, type(key)))

    descriptor_from_obj = from_obj(obj['descriptor'], expected=[str], path=path + '.descriptor')  # type: str

    token_from_obj = from_obj(obj['token'], expected=[str], path=path + '.token')  # type: str

    sender_from_obj = from_obj(obj['sender'], expected=[Entity], path=path + '.sender')  # type: Entity

    recipients_from_obj = from_obj(
        obj['recipients'], expected=[list, Entity], path=path + '.recipients')  # type: List[Entity]

    domain_from_obj = from_obj(obj['domain'], expected=[str], path=path + '.domain')  # type: str

    min_period_from_obj = from_obj(obj['min_period'], expected=[float], path=path + '.min_period')  # type: float

    max_size_from_obj = from_obj(obj['max_size'], expected=[int], path=path + '.max_size')  # type: int

    if 'cc' in obj:
        cc_from_obj = from_obj(obj['cc'], expected=[list, Entity], path=path + '.cc')  # type: Optional[List[Entity]]
    else:
        cc_from_obj = None

    if 'bcc' in obj:
        bcc_from_obj = from_obj(obj['bcc'], expected=[list, Entity], path=path + '.bcc')  # type: Optional[List[Entity]]
    else:
        bcc_from_obj = None

    return Channel(
        descriptor=descriptor_from_obj,
        token=token_from_obj,
        sender=sender_from_obj,
        recipients=recipients_from_obj,
        domain=domain_from_obj,
        min_period=min_period_from_obj,
        max_size=max_size_from_obj,
        cc=cc_from_obj,
        bcc=bcc_from_obj)


def channel_to_jsonable(channel: Channel, path: str = "") -> MutableMapping[str, Any]:
    """
    Generates a JSON-able mapping from an instance of Channel.

    :param channel: instance of Channel to be JSON-ized
    :param path: path to the channel used for debugging
    :return: a JSON-able representation
    """
    res = dict()  # type: Dict[str, Any]

    res['descriptor'] = channel.descriptor

    res['token'] = channel.token

    res['sender'] = to_jsonable(channel.sender, expected=[Entity], path='{}.sender'.format(path))

    res['recipients'] = to_jsonable(channel.recipients, expected=[list, Entity], path='{}.recipients'.format(path))

    res['domain'] = channel.domain

    res['min_period'] = channel.min_period

    res['max_size'] = channel.max_size

    if channel.cc is not None:
        res['cc'] = to_jsonable(channel.cc, expected=[list, Entity], path='{}.cc'.format(path))

    if channel.bcc is not None:
        res['bcc'] = to_jsonable(channel.bcc, expected=[list, Entity], path='{}.bcc'.format(path))

    return res


class ChannelsPage:
    """Lists channels in a paginated manner."""

    def __init__(self, page: int, page_count: int, per_page: int, channels: List[Channel]) -> None:
        """Initializes with the given values."""
        # specifies the index of the page.
        self.page = page

        # specifies the number of pages available.
        self.page_count = page_count

        # specifies the number of items per page.
        self.per_page = per_page

        # contains the channel data.
        self.channels = channels

    def to_jsonable(self) -> MutableMapping[str, Any]:
        """
        Dispatches the conversion to channels_page_to_jsonable.

        :return: JSON-able representation
        """
        return channels_page_to_jsonable(self)


def new_channels_page() -> ChannelsPage:
    """Generates an instance of ChannelsPage with default values."""
    return ChannelsPage(page=0, page_count=0, per_page=0, channels=[])


def channels_page_from_obj(obj: Any, path: str = "") -> ChannelsPage:
    """
    Generates an instance of ChannelsPage from a dictionary object.

    :param obj: a JSON-ed dictionary object representing an instance of ChannelsPage
    :param path: path to the object used for debugging
    :return: parsed instance of ChannelsPage
    """
    if not isinstance(obj, dict):
        raise ValueError('Expected a dict at path {}, but got: {}'.format(path, type(obj)))

    for key in obj:
        if not isinstance(key, str):
            raise ValueError('Expected a key of type str at path {}, but got: {}'.format(path, type(key)))

    page_from_obj = from_obj(obj['page'], expected=[int], path=path + '.page')  # type: int

    page_count_from_obj = from_obj(obj['page_count'], expected=[int], path=path + '.page_count')  # type: int

    per_page_from_obj = from_obj(obj['per_page'], expected=[int], path=path + '.per_page')  # type: int

    channels_from_obj = from_obj(
        obj['channels'], expected=[list, Channel], path=path + '.channels')  # type: List[Channel]

    return ChannelsPage(
        page=page_from_obj, page_count=page_count_from_obj, per_page=per_page_from_obj, channels=channels_from_obj)


def channels_page_to_jsonable(channels_page: ChannelsPage, path: str = "") -> MutableMapping[str, Any]:
    """
    Generates a JSON-able mapping from an instance of ChannelsPage.

    :param channels_page: instance of ChannelsPage to be JSON-ized
    :param path: path to the channels_page used for debugging
    :return: a JSON-able representation
    """
    res = dict()  # type: Dict[str, Any]

    res['page'] = channels_page.page

    res['page_count'] = channels_page.page_count

    res['per_page'] = channels_page.per_page

    res['channels'] = to_jsonable(channels_page.channels, expected=[list, Channel], path='{}.channels'.format(path))

    return res


class RemoteCaller:
    """Executes the remote calls to the server."""

    def __init__(self, url_prefix: str, auth: Optional[requests.auth.AuthBase] = None) -> None:
        self.url_prefix = url_prefix
        self.auth = auth

    def put_channel(self, channel: Channel) -> bytes:
        """
        Updates the channel uniquely identified by a descriptor.

        If there is already a channel associated with the descriptor, the old channel is overwritten with the new one.

        In order to enforce the min_period between messages, the Relay server keeps track of the time of the most
        recently relayed message for each descriptor. If a channel is overwritten, the time of relay of the most
        recent message is erased unless the new channel has the same min_period field as the old one.

        :param channel:

        :return: signals that the channel update was accepted.
        """
        url = self.url_prefix + '/api/channel'

        data = to_jsonable(channel, expected=[Channel])

        resp = requests.request(method='put', url=url, json=data, auth=self.auth)

        with contextlib.closing(resp):
            resp.raise_for_status()
            return resp.content

    def delete_channel(self, descriptor: str) -> bytes:
        """
        Removes the channel associated with the descriptor.

        :param descriptor:

        :return: signals that the channel was correctly erased, or that the channel was not found.
        """
        url = self.url_prefix + '/api/channel'

        data = descriptor

        resp = requests.request(method='delete', url=url, json=data, auth=self.auth)

        with contextlib.closing(resp):
            resp.raise_for_status()
            return resp.content

    def list_channels(self, page: Optional[int] = None, per_page: Optional[int] = None) -> ChannelsPage:
        """
        Lists the available channels information.

        :param page: specifies the index of a page. The default is 1 (first page).
        :param per_page: specifies a desired number of items per page. The default is 100.

        :return: serves the channel information list.
        """
        url = self.url_prefix + '/api/list_channels'

        params = {'page': page, 'per_page': per_page}

        resp = requests.request(method='get', url=url, params=params, auth=self.auth)

        with contextlib.closing(resp):
            resp.raise_for_status()
            return from_obj(obj=resp.json(), expected=[ChannelsPage])


# Automatically generated file by swagger_to. DO NOT EDIT OR APPEND ANYTHING!
