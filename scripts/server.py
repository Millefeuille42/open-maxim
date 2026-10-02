#!/usr/bin/env python

# This is a basic python socket server to respond to the initial config request
#  Will likely be integrated in the chat server once I start working on it

# Haven't started with a http server since I was unsure of what protocol it used.

import socket

s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
s.bind(('0.0.0.0', 80))
s.listen(1)

print("Listening...")

while True:
    conn, addr = s.accept()
    print(f"\nConnection from {addr}")
    request = conn.recv(1024).decode('utf-8', errors='ignore')
    print(request)
    
    # Hardcoded IP to my br0 IP, funnily enough, the port serves no purpose since it stills hit on port 2002
    #  No idea what purpose is the advert URL either, more on that later. 
    body = "chat_server=192.168.137.1\nchat_port=6667\nadvert_url=http://192.168.137.1"
    
    response = (
        "HTTP/1.0 200 OK\r\n"
        "Content-Type: text/plain\r\n"
        f"Content-Length: {len(body)}\r\n"
        "Connection: close\r\n\r\n"
        f"{body}"
    )
    
    conn.sendall(response.encode('utf-8'))
    conn.close()
``