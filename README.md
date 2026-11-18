# GOdot

GOdot, is an open source client-server implimentation using Godot 4.7 and Golang with implimentations of both websocket communication for TCP and enet commucation for UDP, 
and includes a state managment system, please use this as a framework for any project you like and raise any issues if you would like any additional features!

### **IMPORTANAT**:

Before use you must generate the certification keys for the websocket communications run:

`openssl req -x509 -newkey rsa:4096 -keyout server/cert/server.key -out server/cert/server.crt -days 365 -nodes`

[Documentation](https://github.com/Luna-MN/GOdot-Multiplayer-Docs)
