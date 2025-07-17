package ingest

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/xmppo/go-xmpp"
)

const (
	exchangeOut   = "awips.exchange"
	queueOut      = "awips.queue"
	routingKeyOut = "awips.parse"
)

type XmppConfig struct {
	Server   string
	Room     string
	User     string
	Pass     string
	Resource string
}

type Server struct {
	xmpp     *xmpp.Client
	messages chan Message
	rabbit   *amqp.Connection
	out      *amqp.Channel
	ctx      context.Context
	cancel   context.CancelFunc
}

type Message struct {
	Text       string
	ReceivedAt time.Time
}

func New() (*Server, error) {
	nwwsConfig := XmppConfig{
		Server:   os.Getenv("NWWSOI_SERVER") + ":5222",
		Room:     os.Getenv("NWWSOI_ROOM"),
		User:     os.Getenv("NWWSOI_USER"),
		Pass:     os.Getenv("NWWSOI_PASS"),
		Resource: os.Getenv("NWWSOI_RESOURCE"),
	}

	err := nwwsConfig.check()
	if err != nil {
		return nil, err
	}

	xmpp.DefaultConfig = &tls.Config{
		ServerName:         nwwsConfig.serverName(),
		InsecureSkipVerify: false,
	}

	options := xmpp.Options{
		Host:        nwwsConfig.Server,
		User:        nwwsConfig.User + "@" + nwwsConfig.serverName(),
		Password:    nwwsConfig.Pass,
		Resource:    nwwsConfig.Resource,
		NoTLS:       true,
		StartTLS:    true,
		Debug:       false, // Set to true if you want to see debug information
		Session:     true,
		DialTimeout: 60 * time.Second,
	}

	client, err := options.NewClient()
	if err != nil {
		return nil, err
	}

	slog.Info("\033[32m *** NWWS-OI Connected *** \033[m")

	_, err = client.SendOrg(fmt.Sprintf(`<presence xml:lang='en' from='%s@%s' to='%s@%s/%s'><x></x></presence>`, nwwsConfig.User, nwwsConfig.Server, nwwsConfig.Resource, nwwsConfig.Room, nwwsConfig.User))
	if err != nil {
		return nil, err
	}

	rabbit, err := amqp.Dial(os.Getenv("RABBIT"))
	if err != nil {
		return nil, err
	}

	ch, err := rabbit.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(
		exchangeOut, // name
		"direct",    // type
		true,        // durable
		false,       // auto-deleted
		false,       // internal
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(
		queueOut, // name
		true,     // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return nil, err
	}

	slog.Info("\033[32m *** RabbitMQ connected *** \033[m")

	// TODO: Monitoring server

	ctx, cancel := context.WithCancel(context.Background())
	server := &Server{
		xmpp:     client,
		messages: make(chan Message),
		rabbit:   rabbit,
		out:      ch,
		ctx:      ctx,
		cancel:   cancel,
	}

	return server, nil
}

/*
Run the server
*/
func (server *Server) Run() error {

	go func(server *Server) {
		slog.Info("\033[32m *** Starting NWWS-OI *** \033[m")
		for {
			select {
			// Stop and cleanup on cancel
			case <-server.ctx.Done():
				close(server.messages)
				slog.Info("NWWS-OI Stopped")
				return
			// Receive messages
			default:
				// Get message
				chat, err := server.xmpp.Recv()
				if err != nil {
					slog.Error("xmpp receive error", "error", err.Error())
					continue
				}

				// Parse and send message
				switch v := chat.(type) {
				case xmpp.Chat:
					for _, elem := range v.OtherElem {
						if elem.XMLName.Local == "x" {
							text := strings.ReplaceAll(elem.String(), "\n\n", "\n")
							server.messages <- Message{
								Text:       text,
								ReceivedAt: time.Now(),
							}
							// TODO: Implement monitor
							// nwws.last = time.Now().UTC()
						}
					}
				}
			}
		}
	}(server)

	go func(server *Server) {
		for message := range server.messages {
			body, err := json.Marshal(message)
			if err != nil {
				slog.Error(err.Error())
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = server.out.PublishWithContext(ctx,
				exchangeOut,   // exchange
				routingKeyOut, // routing key
				false,         // mandatory
				false,         // immediate
				amqp.Publishing{
					ContentType: "application/json",
					Body:        body,
				})
			if err != nil {
				slog.Error(err.Error())
				return
			}
		}
	}(server)

	return nil
}

func (server *Server) Shutdown() {
	server.cancel()
	err := server.xmpp.Close()
	if err != nil {
		slog.Error("failed to close xmpp connection", "error", err)
	}

	err = server.out.Close()
	if err != nil {
		slog.Error("failed to close out channel", "error", err)
	}

	err = server.rabbit.Close()
	if err != nil {
		slog.Error("failed to close rabbit connection", "error", err)
	}

}

func (conf *XmppConfig) check() error {
	item := ""
	switch "" {
	case conf.Server:
		item = "server"
	case conf.User:
		item = "user"
	case conf.Pass:
		item = "pass"
	case conf.Resource:
		item = "resource"
	case conf.Room:
		item = "room"
	}
	if item != "" {
		return fmt.Errorf("xmpp %s missing in config", item)
	}
	return nil
}

func (conf *XmppConfig) serverName() string {
	return strings.Split(conf.Server, ":")[0]
}
