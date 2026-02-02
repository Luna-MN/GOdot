using Godot;
using System;
using Packets;

public partial class Player : Node
{
    private static readonly PackedScene playerScene = GD.Load<PackedScene>("res://Objects/Player/Player.tscn");
    
    public bool isMe { get; private set; }
    public Actor Actor { get; private set; }
    public Vector3 Velocity { get; private set; } = Vector3.Zero;
    
    private Vector3 inputVector = Vector3.Zero;

    public void SetActor(Actor actor)
    {
        if (Actor == null)
        {
            Actor = actor;
        }
    }
    public static (Player, Actor) instantiate(ulong id, string name, Vector3 pos, float speed, bool isMe)
    {
        var actor = Actor.instantiate(id, name, pos, speed);
        var player = playerScene.Instantiate<Player>();
        player.isMe = isMe;
        actor.AddChild(player);
        return (player, actor);
    }

    public override void _PhysicsProcess(double delta)
    {
        Actor.Position += Velocity * (float) delta;
        if (!isMe) return;
        var dir = new Vector3(Input.GetActionStrength("ui_right") - Input.GetActionStrength("ui_left"), 0, Input.GetActionStrength("ui_down") - Input.GetActionStrength("ui_up")).Normalized();
        if (dir != inputVector)
        {
            Velocity = dir * Actor.Speed;
            inputVector = dir;
            var packet = new Packet()
            {
                PlayerInput = new PlayerInputMessage()
                {
                    X    = dir.X,
                    Y    = dir.Y,
                    Z    = dir.Z
                }
            };
            GD.Print("X " + packet.PlayerInput.X,"Y " + packet.PlayerInput.Y,"Z " + packet.PlayerInput.Z);
            globals.WS.Send(packet);
        }
    }
}
