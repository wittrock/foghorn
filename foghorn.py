import json
import socket
import sys

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)

server_address = ('localhost', 8000)
print('starting on {}'.format(server_address))

sock.bind(server_address)

while True:
    data, address = sock.recvfrom(4096)
    print(data)
