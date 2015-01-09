package junos

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/Juniper/go-netconf/netconf"
	"log"
)

// Session holds the connection information to our Junos device.
type Session struct {
	Conn *netconf.Session
}

// RollbackXML parses our rollback configuration.
type RollbackXML struct {
	XMLName xml.Name `xml:"rollback-information"`
	Config  string   `xml:"configuration-information>configuration-output"`
}

// RescueXML parses our rescue configuration.
type RescueXML struct {
	XMLName xml.Name `xml:"rescue-information"`
	Config  string   `xml:"configuration-information>configuration-output"`
}

// NewSession establishes a new connection to a Junos device that we will use
// to run our commands against.
func NewSession(host, user, password string) *Session {
	s, err := netconf.DialSSH(host, netconf.SSHConfigPassword(user, password))
	if err != nil {
		log.Fatal(err)
	}

	return &Session{
		Conn: s,
	}
}

// Lock locks the candidate configuration.
func (s *Session) Lock() error {
	lockRPC := "<rpc><lock><target><candidate/></target></lock></rpc>"
	resp, err := s.Conn.Exec(lockRPC)
	if err != nil {
        log.Fatal(err)
	}

	if resp.Ok == false {
		for _, m := range resp.Errors {
			return errors.New(m.Message)
		}
	}
    
    return nil
}

// Unlock unlocks the candidate configuration.
func (s *Session) Unlock() error {
	unlockRPC := "<rpc><unlock><target><candidate/></target></unlock></rpc>"
	resp, err := s.Conn.Exec(unlockRPC)
	if err != nil {
        log.Fatal(err)
	}

	if resp.Ok == false {
		for _, m := range resp.Errors {
			return errors.New(m.Message)
		}
	}
    
    return nil
}

// GetRollbackConfig returns the configuration of the given rollback state.
func (s *Session) GetRollbackConfig(number int) (string, error) {
	rb := &RollbackXML{}
	rpcCommand := fmt.Sprintf("<rpc><get-rollback-information><rollback>%d</rollback><format>text</format></get-rollback-information></rpc>", number)
	reply, err := s.Conn.Exec(rpcCommand)
    
    if err != nil {
        log.Fatal(err)
    }
    
	if reply.Ok == false {
		for _, m := range reply.Errors {
			return "", errors.New(m.Message)
		}
	}

	err = xml.Unmarshal([]byte(reply.Data), rb)
	if err != nil {
		log.Fatal(err)
	}

	return rb.Config, nil
}

// RollbackDiff compares the current active configuration to a given rollback configuration.
func (s *Session) RollbackDiff(compare int) (string, error) {
	rb := &RollbackXML{}
	rpcCommand := fmt.Sprintf("<rpc><get-rollback-information><rollback>0</rollback><compare>%d</compare><format>text</format></get-rollback-information></rpc>", compare)
	reply, err := s.Conn.Exec(rpcCommand)
    
    if err != nil {
        log.Fatal(err)
    }
    
	if reply.Ok == false {
		for _, m := range reply.Errors {
			return "", errors.New(m.Message)
		}
	}

	err = xml.Unmarshal([]byte(reply.Data), rb)
	if err != nil {
		log.Fatal(err)
	}

	return rb.Config, nil
}

// GetRescueConfig returns the rescue configuration.
func (s *Session) GetRescueConfig() (string, error) {
	rescue := &RescueXML{}
	command := RPCCommand["get-rescue-information"]
	reply, err := s.Conn.Exec(command)
    
    if err != nil {
        log.Fatal(err)
    }
    
	if reply.Ok == false {
		for _, m := range reply.Errors {
			return "", errors.New(m.Message)
		}
	}

	err = xml.Unmarshal([]byte(reply.Data), rescue)
	if err != nil {
		log.Fatal(err)
	}

    if rescue.Config == "" {
        return "No rescue configuration set.", nil
    }
    
	return rescue.Config, nil
}

// Close disconnects and closes the session to our Junos device.
func (s *Session) Close() {
	s.Conn.Close()
}
