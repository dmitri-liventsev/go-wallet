FROM golang:1.25

RUN apt-get update && apt-get install -y git

RUN go install github.com/air-verse/air@latest
RUN go install github.com/onsi/ginkgo/v2/ginkgo@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV GOFLAGS="-buildvcs=false"

EXPOSE 8080

CMD ["air"]