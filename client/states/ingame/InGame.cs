using Godot;
using System;
using System.Collections.Generic;
using Packets;

public partial class InGame : Node, state
{
    [Export] public Log log { get; set; }
    [Export] private LineEdit chatInput;
    [Export] private Node3D world;
    private Dictionary<ulong, Player> Players = new Dictionary<ulong, Player>();

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
        else if (packet.MsgCase == Packet.MsgOneofCase.Player)
        {
            HandlePlayerMessge(senderId, packet.Player);
        }
    }

    private void HandlePlayerMessge(ulong senderId, PlayerMessage packetPlayer)
    {
        if (!Players.ContainsKey(senderId))
        {
            var pos = new Vector3((float)packetPlayer.X, (float)packetPlayer.Y, (float)packetPlayer.Z);
            var (player, actor) = Player.instantiate(senderId, packetPlayer.Name, pos, (float)packetPlayer.Speed, senderId == globals.GM.clientId);
            player.SetActor(actor);
            Players[senderId] = player; // [ senderId -> actor]
            world.AddChild(actor);
        }
        else
        {
            var player = Players[senderId];
            player.Actor.Position = new Vector3((float)packetPlayer.X, (float)packetPlayer.Y, (float)packetPlayer.Z);
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
