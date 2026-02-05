using Godot;
using System;
using Google.Protobuf;
using Packets;

public partial class ENet : Node
{
    private ENetConnection conn;
    private ENetPacketPeer peer;
    public override void _Ready()
    {
        conn = new ENetConnection();
        conn.CreateHost(1, 1);
        // channels must match what your server is prepared to handle
        const int channels = 3;
        // user data (last param) is optional; only use it if your server reads it
        GD.Print((int)globals.GM.clientId);
        peer = conn.ConnectToHost("127.0.0.1", 8081, channels, (int)globals.GM.clientId);
    }

    public override void _Process(double delta)
    {
        poll();
    }
    private void poll()
    {
        conn.Service(10);
    }
    /// <summary>
    /// Send a packet via ENET, always sent unreliably
    /// </summary>
    /// <param name="packet"></param>
    /// <param name="channel"></param>
    /// <returns></returns>
    public Error Send(Packet packet, int channel)
    {
        packet.SenderId = 0;
        var data = packet.ToByteArray();
        var err = peer.Send(channel, data, 8);
        return err;
    }
}
