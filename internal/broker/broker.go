package broker

import (
	"bufio"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nlewo/comin/pkg/protobuf"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Broker struct {
	stopCh    chan struct{}
	publishCh chan *protobuf.Event
	subCh     chan chan *protobuf.Event
	unsubCh   chan chan *protobuf.Event
}

func New() *Broker {
	return &Broker{
		stopCh:    make(chan struct{}),
		publishCh: make(chan *protobuf.Event, 1),
		subCh:     make(chan chan *protobuf.Event, 1),
		unsubCh:   make(chan chan *protobuf.Event, 1),
	}
}

func (b *Broker) Start() {
	go func() {
		subs := map[chan *protobuf.Event]struct{}{}
		for {
			select {
			case <-b.stopCh:
				return
			case msgCh := <-b.subCh:
				subs[msgCh] = struct{}{}
			case msgCh := <-b.unsubCh:
				delete(subs, msgCh)
			case msg := <-b.publishCh:
				for msgCh := range subs {
					// msgCh is buffered, use non-blocking send to protect the broker:
					select {
					case msgCh <- msg:
					default:
					}
				}
			}
		}
	}()
}

func (b *Broker) Stop() {
	close(b.stopCh)
}

func (b *Broker) Subscribe() chan *protobuf.Event {
	msgCh := make(chan *protobuf.Event, 5)
	b.subCh <- msgCh
	return msgCh
}

func (b *Broker) Unsubscribe(msgCh chan *protobuf.Event) {
	b.unsubCh <- msgCh
}

func (b *Broker) Publish(msg *protobuf.Event) {
	b.publishCh <- msg
}

// GetLogger return a stdout et stderr writers. They have to be closed by the caller.
func (b *Broker) GetLogger(obj, objUuid string) (stdout, stderr io.WriteCloser) {
	stdoutR, stdoutW := io.Pipe()
	stderrR, stderrW := io.Pipe()
	uuid := uuid.New().String()
	b.Publish(
		&protobuf.Event{
			Type: &protobuf.Event_Log_{
				Log: &protobuf.Event_Log{
					Type: &protobuf.Event_Log_Open_{
						Open: &protobuf.Event_Log_Open{
							Uuid: uuid,
						},
					},
				},
			},
			CreatedAt: timestamppb.New(time.Now().UTC()),
		},
	)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		logrus.Debug("broker: starting to scan stdout")
		scanner := bufio.NewScanner(stdoutR)
		for scanner.Scan() {
			line := scanner.Text()
			b.Publish(
				&protobuf.Event{
					Type: &protobuf.Event_Log_{
						Log: &protobuf.Event_Log{
							Type: &protobuf.Event_Log_Line_{
								Line: &protobuf.Event_Log_Line{
									Uuid:       uuid,
									Source:     "stdout",
									Purpose:    obj,
									ObjectUuid: objUuid,
									Msg:        line,
								},
							},
						},
					},
					CreatedAt: timestamppb.New(time.Now().UTC()),
				},
			)
		}
		logrus.Debug("broken: stdout/" + obj + "/" + objUuid + ": stdout closed")
	}()

	go func() {
		defer wg.Done()
		logrus.Debug("broker: starting to scan stderr")
		scanner := bufio.NewScanner(stderrR)
		for scanner.Scan() {
			line := scanner.Text()
			b.Publish(
				&protobuf.Event{
					Type: &protobuf.Event_Log_{
						Log: &protobuf.Event_Log{
							Type: &protobuf.Event_Log_Line_{
								Line: &protobuf.Event_Log_Line{
									Uuid:       uuid,
									Source:     "stderr",
									Purpose:    obj,
									ObjectUuid: objUuid,
									Msg:        line,
								},
							},
						},
					},
					CreatedAt: timestamppb.New(time.Now().UTC()),
				},
			)
		}
		logrus.Debug("broken: stdout/" + obj + "/" + objUuid + ": stdout closed")
	}()

	go func() {
		wg.Wait()
		b.Publish(
			&protobuf.Event{
				Type: &protobuf.Event_Log_{
					Log: &protobuf.Event_Log{
						Type: &protobuf.Event_Log_Close_{
							Close: &protobuf.Event_Log_Close{
								Uuid: uuid,
							},
						},
					},
				},
				CreatedAt: timestamppb.New(time.Now().UTC()),
			},
		)
	}()

	return stdoutW, stderrW
}
