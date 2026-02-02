using Godot;
using System;
using Packets;

public partial class Actor : CharacterBody3D
{
    private static readonly PackedScene actorScene = GD.Load<PackedScene>("res://Objects/Actor/Actor.tscn");
    
    public ulong actorId { get; private set; }
    public string ActorName { get; private set; }
    public float Speed { get; private set; } = 100f;

    public static Actor instantiate(ulong id, string name, Vector3 pos, float speed)
    {
        var actor = actorScene.Instantiate<Actor>();
        actor.actorId = id;
        actor.ActorName = name;
        actor.Position = pos;
        actor.Speed = speed;
        return actor;
    }
    
}
