import socket
import logger
import safe_socket
from lottery import Lottery
from .serialization import (
    deserialize_bet,
    deserialize_bets,
    deserialize_message_size,
    serialize_bet,
    serialize_message_size,
)

_MESSAGE_HEADER_SIZE = 4

class Server:
    def __init__(self, server_host: str, server_port: int) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.lottery = Lottery("/tmp/lottery.csv")

    def _receive_message(self, client_socket):
        header = safe_socket.recv_all(client_socket, _MESSAGE_HEADER_SIZE)
        if len(header) != _MESSAGE_HEADER_SIZE:
            raise ConnectionError("incomplete message header")
        size = deserialize_message_size(header)
        return safe_socket.recv_all(client_socket, size)

    def _send_message(self, client_socket, payload: bytes):
        safe_socket.send_all(client_socket, serialize_message_size(payload))
        safe_socket.send_all(client_socket, payload)

    def _handle_client(self, client_socket):
        action = "handle-client"
        message_amount = 0
        try:
            logger.info(action, logger.LogResult.in_progress)
            bets = []
            while True:
                client_message = self._receive_message(client_socket)
                if not client_message:
                    break
                message_amount += 1
                bets.extend(deserialize_bets(client_message))

            self.lottery.store_bets(bets)
            agency_id = bets[0].agency_id if bets else None
            if agency_id is not None:
                winners = self.lottery.get_winners_for_agency(
                    self.lottery.load_bets(), agency_id
                )
                for winner in winners:
                    self._send_message(client_socket, serialize_bet(winner))
            self._send_message(client_socket, b"")
            logger.info(
                action,
                logger.LogResult.success,
                "messages-amount",
                message_amount,
            )
        except Exception as e:
            logger.error(
                action, logger.LogResult.fail, "messages-amount", message_amount
            )
            raise e

    def run(self):
        action = "accept-connection"
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            while True:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

                self._handle_client(client_socket)
