package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nickrisaro/invisible-bot/lamaga"
	"github.com/nickrisaro/invisible-bot/modelo"
	"github.com/nickrisaro/invisible-bot/telegram"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Print("Iniciando invisible-bot")

	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	token := os.Getenv("TELEGRAM_API_TOKEN")
	urlPublica := os.Getenv("APP_URL")
	urlDB := os.Getenv("DATABASE_URL")

	if host == "" {
		log.Print("Usando host default")
		host = "0.0.0.0"
	}

	if port == "" {
		log.Print("Usando port default")
		port = "3000"
	}

	if token == "" {
		log.Fatal("Se debe setear la variable TELEGRAM_API_TOKEN")
		return
	}

	if urlPublica == "" {
		log.Fatal("Se debe setear la variable APP_URL")
		return
	}

	if urlDB == "" {
		log.Fatal("Se debe setear la variable DATABASE_URL")
		return
	}

	db, err := gorm.Open(postgres.Open(urlDB), &gorm.Config{})
	if err != nil {
		log.Fatal("No pude conectarme a la base de datos", err)
		return
	}
	db.AutoMigrate(&modelo.Grupo{}, &modelo.Participante{})
	if err != nil {
		log.Fatal("No pude migrar las tablas", err)
		return
	}

	b, err := telegram.Configurar(urlPublica, fmt.Sprintf("%s:%s", host, port), token, lamaga.NewMaga(db))
	if err != nil {
		log.Fatal("No pude iniciar el bot", err)
		return
	}

	log.Printf("Webhook escuchando en %s -> %s:%s", urlPublica, host, port)
	b.Start()
}
