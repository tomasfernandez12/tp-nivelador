import socket


def recv_all(socket: socket.socket, size):
    received = b""
    while len(received) < size:
        chunk = socket.recv(size - len(received))
        if not chunk:
            break
        received += chunk
    return received


def send_all(socket: socket.socket, bytes):
    sent = 0
    while sent < len(bytes):
        sent_now = socket.send(bytes[sent:])
        if sent_now == 0:
            raise RuntimeError("socket connection broken")
        sent += sent_now
