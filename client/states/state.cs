using Godot;
using System;
using Packets;

public interface state
{
    Log log { get; set; }
    void OnPacketReceived(Packet packet);
    void OnWSConnectionClosed();
    
}
