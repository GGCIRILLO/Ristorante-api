# 🍽️ Ristorante API

  
Sistema per la gestione di un ristorante, sviluppato in Go, con persistenza su PostgreSQL e caching Redis. Progetto realizzato come parte del corso di **Ingegneria del Software**.

  
## 📂 Struttura del Progetto
``` bash 
Ristorante-api/
├── api/
	├── handlers/  # HTTP handlers per ogni risorsa
	routes.go # Definizione delle API e routing  
├── cache/  # Gestione Redis cache per risorse 
├── config/  # Configurazione applicativa (es. variabili env)
├── database/ 
	├── migrations/ # Schema SQL (per consistenza e script seed)
	postgres.go # Definizione schema in go e creazione connessione al db e cache
	redis.go # Creazione RedisClient e funzioni monitoring
	seed.go # Funzione di seeding da file .sql
├── docker/Dockerfile  # Containerizzazione dell’applicazione Go
├── models/  # Modelli Go delle entità (struct)
├── repository/  # Query SQL e logica di accesso a database
├── .env # Variabili di ambiente
├── docker-compose.yml  # Setup completo con PostgreSQL e Redis
├── main.go  # Entrypoint dell’applicazione
├── go.mod / go.sum  # Gestione delle dipendenze Go
├── manage.sh # Script di utilità
└── README.md  # Documentazione 
```

## 🚀 Esecuzione del progetto
### Requisiti
- Docker
- Docker Compose
### Opzione 1: `manage.sh`
Lo script  manage.sh  semplifica le operazioni comuni. Ecco i comandi disponibili:
```bash
./manage.sh up   	# Avvia i container (equivale a docker-compose up -d)
./manage.sh start 	# Avvia tutti i container in background (docker-compose up -d)
./manage.sh stop 	# Ferma e rimuove i container (docker-compose down)
./manage.sh rebuild # Ricostruisce i container senza rimuovere volumi
./manage.sh reset 	# Ferma, rimuove i volumi, ricostruisce tutto da zero 
./manage.sh logs 	# Mostra i log del container  ristorante-api-db
./manage.sh shell 	# Accede alla shell del container PostgreSQL
./manage.sh seed-check # Esegue una query per verificare il contenuto della tabella  ristorante(complete_seed.sql)
```
Allora, per eseguire l’intero progetto (API + PostgreSQL + Redis):
```bash
chmod +x manage.sh # per rendere eseguibile lo script 
./manage.sh start 
```
>  ⚠️ Assicurati che il file  .env  sia presente nella root del progetto.

L’API sarà accessibile su:  [http://localhost:8080](http://localhost:8080/), mentre adminer per la gestione web del database su: [http://localhost:8082](http://localhost:8082). 

### Opzione 2: avvio manuale
Dal terminale eseguire il comando :
```bash 
docker-compose up --build
```
L’API sarà accessibile su:  [http://localhost:8080](http://localhost:8080/), mentre adminer per la gestione web del database su: [http://localhost:8082](http://localhost:8082). 

## Esempi di API
Tali richieste sono state fatte da terminale con `curl`, ma si può usare anche Postman o simili. 

### **📍 Recuperare tutti i tavoli**
```bash
curl http://localhost:8080/api/tavoli
```
### **➕ Creare un nuovo ordine**
```bash
curl -X POST http://localhost:8080/api/ordini \
-H "Content-Type: application/json" \
-d  '{
		"id_tavolo": 2,
		"num_persone": 2,
		"id_ristorante": 1
	}'
```
### **🔄 Aggiornare stato ordine**
```bash
curl -X PATCH http://localhost:8080/api/ordini/3/stato \
  -H "Content-Type: application/json" \
  -d '{"stato": "in preparazione"}'
  ```

## **⚡ Caching con Redis**
L’applicazione utilizza  **Redis**  per memorizzare in cache le informazioni più usate e più utili, recuperandole in tempo minimo.
In caso di modifica, la cache viene invalidata automaticamente per garantire consistenza. Un esempio direttamente dai log di Docker: 
```bash
# Prima richiesta (cache miss)
2025/06/05 20:46:54 stdout: 2025/06/05 20:46:54 "GET http://localhost:8080/api/ordini/ HTTP/1.1" from .. - 200 140B in 3.596125ms
# Seconda richiesta (cache hit)
2025/06/05 20:47:16 stdout: 2025/06/05 20:47:16 "GET http://localhost:8080/api/ordini/ HTTP/1.1" from .. - 200 140B in 533.833µs
```
➡️ **Riduzione da 3.6 ms a 0.5 ms** grazie a Redis.
Si può verificare lo stato di Redis con: `GET /monitoring/redis`.

## **🛠 Stack Tecnologico**

**Componente** | **Tecnologia**
|--------------------|-------- 
Linguaggio | Go (Golang)|
Framework | Chi Router
Database | PostgreSQL
Caching | Redis
Architettura | MVC-like + Repo
Container | Docker + Compose

