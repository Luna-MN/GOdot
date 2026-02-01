using Godot;
using System;
using System.Collections.Generic;

public partial class GameManager : Node
{
    public enum state
    {
        Entered,
        Connected,
        InGame,
    }
    public Dictionary<state, string> stateScenes = new Dictionary<state, string>()
    {
        { state.Entered, "res://states/entered/entered.tscn" },
        { state.InGame, "res://states/ingame/inGame.tscn" },
        { state.Connected, "res://states/connected/connected.tscn" }
    };

    public ulong clientId;
    public Node CurrentSceneRoot;

    public void SetState(state newState)
    {
        if(CurrentSceneRoot != null) CurrentSceneRoot.QueueFree();
        
        var scene = (PackedScene)ResourceLoader.Load(stateScenes[newState]);

        CurrentSceneRoot = scene.Instantiate();
        AddChild(CurrentSceneRoot);
    }

}
