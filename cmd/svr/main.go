package main

import (
	"bufio"
	"net"
	"strings"

	"gochat/pkg/constants"
)

var (
	// conns is connected clients
	conns map[net.Conn]struct{}
)

func main() {
	println("START", "ServerAddr:", constants.ServerAddr)
	conns = make(map[net.Conn]struct{})

	l, err := createListener()
	if err != nil {
		panic(err)
	}

	// waiting new connection from client
	for {
		conn, err := acceptConnection(l)
		if err != nil {
			panic(err)
		}
		conns[conn] = struct{}{}

		go func() {
			err := doLoop(conn)
			if err != nil {
				panic(err)
			}
		}()
	}
}

// doLoop is main loop for connected client
func doLoop(conn net.Conn) error {
	// broadcast joined message
	err := broadcast(conn, "JOINED")
	if err != nil {
		return err
	}

	for {
		msg, err := readMessage(conn)
		if err != nil {
			// break loop if message is EOF
			if err.Error() == "EOF" {
				println(addr(conn), "EOF")
				delete(conns, conn)
				// broadcast leaved message
				err = broadcast(conn, "LEAVED")
				if err != nil {
					return err
				}
				break
			}
			return err
		}

		// broadcast read message
		err = broadcast(conn, msg)
		if err != nil {
			return err
		}
	}
	return nil
}

// createListener creates net.Listener
func createListener() (net.Listener, error) {
	l, err := net.Listen("tcp", constants.ServerAddr)
	if err != nil {
		return nil, err
	}
	return l, nil
}

// acceptConnection creates net.Conn
func acceptConnection(l net.Listener) (net.Conn, error) {
	conn, err := l.Accept()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// readMessage reads message from client
func readMessage(conn net.Conn) (string, error) {
	println(addr(conn), "waiting")
	message, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", err
	}
	msg := strings.TrimRight(message, "\n")
	println(addr(conn), "<<", msg)
	return msg, nil
}

// broadcast sends message to all connected clients (expects myself)
func broadcast(myConn net.Conn, msg string) error {
	for conn, _ := range conns {
		if conn == myConn {
			continue
		}

		_, err := conn.Write([]byte(addr(myConn) + " >> " + msg + "\n"))
		if err != nil {
			return err
		}
		println(addr(myConn), ">>", msg)
	}
	return nil
}

// addr gets conn.RemoteAddr().String()
func addr(conn net.Conn) string {
	return conn.RemoteAddr().String()
}
