//go:generate env GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -ldflags "-s -w" -o db.wasm main.go
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/extism/go-pdk"
	"github.com/jackc/pgx/v5"
	"github.com/mymmrac/wape/plugin/net"

	"github.com/mymmrac/lithium/pkg/plugin/protocol"
)

//go:wasmexport handler
func Handle() {
	var request protocol.Request
	if err := pdk.InputJSON(&request); err != nil {
		pdk.SetError(fmt.Errorf("unmarshal request: %w", err))
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := pgx.ParseConfig(os.Getenv("POSTGRES_CONNECTION_STRING"))
	if err != nil {
		respond(protocol.Response{
			StatusCode: 500,
			Body:       "Failed to parse DB config!\n\n" + err.Error(),
		})
		return
	}
	cfg.DialFunc = net.DefaultDialer.DialContext
	cfg.LookupFunc = net.DefaultResolver.LookupHost

	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		respond(protocol.Response{
			StatusCode: 500,
			Body:       "Failed to connect to DB!\n\n" + err.Error(),
		})
		return
	}
	defer func() { _ = conn.Close(context.Background()) }()

	if err = conn.Ping(ctx); err != nil {
		respond(protocol.Response{
			StatusCode: 500,
			Body:       "Failed to ping DB!\n\n" + err.Error(),
		})
		return
	}

	rows, err := conn.Query(ctx, `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';`)
	if err != nil {
		respond(protocol.Response{
			StatusCode: 500,
			Body:       "Failed to query DB!\n\n" + err.Error(),
		})
		return
	}

	var tableName string
	var tableNames []string
	for rows.Next() {
		if err = rows.Scan(&tableName); err != nil {
			respond(protocol.Response{
				StatusCode: 500,
				Body:       "Failed to scan result!\n\n" + err.Error(),
			})
			return
		}
		tableNames = append(tableNames, tableName)
	}

	respond(protocol.Response{
		StatusCode: 200,
		Body:       "DB Connected!\n\nTables:\n\t" + strings.Join(tableNames, "\n\t"),
	})
}

func respond(response protocol.Response) {
	if err := pdk.OutputJSON(response); err != nil {
		pdk.SetError(fmt.Errorf("marshal response: %w", err))
		return
	}
}

func main() {}
