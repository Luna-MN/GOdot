using Godot;
using System;
using System.Linq;
using Google.Protobuf;
using Packets;

public partial class main : Node
{
    public override void _Ready()
    {
        globals.WS = GetTree().Root.GetNode<websockets>("WS");
        globals.GM = GetTree().Root.GetNode<GameManager>("GameManager");
        
        globals.GM.SetState(GameManager.state.Entered);
    }
}
