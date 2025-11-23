package service

import (
	"github.com/bytemare/opaque"
	"github.com/google/uuid"
)

type Opaque struct {
	server     *opaque.Server
	privateKey []byte
	publicKey  []byte
	oprfSeed   []byte // TODO: not ideal
	serverID   []byte
	config     []byte
	servers    map[uuid.UUID]*opaque.Server
}

func NewOpaque(server *opaque.Server, serverID, privateKey, publicKey, oprfSeed, config []byte) *Opaque {
	return &Opaque{
		server:     server,
		privateKey: privateKey,
		publicKey:  publicKey,
		oprfSeed:   oprfSeed,
		serverID:   serverID,
		config:     config,
		servers:    make(map[uuid.UUID]*opaque.Server),
	}
}

func (s *Service) newOpaqueServer() (*opaque.Server, error) {
	conf := opaque.DefaultConfiguration()
	server, err := conf.Server()
	if err != nil {
		return nil, err
	}

	if err := server.SetKeyMaterial(s.opaque.serverID, s.opaque.privateKey, s.opaque.publicKey, s.opaque.oprfSeed); err != nil {
		return nil, err
	}

	return server, nil
}

func (s *Service) setOpaqueLoginServer(loginID uuid.UUID, server *opaque.Server) error {
	s.opaque.servers[loginID] = server
	return nil
}

func (s *Service) getOpaqueLoginServer(loginID uuid.UUID) *opaque.Server {
	server, ok := s.opaque.servers[loginID]
	if !ok {
		return nil
	}
	return server
}

func (s *Service) deleteOpaqueLoginServer(loginID uuid.UUID) {
	delete(s.opaque.servers, loginID)
}
