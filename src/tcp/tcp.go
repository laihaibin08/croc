package tcp

import (
	"net"
	"sync"
	"time"

	log "github.com/cihub/seelog"
	"github.com/pkg/errors"
	"github.com/schollz/croc/src/comm"
	"github.com/schollz/croc/src/logger"
)

const TCP_BUFFER_SIZE = 1024 * 64

type roomInfo struct {
	first  comm.Comm
	second comm.Comm
	opened time.Time
	full   bool
}

type roomMap struct {
	rooms map[string]roomInfo
	sync.Mutex
}

var rooms roomMap

// Run starts a tcp listener, run async
func Run(debugLevel, port string) {
	logger.SetLogLevel(debugLevel)
	rooms.Lock()
	rooms.rooms = make(map[string]roomInfo)
	rooms.Unlock()

	// delete old rooms
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			rooms.Lock()
			for room := range rooms.rooms {
				if time.Since(rooms.rooms[room].opened) > 3*time.Hour {
					delete(rooms.rooms, room)
				}
			}
			rooms.Unlock()
		}
	}()

	err := run(port)
	if err != nil {
		log.Error(err)
	}
}

func run(port string) (err error) {
	log.Debugf("starting TCP server on tcp: 0.0.0.0:" + port)
	server, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Error(err)
		return errors.Wrap(err, "Error listening on :"+port)
	}
	defer server.Close()
	// spawn a new goroutine whenever a client connects
	for {
		log.Debugf("waiting for connection")
		connection, err := server.Accept()
		if err != nil {
			err = errors.Wrap(err, "problem accepting connection")
			log.Error(err)
			return err
		}
		log.Debugf("client %s connected", connection.RemoteAddr().String())
		go func(port string, connection net.Conn) {
			errCommunication := clientCommuncation(port, comm.New(connection))
			if errCommunication != nil {
				log.Warnf("relay-%s: %s", connection.RemoteAddr().String(), errCommunication.Error())
			}
		}(port, connection)
	}
}

func clientCommuncation(port string, c comm.Comm) (err error) {
	// send ok to tell client they are connected
	log.Debug("sending ok")
	err = c.Send([]byte("ok"))
	if err != nil {
		return
	}

	// wait for client to tell me which room they want
	log.Debug("waiting for answer")
	roomBytes, err := c.Receive()
	if err != nil {
		return
	}
	room := string(roomBytes)

	rooms.Lock()
	// create the room if it is new
	if _, ok := rooms.rooms[room]; !ok {
		rooms.rooms[room] = roomInfo{
			first:  c,
			opened: time.Now(),
		}
		rooms.Unlock()
		// tell the client that they got the room
		err = c.Send([]byte("ok"))
		if err != nil {
			log.Error(err)
			return
		}
		log.Debugf("room %s has 1", room)
		return nil
	}
	if rooms.rooms[room].full {
		rooms.Unlock()
		err = c.Send([]byte("room full"))
		if err != nil {
			log.Error(err)
			return
		}
		return nil
	}
	log.Debugf("room %s has 2", room)
	rooms.rooms[room] = roomInfo{
		first:  rooms.rooms[room].first,
		second: c,
		opened: rooms.rooms[room].opened,
		full:   true,
	}
	otherConnection := rooms.rooms[room].first
	rooms.Unlock()

	// second connection is the sender, time to staple connections
	var wg sync.WaitGroup
	wg.Add(1)

	// start piping
	go func(com1, com2 comm.Comm, wg *sync.WaitGroup) {
		log.Debug("starting pipes")
		pipe(com1.Connection(), com2.Connection())
		wg.Done()
		log.Debug("done piping")
	}(otherConnection, c, &wg)

	// tell the sender everything is ready
	err = c.Send([]byte("ok"))
	if err != nil {
		return
	}
	wg.Wait()

	// delete room
	rooms.Lock()
	log.Debugf("deleting room: %s", room)
	rooms.rooms[room].first.Close()
	rooms.rooms[room].second.Close()
	delete(rooms.rooms, room)
	rooms.Unlock()
	return nil
}

// chanFromConn creates a channel from a Conn object, and sends everything it
//  Read()s from the socket to the channel.
func chanFromConn(conn net.Conn) chan []byte {
	c := make(chan []byte)

	go func() {
		b := make([]byte, TCP_BUFFER_SIZE)

		for {
			n, err := conn.Read(b)
			if n > 0 {
				res := make([]byte, n)
				// Copy the buffer so it doesn't get changed while read by the recipient.
				copy(res, b[:n])
				c <- res
			}
			if err != nil {
				log.Debug(err)
				c <- nil
				break
			}
		}
	}()

	return c
}

// pipe creates a full-duplex pipe between the two sockets and
// transfers data from one to the other.
func pipe(conn1 net.Conn, conn2 net.Conn) {
	chan1 := chanFromConn(conn1)
	chan2 := chanFromConn(conn2)

	for {
		select {
		case b1 := <-chan1:
			if b1 == nil {
				return
			}
			conn2.Write(b1)

		case b2 := <-chan2:
			if b2 == nil {
				return
			}
			conn1.Write(b2)
		}
	}
}
