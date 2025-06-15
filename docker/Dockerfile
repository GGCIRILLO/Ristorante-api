FROM golang:1.21-alpine

WORKDIR /app

# Installa git e aggiorna Alpine
RUN apk update && apk upgrade && apk add --no-cache git

# Copia go.mod e go.sum prima del sorgente per sfruttare la cache Docker
COPY go.mod ./
COPY go.sum ./

# Risolve le dipendenze (meglio separare per caching)
RUN go mod tidy && go mod download

# Copia il codice sorgente (solo ora)
COPY . .

# Copia i file di SQL 
COPY database/ ./database/


# Compila il binario (dopo che tutto il codice è stato copiato)
RUN go build -o main .

# Esegui con CompileDaemon per hot-reload (solo in dev)
RUN go install github.com/githubnemo/CompileDaemon@latest

# Espone la porta
EXPOSE 8080

CMD ["CompileDaemon", "--build=go build -buildvcs=false -o main .", "--command=./main"]