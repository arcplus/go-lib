package nsq

import (
	"os"
	"testing"
	"time"
)

var lupdAddr = func() string {
	if lupd := os.Getenv("nsq_lupd"); lupd != "" {
		return lupd
	}
	return "10.241.11.8:4161"
}()

func TestPubSub(t *testing.T) {
	err := RegisterPub(lupdAddr, nil, "test")
	if err != nil {
		t.Fatal(err)
	}

	err = Publish("test", []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}

	err = PublishWithJsonExt("test", []byte("hello ext"), &MsgExt{
		DispatchTag: "tag",
		Custom:      map[string]string{"tid": "x-req-id"},
	})
	if err != nil {
		t.Fatal(err)
	}

	SubscribeHandleFunc(lupdAddr, nil, "test", "default", func(msg *Message) error {
		t.Log(string(msg.Body), string(msg.ExtBytes))
		return nil
	}, 1)

	time.Sleep(time.Second * 5)
}
