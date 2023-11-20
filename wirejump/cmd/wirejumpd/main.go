package main

import (
	"context"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"wirejump/internal/ipc"
	"wirejump/internal/state"
	"wirejump/internal/version"
)

func startServer(ctx context.Context) error {
	if err := os.RemoveAll(ipc.SocketFile); err != nil {
		return fmt.Errorf("failed to clean socket file: %s", err)
	}

	// Create server manually, since
	// listener is also managed explicitly
	server := rpc.NewServer()
	wjrpc := new(IpcWrappedHandler)

	// Register RPC
	if err := server.Register(wjrpc); err != nil {
		return fmt.Errorf("failed to register RPC: %s", err)
	}

	// Open socket
	listener, err := net.Listen("unix", ipc.SocketFile)

	if err != nil {
		return err
	}

	// Chmod socket: everyone from wirejump group should have access
	if err := os.Chmod(ipc.SocketFile, 0660); err != nil {
		return fmt.Errorf("failed to chmod socket: %s", err)
	}

	// Close socket and remove it on return
	defer listener.Close()
	defer os.Remove(ipc.SocketFile)

	failed := make(chan error, 1)
	incoming := make(chan net.Conn, 1)

	// Start listener goroutine
	go func() {
		for {
			conn, err := listener.Accept()

			if err != nil {
				failed <- err
			}
			incoming <- conn
		}
	}()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Stopping WireJump server...")

			return nil
		case err := <-failed:
			// Accept has failed
			return err
		case conn := <-incoming:
			// Someone has connected
			go server.ServeConn(conn)
		}
	}
}

func main() {
	// Handle default args and get config file path
	configPath := HandleProgramArgs()

	// Get configuration
	configState, err := ParseConfig(configPath)

	if err != nil {
		ErrorExit("Invalid config file: ", err)
	}

	// Create initial app state
	applicationState := state.GetStateInstance()

	// Update app state according to the configuration
	if err := UpdateAppState(applicationState, &configState); err != nil {
		ErrorExit("Failed to update app state:", err)
	}

	// Register signal context
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	fmt.Println("Starting WireJump server...")
	fmt.Println(version.VersionString())

	if err := startServer(ctx); err != nil {
		ErrorExit(err)
	}
}
