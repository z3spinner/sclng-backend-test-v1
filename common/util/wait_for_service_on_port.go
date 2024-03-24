package util

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

// detect - does the server appear to be online?
func detect(addr string) bool {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return false
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			logrus.Errorf("error closing connection: %v", err)
		}
	}(conn)
	return true
}

// WaitForServiceOnPort - wait for a service on a port
func WaitForServiceOnPort(
	log logrus.FieldLogger, quit <-chan struct{}, serviceName, hostPort string, timeOut time.Duration,
) error {
	snoozeTime := time.Second * 1
	maxAttempts := int(timeOut / snoozeTime)
	attemptNum := 0
	for !detect(hostPort) {
		select {
		case <-quit:
			log.Infof("quitting")
			return errors.New("quitting")
		default:
			if attemptNum > maxAttempts {
				return fmt.Errorf("could not detect %s @ %s after %d attempts", serviceName, hostPort, maxAttempts)
			}
			log.Infof(
				"could not detect %s @ %s. sleeping for %s %d/%d before retrying", serviceName, hostPort, snoozeTime,
				attemptNum, maxAttempts,
			)
			time.Sleep(snoozeTime)
			attemptNum++
		}
	}
	log.Infof("detected %s @ %s", serviceName, hostPort)
	return nil
}
