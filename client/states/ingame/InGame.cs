using Godot;
using System;
using Packets;

public partial class InGame : Node
{
    [Export] private Log log;
    [Export] private LineEdit chatInput;

    public override void _Ready()
    {
        globals.WS.connectionClosed += OnWSConnectionClosed;
        globals.WS.packetRecived += OnPacketReceived;
        
        chatInput.TextSubmitted += chatTextSubmitted;
    }
    public void OnWSConnectionClosed()
    {
        log.warning("Connection closed.");
    }
    public void OnPacketReceived(Packet packet)
    {
        var senderId = packet.SenderId;
        if (packet.MsgCase == Packet.MsgOneofCase.Chat)
        {
            HandleChatMessage(senderId, packet.Chat);
        }
    }
    public void HandleChatMessage(ulong senderId, ChatMessage msg)
    {
        log.chat($"Client {senderId}", msg.Msg);
    }
    public void chatTextSubmitted(string text)
    {
        var packet = new Packet();
        packet.Chat = new ChatMessage {Msg = text};
        
        var err = globals.WS.Send(packet);
        if (err != Error.Ok)
        {
            log.error($"Failed to send packet: {err}");
        }
        else
        {
            log.chat("Me", text);
        }
        chatInput.Clear();
    }
}
