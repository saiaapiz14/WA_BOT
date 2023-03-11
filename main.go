package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

var client *whatsmeow.Client
var recipientNumbers = []string{"601160564476@s.whatsapp.net", "60122412026@s.whatsapp.net"} // List of recipient numbers

func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		if !v.Info.IsFromMe {
			for _, num := range recipientNumbers { //scan through recipient number list
				if v.Message.GetConversation() == "/admin" { //check if the message use admin keyword
					if v.Info.Sender.String() == num { // Check if the incoming message is from one of the recipient numbers
						fmt.Println("PESAN DITERIMA DARIPADA ADMIN!", v.Message.GetConversation()) //show message in terminal that admin sent the message
						client.SendMessage(v.Info.Sender, "", &waProto.Message{                    //send message response to the keyword
							Conversation: proto.String("YA TUAN APA SAYA BOLEH BANTU"), //content of the message
						})
						break //exit after running function
					} else {
						fmt.Println("PESAN DITERIMA DARIPADA USER!", v.Message.GetConversation())
						client.SendMessage(v.Info.Sender, "", &waProto.Message{
							Conversation: proto.String("Pesan ini automatik, menggunakan GO!. Anda mengirim pesan: " + v.Message.GetConversation()),
						})
						break

					}
				} else if v.Message.GetConversation() != "" {
					if v.Info.Sender.String() != num { // Check if the incoming message is from one of the recipient numbers
						fmt.Println("PESAN DITERIMA DARIPADA USER!", v.Message.GetConversation())
						client.SendMessage(v.Info.Sender, "", &waProto.Message{
							Conversation: proto.String("Pesan ini automatik, menggunakan GO!. Anda mengirim pesan: " + v.Message.GetConversation()),
						})
						break

					}
				}
			}
		}
	}
}

func main() {
	dbLog := waLog.Stdout("Database", "DEBUG", false)
	// Make sure you add appropriate DB connector imports, e.g. github.com/mattn/go-sqlite3 for SQLite
	container, err := sqlstore.New("sqlite3", "file:wsap.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "DEBUG", false)
	client = whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(eventHandler)

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				// e.g. qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
				fmt.Println("QR code:", evt.Code)
				//qrterminal.Generate(evt.Code, qrterminal.L, os.Stdout)
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}
