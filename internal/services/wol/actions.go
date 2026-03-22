package wol

import (
	"context"
	"fmt"
	"net"

	"github.com/Grino777/wol-server/internal/core/entity"
)

func (s *wolService) WakeServer(ctx context.Context, id int) error {
	server, err := s.serverRepository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if server == nil {
		return fmt.Errorf("server with id=%d not found", id)
	}

	mp, err := s.generateMagicPacket(server.MacAddress)
	if err != nil {
		return err
	}

	if err := s.sendMagicPacket(server.IpAddress, mp); err != nil {
		return err
	}

	return nil
}

func (s *wolService) OffServer(ctx context.Context, id int) error {
	server, err := s.serverRepository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if server == nil {
		return fmt.Errorf("server with id=%d not found", id)
	}

	// TODO: graceful shutdown/turn-off command integration.
	return nil
}

func (s *wolService) generateMagicPacket(mac string) ([]byte, error) {
	hw, err := net.ParseMAC(mac)
	if err != nil {
		return nil, err
	}

	packet := make([]byte, 0, 6+16*len(hw))

	// 6 байт FF
	for range 6 {
		packet = append(packet, 0xFF)
	}

	// 16 повторений MAC
	for range 16 {
		packet = append(packet, hw...)
	}

	return packet, nil
}

func (s *wolService) sendMagicPacket(ip string, mp []byte) error {
	conn, err := net.Dial("udp", ip+":9")
	if err != nil {
		return entity.ErrMagicPacketNotSent
	}
	defer conn.Close()

	if _, err := conn.Write(mp); err != nil {
		return entity.ErrMagicPacketNotSent
	}

	return nil
}
