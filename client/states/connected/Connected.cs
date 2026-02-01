using Godot;
using System;
using Packets;

public partial class Connected : Node
{
    private Callable ActionOnOkReceived;
    
    [Export] private Log log;
    [Export] private LineEdit Username, Password;
    [Export] private Button LoginButton, RegisterButton;

    public override void _Ready()
    {
        globals.WS.packetRecived += OnWsPacketRecieved;
        globals.WS.connectionClosed += OnWsConnectionClosed;

        LoginButton.Pressed += OnLoginPressed;
        RegisterButton.Pressed += OnRegisterPressed;
    }
    private void OnLoginPressed()
    {
        var packet = new Packet();
        packet.LoginRequest = new LoginRequestMessage {Username = Username.Text, Password = Password.Text};
        globals.WS.Send(packet);
        ActionOnOkReceived = Callable.From(() => globals.GM.SetState(GameManager.state.InGame));
    }
    private void OnRegisterPressed()
    {
        var packet = new Packet();
        packet.RegisterRequest = new RegisterRequestMessage {Username = Username.Text, Password = Password.Text};
        globals.WS.Send(packet);
        ActionOnOkReceived = Callable.From(() => log.success("Registered successfully."));
    }

    private void OnWsPacketRecieved(Packet obj)
    {
        var senderId = obj.SenderId;
        if (obj.MsgCase == Packet.MsgOneofCase.Deny)
        {
            log.error("Login failed." + obj.Deny.Reason);
        }
        else if (obj.MsgCase == Packet.MsgOneofCase.OK)
        {
            ActionOnOkReceived.Call();
        }
    }
    private void OnWsConnectionClosed()
    {
        log.warning("Connection closed.");
    }


}
