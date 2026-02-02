using Godot;
using System;

public partial class Player : Node
{
    private static readonly PackedScene playerScene = GD.Load<PackedScene>("res://Objects/Player/Player.tscn");
    
    public bool isMe { get; private set; }
    
    public static (Player, Actor) instantiate(ulong id, string name, Vector3 pos, float speed, bool isMe)
    {
        var actor = Actor.instantiate(id, name, pos, speed);
        var player = playerScene.Instantiate<Player>();
        player.isMe = isMe;
        actor.AddChild(player);
        return (player, actor);
    }
}
