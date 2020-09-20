package h2server

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

type (
	Server struct {
		logger  Logger
		cert    tls.Certificate
		addr    string
		preface func() *SettingsFrameBuilder
		mp      func(Conn) Multiplexer
	}

	ServerConfig struct {
		Logger      Logger
		Certificate tls.Certificate
		Address     string
		Preface     func() *SettingsFrameBuilder
		Multiplexer func(Conn) Multiplexer
	}
)

const (
	alpnH2 = "h2"
)

var (
	expectedClientPreface = []byte("PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n")
)

func NewServer(config *ServerConfig) *Server {
	logger := config.Logger
	if logger == nil {
		logger = NullLogger()
	}

	return &Server{
		logger:  config.Logger,
		cert:    config.Certificate,
		addr:    config.Address,
		preface: config.Preface,
		mp:      config.Multiplexer,
	}
}

func (sv *Server) ListenAndServe() error {
	ln, err := tls.Listen("tcp", sv.addr, &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{sv.cert},
		NextProtos:   []string{alpnH2},
	})
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			sv.logger.Write(ErrorLog, "failed to accept connection: %s\n", err.Error())
			continue
		}

		go func() {
			if err := sv.handleConn(conn.(*tls.Conn)); err != nil {
				sv.connLog(conn, ErrorLog, err.Error())
			}
		}()
	}
}

func (sv *Server) handleConn(conn *tls.Conn) error {
	defer func() {
		conn.Close()
	}()

	sv.connLog(conn, DebugLog, "connected")

	// Handshake
	if err := conn.Handshake(); err != nil {
		return fmt.Errorf("failed to handshake: %w", err)
	}

	negotiated := conn.ConnectionState().NegotiatedProtocol
	if negotiated != alpnH2 {
		return fmt.Errorf("invalid negotiated protocol: %s", negotiated)
	}

	sv.connLog(conn, DebugLog, "handshake completed")

	// Exchange server and client preface
	preface, err := sv.preface().Build()
	if err != nil {
		return fmt.Errorf("failed to generate server preface: %w", err)
	}
	if _, err := preface.WriteTo(conn); err != nil {
		return err
	}

	clientPreface := make([]byte, len(expectedClientPreface))
	if _, err := conn.Read(clientPreface); err != nil {
		return err
	}

	if bytes.Compare(clientPreface, expectedClientPreface) != 0 {
		return errors.New("read invalid client preface")
	}

	sv.connLog(conn, DebugLog, "accept client preface")

	// Start reader and writer
	pseudoConn := newPseudoConn(conn)
	ctx, cancel := context.WithCancel(contextWithConn(context.Background(), pseudoConn))

	var rErr, wErr error
	cancelOnce := &sync.Once{}
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		wErr = sv.runWriter(ctx, conn)
		cancelOnce.Do(cancel)
		wg.Done()
	}()

	mp := sv.mp(pseudoConn)

	go func() {
		rErr = sv.runReader(ctx, conn, mp)
		cancelOnce.Do(cancel)
		wg.Done()
	}()

	wg.Wait()

	// TODO: Handle rErr and wErr
	mp.Terminated()

	return nil
}

func (sv *Server) connLog(conn net.Conn, level LogLevel, format string, args ...interface{}) {
	sv.logger.Write(level, fmt.Sprintf("<%s> ", conn.RemoteAddr())+format, args...)
}

func (sv *Server) runReader(ctx context.Context, conn net.Conn, mp Multiplexer) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		f, err := Read(conn)
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			return err
		}

		mp.Received(f)
	}
}

func (*Server) runWriter(ctx context.Context, conn net.Conn) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case outgoing, ok := <-connFromContext(ctx).writeCh():
			if !ok {
				return nil
			}

			if _, err := outgoing.WriteTo(conn); err != nil {
				return fmt.Errorf("failed to send frame: %w", err)
			}
		}
	}
}
