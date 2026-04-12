# Goのシングルバイナリをビルドするためのコンテナ
FROM golang:1.25.4-trixie as deploy-builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -ldflags "-w -s" -o app

# ビルドしたバイナリをデプロイするためのコンテナ
FROM debian:trixie-slim as deploy

RUN apt-get update

COPY --from=deploy-builder /app/app .

CMD ["./app"]

# ローカル環境で利用するホットリロード環境
FROM golang:1.25.4 as dev
WORKDIR /app

# devコンテナのビルド時に依存関係を自動でダウンロード・インストールするためにgo.modとgo.sumを先にコピーしている
COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/air-verse/air@latest
CMD ["air"]