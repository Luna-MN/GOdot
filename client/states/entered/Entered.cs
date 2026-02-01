using Godot;
using System;
using Packets;

public partial class Entered : Node
{
    [Export] private Log log;

    public override void _Ready()
    {
        
        globals.WS.connectedToServer += OnWSConnectedToServer;
        globals.WS.connectionClosed += OnWSConnectionClosed;
        globals.WS.packetRecived += OnPacketReceived;

        log.info("Connecting to server...");
        globals.WS.connectToUrl("ws://localhost:8080/ws", null);
    }

    public void OnWSConnectedToServer()
    {
        log.success("Connected to server.");
        
    }
    public void OnWSConnectionClosed()
    {
        log.warning("Connection closed.");
    }

    public void OnPacketReceived(Packet packet)
    {
        var senderId = packet.SenderId;
        if (packet.MsgCase == Packet.MsgOneofCase.Id)
        {
            HandleIdMessage(senderId, packet.Id);
        }
    }

    public void HandleIdMessage(ulong senderId, IdMessage msg)
    {
        globals.GM.clientId = msg.Id;
        log.info($"Client ID: {globals.GM.clientId}");
        globals.GM.SetState(GameManager.state.Connected);
    }
}
