package abm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"
)

// Data Logging process local to the model instance.
func (m *Model) log(ec chan<- error) {
	fmt.Println("starting logging...")
	time.Sleep(pause)
	_ = "breakpoint" // godebug

	signature := "LOG_" + m.SessionIdentifier
	turn, clash := m.turnSignal.Register(signature)
	if clash {
		errStr := "Clash when registering Model: " + m.SessionIdentifier + " log: for sync with m.turnSignal"
		ec <- errors.New(errStr)
		return
	}

	defer func() {
		m.turnSignal.Deregister(signature)
		// Need to wipe the agent records too? -yes, probably... but should be in Stop()?
	}()

	if m.UseCustomLogPath {
		m.LogPath = path.Join(os.Getenv("HOME")+os.Getenv("HOMEPATH"), m.CustomLogPath, abmlogPath, m.SessionIdentifier, m.timestamp)
	} else {
		m.LogPath = path.Join(os.Getenv("HOME")+os.Getenv("HOMEPATH"), abmlogPath, m.SessionIdentifier, m.timestamp)
	}

	for {
		select {
		case <-m.rc: // run finished as rc channel closed!
			time.Sleep(time.Second)
			// clean up?
			return
		case <-turn:
			func() {
				reccpp := m.cppRecordCopy()
				recvp := m.vpRecordCopy()
				go func(rc map[string]ColourPolymorphicPrey, errCh chan<- error) {
					// write map as json to file.
					tc := fmt.Sprintf("%08v", m.Turn)
					dir := m.LogPath
					path := dir + string(filepath.Separator) + tc + "_cpp_pop_record.dat"

					msg, err := json.MarshalIndent(rc, "", "  ")
					if err != nil {
						log.Printf("model: logging: json.Marshal failed, error: %v\n source: %s : %s : %v\n", err, m.SessionIdentifier, m.timestamp, m.Turn)
						errCh <- err
						return
					}

					var buff []byte
					out := bytes.NewBuffer(buff)
					out.Write(msg)
					output := make([]byte, 1024*10)
					n, rerr := out.Read(output)
					if n == 0 || rerr != nil {
						fmt.Println("n:", n, "rerr:", rerr.Error())
						errCh <- err
						return
					}

					err = os.MkdirAll(dir, 0777)
					if err != nil {
						errCh <- err
						return
					}
					err = ioutil.WriteFile(path, output, 0777)
					if err != nil {
						errCh <- err
						return
					}
				}(reccpp, ec)
				go func(rv map[string]VisualPredator, errCh chan<- error) {
					// write map as json to file.
				}(recvp, ec)
			}()
		}
	}
}