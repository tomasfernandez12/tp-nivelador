from lottery import Bet


def _deserialize_bet(payload: bytes, position: int) -> tuple[Bet, int]:
    def read_integer():
        nonlocal position
        if len(payload) - position < 4:
            raise ValueError("invalid bet")
        value = int.from_bytes(payload[position : position + 4], "big")
        position += 4
        return value

    def read_string():
        nonlocal position
        length = read_integer()
        if len(payload) - position < length:
            raise ValueError("invalid bet")
        value = payload[position : position + length].decode()
        position += length
        return value

    agency_id = read_integer()
    first_name = read_string()
    last_name = read_string()
    document = read_integer()
    birthdate = read_string()
    number = read_integer()
    return Bet(
        int(agency_id),
        first_name,
        last_name,
        int(document),
        birthdate,
        int(number),
    ), position


def deserialize_bet(payload: bytes) -> Bet:
    bet, position = _deserialize_bet(payload, 0)
    if position != len(payload):
        raise ValueError("invalid bet")
    return bet


def deserialize_bets(payload: bytes) -> list[Bet]:
    bets = []
    position = 0
    while position < len(payload):
        bet, position = _deserialize_bet(payload, position)
        bets.append(bet)
    return bets


def serialize_bet(bet: Bet) -> bytes:
    payload = bytearray()

    def add_integer(value):
        payload.extend(value.to_bytes(4, "big"))

    def add_string(value):
        encoded = value.encode()
        add_integer(len(encoded))
        payload.extend(encoded)

    add_integer(bet.agency_id)
    add_string(bet.first_name)
    add_string(bet.last_name)
    add_integer(bet.document)
    add_string(bet.birthdate)
    add_integer(bet.number)
    return bytes(payload)


def deserialize_message_size(header: bytes) -> int:
    return int.from_bytes(header, "big")


def serialize_message_size(payload: bytes) -> bytes:
    return len(payload).to_bytes(4, "big")