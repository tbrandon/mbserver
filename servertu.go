package mbserver

import (
	"errors"
	"io"
	"log"
	"syscall"
	"time"

	"github.com/goburrow/serial"
)

type ListenRTUOption func(options *rtuOptions)
type rtuOptions struct {
	notifyExit chan error
}

func NotifyExit(ch chan error) ListenRTUOption {
	return func(s *rtuOptions) {
		s.notifyExit = ch
	}
}

// ListenRTU starts the Modbus server listening to a serial device.
// For example:  err := s.ListenRTU(&serial.Config{Address: "/dev/ttyUSB0"})
func (s *Server) ListenRTU(serialConfig *serial.Config, options ...ListenRTUOption) (err error) {
	var opts rtuOptions
	for _, option := range options {
		option(&opts)
	}
	port, err := serial.Open(serialConfig)
	if err != nil {
		log.Fatalf("failed to open %s: %v\n", serialConfig.Address, err)
	}
	s.ports = append(s.ports, port)

	s.portsWG.Add(1)
	go func() {
		defer s.portsWG.Done()
		err := s.acceptSerialRequests(port)
		if opts.notifyExit != nil {
			opts.notifyExit <- err
		}
	}()

	return err
}

func (s *Server) acceptSerialRequests(port serial.Port) error {
SkipFrameError:
	for {
		select {
		case <-s.portsCloseChan:
			return nil
		default:
		}

		buffer := make([]byte, 512)

		bytesRead, err := port.Read(buffer)
		if errors.Is(err, serial.ErrTimeout) {
			continue
		}
		if err != nil {
			if errors.Is(err, serial.ErrTimeout) || errors.Is(err, syscall.EWOULDBLOCK) {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			if err != io.EOF {
				log.Printf("serial read error %v\n", err)
			}
			return err
		}

		if bytesRead != 0 {

			// Set the length of the packet to the number of read bytes.
			packet := buffer[:bytesRead]

			frame, err := NewRTUFrame(packet)
			if err != nil {
				log.Printf("bad serial frame error %v\n", err)
				//The next line prevents RTU server from exiting when it receives a bad frame. Simply discard the erroneous 
				//frame and wait for next frame by jumping back to the beginning of the 'for' loop.
				log.Printf("Keep the RTU server running!!\n")
				continue SkipFrameError
				//return
			}

			request := &Request{port, frame}

			s.requestChan <- request
		}
	}
}
