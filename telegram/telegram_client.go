package telegram

import (
	"fmt"

	"github.com/nickrisaro/invisible-bot/lamaga"
	"gopkg.in/tucnak/telebot.v2"
	tb "gopkg.in/tucnak/telebot.v2"
)

func Configurar(urlPublica string, urlPrivada string, token string, maga *lamaga.LaMaga) (*tb.Bot, error) {
	b, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.Webhook{Listen: urlPrivada, Endpoint: &tb.WebhookEndpoint{PublicURL: urlPublica, Cert: ""}},
	})

	if err != nil {
		fmt.Println("Error al configurar", err)
		return nil, err
	}

	b.Handle("/ping", func(m *tb.Message) {
		b.Send(m.Chat, "Pong!")
	})

	b.Handle("/start", func(m *tb.Message) {
		b.Send(m.Chat, "Hola soy La Maga, si querés jugar al amigo, amiga, amigue, amigx invisble yo te puedo ayudar")
		b.Send(m.Chat, "Si ya estás jugando en un grupo te voy a avisar por acá a quién le tenés que regalar algo")
		b.Send(m.Chat, "Si todavía no estás jugando, agregame en alguno de tus grupos y empezá el juego!")
	})

	b.Handle("/help", func(m *tb.Message) {
		ayuda := "Hola soy La Maga, si querés jugar al amigo, amiga, amigue, amigx invisble yo te puedo ayudar\n"
		ayuda += "Para empezar mandá el comando /comenzar así preparo todo\n"
		ayuda += "Cada persona que quiera participar tiene que mandar /sumame\n"
		ayuda += "Cuando todas las personas se hayan sumado mandá /sortear\n"
		b.Send(m.Chat, ayuda)
	})

	b.Handle("/comenzar", func(m *tb.Message) {
		nombreDelGrupo := m.Chat.Title
		if len(nombreDelGrupo) == 0 {
			nombreDelGrupo = m.Chat.FirstName + " " + m.Chat.LastName
		}
		err := maga.NuevoGrupo(m.Chat.ID, nombreDelGrupo)
		if err != nil {
			fmt.Println("Error al crear grupo", err)
			b.Send(m.Chat, "Ups, no pude crear tu grupo, probá más tarde")
		} else {
			b.Send(m.Chat, "Listo, ya creé tu grupo, ahora cada persona que quiera jugar tiene que mandar /sumame")
		}
	})

	b.Handle("/sumame", func(m *tb.Message) {
		nombreDelGrupo := m.Chat.Title
		if len(nombreDelGrupo) == 0 {
			nombreDelGrupo = m.Chat.FirstName + " " + m.Chat.LastName
		}
		err := maga.NuevoParticipante(m.Chat.ID, m.Sender.ID, m.Sender.FirstName+" "+m.Sender.LastName)
		if err != nil {
			fmt.Println("Error al agregar persona al grupo", err)
			b.Send(m.Chat, "Ups, no pude agregar a la persona al grupo ¿Ya creaste el grupo con /comenzar ?")
		} else {
			_, err = b.Send(m.Sender, "Hola, te anoté para jugar al amigx invisible en el grupo "+nombreDelGrupo+". Cuando hagan el sorteo te voy a avisar a quién le tenés que regalar algo.")
			if err != nil {
				b.Send(m.Chat, "@"+m.Sender.Username+" no te puedo mandar mensajes, me tenés que hablar vos primero, andá a @amigxinvisiblebot y tocá Start")
			}
			b.Send(m.Chat, "Listo, ya agregué a @"+m.Sender.Username+" al grupo.\nSi ya se sumaron todas las personas mandá /sortear\nSi querés ver quienes se sumaron mandá /listar")
		}
	})

	b.Handle("/listar", func(m *tb.Message) {
		participantes, err := maga.QuienesParticipan(m.Chat.ID)
		if err != nil {
			fmt.Println("Error al listar participantes", err)
			b.Send(m.Chat, "Ups, no pude encontrar a las personas que participan ¿Ya creaste el grupo con /comenzar ?")
		} else {
			if len(participantes) == 0 {
				b.Send(m.Chat, "Todavía no se anotó nadie, se pueden sumar al juego con /sumame")
			} else {
				listaDeParticipantes := "Ya se anotaron para jugar:\n"
				for _, participante := range participantes {
					listaDeParticipantes += " * " + participante + "\n"
				}
				b.Send(m.Chat, listaDeParticipantes)
			}
		}
	})

	b.Handle("/sortear", func(m *tb.Message) {
		sorteados, err := maga.Sortear(m.Chat.ID)
		nombreDelGrupo := m.Chat.Title
		if len(nombreDelGrupo) == 0 {
			nombreDelGrupo = m.Chat.FirstName + " " + m.Chat.LastName
		}

		if err != nil {
			fmt.Println("Error al sortear", err)
			if err.Error() == "faltanParticipantes" {
				b.Send(m.Chat, "Necesito al menos dos personas para poder sortear")
			} else {
				b.Send(m.Chat, "Ups, no pude sortear ¿Ya creaste el grupo con /comenzar ?")
			}
		} else {
			notifiquéA := 0
			for _, participante := range sorteados {
				mensaje := "Hola, " + participante.Nombre +
					" soy La Maga y te escribo porque estás jugando al amigx invisible en el grupo " + nombreDelGrupo +
					". La persona a la que le tenés que hacer un regalo es: " + participante.Amigx.Nombre + "!! Pensá en algo lindo para regalarle!"
				_, err = b.Send(&telebot.User{ID: participante.Identificador}, mensaje)
				if err == nil {
					notifiquéA++
				} else {
					fmt.Println("Error al notificar", err)
					b.Send(m.Chat, participante.Nombre+" no te pude mandar un mensaje, andá a @amigxinvisiblebot y tocá Start")
				}
			}
			if notifiquéA < len(sorteados) {
				b.Send(m.Chat, "Ups, no le pude mandar el mensaje a algunas personas, probá de sortear de nuevo en un rato")
			} else {
				b.Send(m.Chat, "Listo, ya hice el sorteo, cada participante recibió un mensaje privado con el nombre de la persona a la que le tiene que regalar algo")
			}
		}
	})

	return b, nil
}
