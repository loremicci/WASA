<div align="center">

# 📱💬 WASAText

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Vue.js](https://img.shields.io/badge/Vue.js-35495E?style=for-the-badge&logo=vue.js&logoColor=4FC08D)
![Vite](https://img.shields.io/badge/Vite-B73BFE?style=for-the-badge&logo=vite&logoColor=FFD62E)
![SQLite](https://img.shields.io/badge/SQLite-07405E?style=for-the-badge&logo=sqlite&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2CA5E0?style=for-the-badge&logo=docker&logoColor=white)
![API](https://img.shields.io/badge/API-REST-000000?style=for-the-badge&logo=json&logoColor=white)
![YAML](https://img.shields.io/badge/Config-YAML-cb171e?style=for-the-badge&logo=yaml&logoColor=white)

*Un'applicazione web di messaggistica full-stack in tempo reale con chat private, gruppi, condivisione di media e reazioni con emoji.*

</div>

<br />

## 🚀 Panoramica del Progetto

**WASAText** è un'applicazione web di messaggistica sviluppata per il corso di Web and Software Architecture (WASA). L'obiettivo del progetto è costruire un'architettura Client-Server robusta, mettendo in pratica concetti avanzati di sviluppo backend, interfacce reattive e containerizzazione.

L'applicazione permette agli utenti di chattare in tempo reale, scambiarsi foto, creare gruppi e gestire reazioni ai messaggi, tutto all'interno di un'interfaccia moderna e "glassmorphic".

## ✨ Funzionalità Principali

- 👤 **Gestione Profilo**: Login semplificato, aggiornamento di username e foto profilo in tempo reale.
- 💬 **Chat Private & Gruppi**: Avvia conversazioni 1-a-1 o crea gruppi invitando altri utenti.
- 📸 **Condivisione Media**: Invia messaggi contenenti testo e immagini contemporaneamente.
- ✔️ **Spunte di Lettura**: Monitoraggio avanzato dei messaggi ricevuti (singola spunta) e letti (doppia spunta blu).
- 😂 **Reazioni Emoji**: Lascia una reazione su qualsiasi messaggio (massimo una per utente).
- ➡️ **Inoltro Messaggi**: Inoltra rapidamente messaggi tra diverse chat o cerca nuovi utenti a cui inviarli.
- 🌓 **Dark Mode & UI Premium**: Tema scuro integrato, interfacce "glassmorphism" e sfondi animati, con preferenze salvate in locale.
- 🔒 **Sicurezza Base**: Controllo autorizzazioni tramite token Bearer e check rigorosi dei permessi di lettura per i gruppi dal backend.

## 🛠️ Architettura e Tecnologie

L'applicazione è divisa in due componenti principali:

### Backend (API REST)

- **Linguaggio**: Golang
- **Database**: SQLite (gestito nativamente dal backend senza ORM pesanti)
- **Funzionamento**: Espone un'API RESTful sulla porta 3000, gestisce la logica di business, l'autenticazione tramite Bearer Token e la validazione dei permessi e dei dati.

### Frontend (SPA)

- **Framework**: Vue 3 (Composition API)
- **Build Tool**: Vite
- **Funzionamento**: Single Page Application reattiva. Utilizza *Axios* con un interceptor globale per le chiamate di rete asincrone verso il backend.

## 🐳 Docker Deployment

L'intero progetto è facilmente avviabile tramite **Docker** con immagini ottimizzate *multi-stage*.

Per fare la build delle immagini e far partire i container, utilizza i seguenti comandi:

```bash
# Build e avvio del Backend
docker build -t wasa-backend -f Dockerfile.backend .
docker run --rm --user 1000 -p 3000:3000 wasa-backend

# Build e avvio del Frontend
docker build -t wasa-frontend -f Dockerfile.frontend .
docker run --rm -p 8080:80 wasa-frontend

# Pagina Web
http://localhost:8080/
```

## 📚 API Endpoints (Panoramica)

Il backend espone decine di endpoint protetti per gestire le operazioni. Alcuni esempi principali:

- `POST /session` - Autenticazione utente (Login)
- `DELETE /session` - Chiusura della sessione (Logout)
- `GET /conversations` - Recupero delle chat dell'utente (Private e Gruppi)
- `POST /conversations/:id/messages` - Invio di un messaggio testuale e/o multimediale
- `PUT /messages/:id/comments/:emoticon` - Aggiunta di una reazione emoji

*(Tutti gli endpoint, ad eccezione di `/session`, richiedono l'header HTTP `Authorization: Bearer <user_id>`)*

---
<div align="center">
Sviluppato come progetto universitario finale per il corso di Web and Software Architecture.
</div>
