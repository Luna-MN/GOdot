using Godot;
using System;
using Packets;

public partial class Entered : Node, State
{
    [Export] public Log log { get; set; }
    public bool enetSet, idSet;
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
        log.info("Handling ID message. " + senderId);
        log.info($"Received ID: {msg.Id}");
        log.info("Setting client ID.");
        globals.GM.clientId = msg.Id;
        log.info($"Client ID: {globals.GM.clientId} sender ID: {senderId}");
        idSet = true;
        var e = new ENet();
        GetTree().Root.AddChild(e);
        globals.WS.Send(new Packet {SenderId = senderId, Id = new IdMessage {Id = globals.GM.clientId}});
        log.info("Entering game state.");
        globals.GM.SetState(GameManager.state.Connected);

    }
    public void ExitTree()
    {
        globals.WS.connectedToServer -= OnWSConnectedToServer;
        ((State)this).ExitTree();
    }
}
