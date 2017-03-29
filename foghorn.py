import json
import socket
import subprocess
import sys

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)

server_address = ('localhost', 10110)
print('starting on {}'.format(server_address))

sock.bind(server_address)

subprocess.Popen(['/home/jwittrock/src/rtl-ais/rtl_ais'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

print('Continuing...')

while True:
    data, address = sock.recvfrom(4096)
    print(data)
