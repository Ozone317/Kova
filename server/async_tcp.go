// =============================================================================
// HOW THIS ASYNC SERVER WORKS — A BLOCK BY BLOCK BREAKDOWN
// =============================================================================
//
// BACKGROUND: WHY NOT JUST USE GOROUTINES?
// -----------------------------------------
// The sync implementation spawned one goroutine per client connection.
// Each goroutine blocked on conn.Read(), waiting for the client to send data.
// With 10,000 clients, you have 10,000 goroutines all sleeping and burning
// memory. This is the C10K problem.
//
// The solution is epoll; a Linux kernel feature that lets a SINGLE goroutine
// watch thousands of file descriptors and only wake up when one is actually
// ready. Zero goroutines sleeping. Zero wasted memory.
//
//
// WHAT IS A FILE DESCRIPTOR (FD)?
// --------------------------------
// In Linux, everything is a file; network sockets, disk files, pipes, timers.
// A file descriptor is just an integer the kernel gives you as a handle to one
// of these resources. When you create a socket, the kernel returns fd=3.
// From then on, you refer to that socket by the number 3. The kernel maintains
// an internal table mapping these integers to the actual resources.
//
//   Process FD table:
//     0 → stdin
//     1 → stdout
//     2 → stderr
//     3 → server socket  (assigned by the kernel when we call Socket())
//     4 → client 1       (assigned by the kernel when we call Accept())
//     5 → client 2
//     ...
//
//
// BLOCK 1 - CREATING THE SERVER SOCKET
// --------------------------------------
//
//   serverFD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
//
// syscall.Socket() asks the kernel to create a socket and return its fd.
// The three arguments tell the kernel what kind of socket to create:
//
//   AF_INET      - use IPv4 (AF_INET6 = IPv6, AF_UNIX = local/process sockets)
//   SOCK_STREAM  - use TCP, which keeps the connection alive across multiple
//                  reads and writes. The alternative, SOCK_DGRAM, is UDP which
//                  fires a packet and forgets the connection immediately.
//   0            - use the default protocol for this socket type, which is TCP
//                  for SOCK_STREAM.
//
// The kernel returns an integer like 3. That is serverFD. Every subsequent
// operation on this socket uses that number.
//
//
// BLOCK 2 - NON-BLOCKING MODE
// ----------------------------
//
//   syscall.SetNonblock(serverFD, true)
//
// By default, socket operations BLOCK - if you call read() and no data has
// arrived yet, your goroutine freezes and waits. Non-blocking mode flips that
// behaviour: every operation returns immediately. If there is no data ready,
// you get the error EAGAIN instead of a freeze.
//
// This is essential for the event loop. You never want the loop to get stuck
// waiting on one client while thousands of others are ready to be served.
//
//
// BLOCK 3 - BIND AND LISTEN
// --------------------------
//
//   syscall.Bind(serverFD, &syscall.SockaddrInet4{Port: config.PORT, ...})
//   syscall.Listen(serverFD, max_clients)
//
// Bind answers the question "which address and port does this socket own?"
// Without it, the socket exists in the kernel but is not attached to any
// address - no client could reach it. We are telling the kernel: this fd
// owns port 7379 on this IP.
//
// Listen transitions the socket from "just a socket" to "a socket that accepts
// incoming connections". The second argument is the backlog - how many pending
// connections the kernel will queue up while our code is busy processing other
// events. If the queue fills up, new connection attempts are dropped.
//
//
// BLOCK 4 - CREATING THE EPOLL INSTANCE
// ---------------------------------------
//
//   epollFD, err := syscall.EpollCreate1(0)
//
// This creates the epoll instance. Think of it as creating an empty watchlist.
// The kernel allocates an internal data structure to track which fds you want
// monitored, and hands you back another file descriptor (epollFD) to refer to
// it. Notice that epoll itself is represented as an fd - in Linux, everything
// is a file, including the thing that watches other files.
//
// The 0 argument means no special flags. EPOLL_CLOEXEC is the only other
// option - it auto-closes the epoll fd if the process forks, which we do
// not need here.
//
//
// BLOCK 5 - REGISTERING THE SERVER SOCKET WITH EPOLL
// ----------------------------------------------------
//
//   var socketServerEvent syscall.EpollEvent = syscall.EpollEvent{
//       Events: syscall.EPOLLIN,
//       Fd:     int32(serverFD),
//   }
//   syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, serverFD, &socketServerEvent)
//
// EpollCtl is the control function for epoll. It has three modes:
//   EPOLL_CTL_ADD - add a new fd to the watchlist
//   EPOLL_CTL_MOD - change what events to watch on an existing fd
//   EPOLL_CTL_DEL - stop watching an fd entirely
//
// The EpollEvent struct has two fields:
//
//   Events: EPOLLIN - watch for incoming data (read-ready). For the server
//                     socket specifically, EPOLLIN fires when a new client
//                     has completed the TCP handshake and is waiting to be
//                     accepted. For client sockets, it fires when the client
//                     has sent bytes we can read.
//
//   Fd: int32(serverFD) — this is the tag we will read back when epoll wakes
//                         us up. When the kernel fires an event, it hands back
//                         this same EpollEvent struct, so we need the fd stored
//                         inside it to know which socket triggered the event.
//
//
// BLOCK 6 - THE EVENTS BUFFER
// ----------------------------
//
//   var events []syscall.EpollEvent = make([]syscall.EpollEvent, max_clients)
//
// This is a pre-allocated slice that EpollWait fills in on every iteration.
// When the kernel wakes the loop, it writes the ready events into this slice.
// Allocating max_clients slots means that even under peak load, a single
// EpollWait call can return all ready events at once rather than requiring
// multiple calls.
//
//
// BLOCK 7 - THE EVENT LOOP CORE
// ------------------------------
//
//   nevents, err := syscall.EpollWait(epollFD, events[:], -1)
//
// This is the most important line in the file. EpollWait puts the process to
// sleep until at least one watched fd is ready. The -1 timeout means sleep
// forever - do not wake up until something actually happens. Zero CPU is
// consumed while waiting.
//
// When the kernel detects activity on any watched fd, it wakes the process and
// fills events[0..nevents-1] with exactly the fds that are ready. nevents
// tells you how many came back. Everything beyond nevents in the slice is stale
// data from a prior iteration and must be ignored.
//
//
// BLOCK 8 - DISPATCHING ON WHICH FD FIRED
// -----------------------------------------
//
//   if int(events[i].Fd) == serverFD { ... } else { ... }
//
// For each ready event, we ask: is this the server socket or a client socket?
//
// IF IT IS THE SERVER SOCKET:
//   EPOLLIN on the server fd does not mean data arrived. It means a new TCP
//   handshake completed and a client is waiting to be accepted. We call
//   Accept() to get the client fd, set it non-blocking, then register it with
//   epoll so future reads from that client will also wake the loop.
//
// IF IT IS A CLIENT SOCKET:
//   EPOLLIN on a client fd means that client sent bytes. We read them, decode
//   the RESP frame, execute the command, and write the response. If reading
//   returns an error, the client disconnected - we close the fd, remove it
//   from epoll, and decrement the client counter.
//
//
// THE FULL FLOW IN ONE PICTURE
// -----------------------------
//
//   startup
//     │
//     ├─ socket() ──────────────────── create server fd
//     ├─ setnonblock(serverFD)
//     ├─ bind() + listen() ─────────── attach to port 7379
//     ├─ epoll_create() ────────────── create empty watchlist
//     └─ epoll_ctl(ADD, serverFD) ──── watch for new connections
//           │
//           ▼
//        epoll_wait()  ◄──────────────────────────────────────┐
//           │   (sleeps here, zero CPU, until something fires) │
//           │                                                   │
//     ┌─────┴──────┐                                           │
//     │            │                                           │
//  server fd    client fd                                      │
//  is ready     is ready                                       │
//     │            │                                           │
//  accept()     read() → decode RESP → run command → write()  │
//  setnonblock  (on error: close fd, epoll_ctl DEL)            │
//  epoll_ctl                                                   │
//  (ADD client)                                                │
//     │            │                                           │
//     └─────┬──────┘                                           │
//           └───────────────────────────────────────────────────┘
//
// =============================================================================

