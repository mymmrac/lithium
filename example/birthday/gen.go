//go:generate env GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -ldflags "-s -w" -o ./home.wasm ./home/main.go
//go:generate env GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -ldflags "-s -w" -o ./chat.wasm ./chat/main.go
//go:generate env GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -ldflags "-s -w" -o ./session.wasm ./session/main.go
//go:generate env GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -ldflags "-s -w" -o ./me.wasm ./me/main.go
//go:generate env GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -ldflags "-s -w" -o ./logout.wasm ./logout/main.go

package birthday
