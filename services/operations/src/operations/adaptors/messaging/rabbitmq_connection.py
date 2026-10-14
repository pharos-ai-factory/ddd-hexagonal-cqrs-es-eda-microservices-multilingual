"""Configure bounded RabbitMQ connections for publishers and consumers."""

import pika
from pika.adapters.blocking_connection import BlockingConnection


def broker_connection(url: str) -> BlockingConnection:
    parameters = pika.URLParameters(url)
    parameters.heartbeat = 30
    parameters.blocked_connection_timeout = 5
    parameters.socket_timeout = 5
    return pika.BlockingConnection(parameters)