package server

import (
	"log"
	"net"
	"syscall"
	"time"

	"github.com/Ozone317/Kova/config"
	"github.com/Ozone317/Kova/core"
)

var concurrent_clients int = 0
var cleanup_interval time.Duration = 1 * time.Second
var last_cleanup_time time.Time = time.Now()

func RunAsyncTCPServer() error {
	log.Println("starting an asynchronous TCP server on port 7379")

	max_clients := 20000

	// Create a socket (raw socket as we need access to file descriptors)
	// SOCK_STREAM tells that the TCP connection should not be disconnected as soon as a reply is sent
	// Also, this serverFD is what we want monitored by epoll
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}

	defer syscall.Close(serverFD)

	// Set the socket to operate in non-blocking mode
	err = syscall.SetNonblock(serverFD, true)
	if err != nil {
		return err
	}

	// Bind the socket to an address and port
	ip4 := net.ParseIP(config.HOST).To4()
	err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.PORT,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	})

	if err != nil {
		return err
	}

	// Start listening
	err = syscall.Listen(serverFD, max_clients)
	if err != nil {
		return err
	}

	// AsyncIO starts here

	// Create an epoll instance
	epollFD, err := syscall.EpollCreate1(0)
	if err != nil {
		log.Fatal(err)
	}
	defer syscall.Close(epollFD)

	// Specify the events we want to get hints about
	// and set the server socket to be monitored
	var socketServerEvent syscall.EpollEvent = syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(serverFD),
	}

	// Register the server socket with epoll
	err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, serverFD, &socketServerEvent)
	if err != nil {
		return err
	}

	// Create EPOLL Event Objects to hold events. This will be holding file descriptors that are ready for IO by epoll system call
	var events []syscall.EpollEvent = make([]syscall.EpollEvent, max_clients)

	for {
		if time.Since(last_cleanup_time) >= cleanup_interval {
			core.DeleteExpiredKeys()
			last_cleanup_time = time.Now()
		}

		// see if any FD is ready for IO
		nevents, err := syscall.EpollWait(epollFD, events[:], -1)
		if err != nil {
			continue
		}

		for i := 0; i < nevents; i++ {
			// if the socket server itself is ready for an IO
			if int(events[i].Fd) == serverFD {
				// accept the connection
				clientFD, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Println("err", err)
					continue
				}

				// increase the number of concurrent clients
				concurrent_clients++
				log.Println("New client connected: ", clientFD, "Total connected clients: ", concurrent_clients)

				// set the client socket to be non-blocking
				err = syscall.SetNonblock(clientFD, true)
				if err != nil {
					log.Fatal(err)
				}

				// add this new client-to-server TCP connection to be monitored
				var socketClientEvent syscall.EpollEvent = syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(clientFD),
				}

				err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, clientFD, &socketClientEvent)
				if err != nil {
					log.Fatal(err)
				}

			} else {
				// read the command as some client is sending a command
				comm := core.FDComm{Fd: int(events[i].Fd)}
				cmd, err := readCommand(comm)
				if err != nil {
					syscall.Close(int(events[i].Fd))
					concurrent_clients--
					log.Println("Client disconnected: ", events[i].Fd, "Total connected clients: ", concurrent_clients)
					continue
				}
				respond(cmd, comm)
			}
		}
	}
}
