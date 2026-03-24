package pcap

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"time"

	"shadowStack/internal/protocol"
	"shadowStack/internal/ui"
)

func RunProxy(
	localPort string,
	remoteAddr string,
	eventsChan chan<- ui.QueryMsg,
	factory *protocol.ParserFactory,
) error {
	listener, err := net.Listen("tcp", "0.0.0.0:"+localPort)
	if err != nil {
		return fmt.Errorf("failed to start proxy listener: %v", err)
	}

	log.Printf("📡 Proxy active: 0.0.0.0:%s -> %s", localPort, remoteAddr)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			continue
		}

		go func(client net.Conn) {
			defer client.Close()
			var remote net.Conn
			var dialErr error

			remote, dialErr = tls.Dial("tcp", remoteAddr, &tls.Config{InsecureSkipVerify: true})
			if dialErr != nil {
				remote, dialErr = net.Dial("tcp", remoteAddr)
			}
			if dialErr != nil {
				return
			}
			defer remote.Close()

			// Create a unique ID for this connection to accurately match requests and responses
			clientPort := uint16(client.RemoteAddr().(*net.TCPAddr).Port)

			// Channel to hold the start time of the query
			inFlight := make(chan time.Time, 1)

			// 1. Sniff Client -> Server (Requests)
			go func() {
				buf := make([]byte, 4096)
				for {
					n, err := client.Read(buf)
					if err != nil || n == 0 {
						return
					}

					_, portStr, _ := net.SplitHostPort(remoteAddr)
					var dbPort uint16
					fmt.Sscanf(portStr, "%d", &dbPort)

					if parser := factory.GetParser(dbPort); parser != nil {
						query, err := parser.Parse(buf[:n])

						// Only send to UI if it wasn't ignored (e.g., not a heartbeat)
						if err == nil {
							inFlight <- time.Now() // Start the timer
							eventsChan <- ui.QueryMsg{
								Comm:   "proxy-tap",
								DBType: parser.Name(),
								Query:  query,
								Port:   clientPort,
							}
						}
					}
					remote.Write(buf[:n])
				}
			}()

			// 2. Sniff Server -> Client (Responses)
			buf := make([]byte, 4096)
			for {
				n, err := remote.Read(buf)
				if err != nil || n == 0 {
					break
				}

				// If we have a pending query, calculate the latency
				select {
				case startTime := <-inFlight:
					eventsChan <- ui.QueryMsg{
						IsUpdate: true,
						Port:     clientPort,
						Latency:  time.Since(startTime),
					}
				default:
					// No pending query, just pass data through
				}

				client.Write(buf[:n])
			}
		}(clientConn)
	}
}
