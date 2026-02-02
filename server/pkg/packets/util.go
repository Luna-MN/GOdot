package packets

import (
	"server/internal/server/objects"
)

type Msg = isPacket_Msg

func NewChat(msg string) *Packet_Chat {
	return &Packet_Chat{
		Chat: &ChatMessage{
			Msg: msg,
		},
	}
}

func NewId(id uint64) *Packet_Id {
	return &Packet_Id{
		Id: &IdMessage{
			Id: id,
		},
	}
}

func NewOK() Msg {
	return &Packet_OK{
		OK: &OkResponseMessage{},
	}
}

func NewDeny(reason string) Msg {
	return &Packet_Deny{
		Deny: &DenyResponseMessage{
			Reason: reason,
		},
	}
}

func NewPlayer(id uint64, player *objects.Player) Msg {
	return &Packet_Player{
		Player: &PlayerMessage{
			Id:    id,
			Name:  player.Name,
			X:     player.Position.X,
			Y:     player.Position.Y,
			Z:     player.Position.Z,
			Speed: player.Speed,
		},
	}
}
