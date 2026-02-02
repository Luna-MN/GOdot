using Godot;
using System;
using Packets;

public interface State
{
    Log log { get; set; }
    void OnPacketReceived(Packet packet);
    void OnWSConnectionClosed();

    void ExitTree()
    {
        globals.WS.connectionClosed -= OnWSConnectionClosed;
        globals.WS.packetRecived -= OnPacketReceived;
    }
}
