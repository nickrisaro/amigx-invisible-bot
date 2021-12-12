package telegram

import (
	"fmt"

	"github.com/nickrisaro/invisible-bot/lamaga"
	"github.com/nickrisaro/invisible-bot/modelo"

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
		b.Send(m.Chat, "Si querés ver en que grupos estás jugando mandá /misGrupos y si querés ver a quién le tenés que regalar mandá /misAmigxs")
	})

	b.Handle("/help", func(m *tb.Message) {
		ayuda := "Hola soy La Maga, si querés jugar al amigo, amiga, amigue, amigx invisble yo te puedo ayudar\n"
		ayuda += "Para empezar mandá el comando /comenzar así preparo todo\n"
		ayuda += "Cada persona que quiera participar tiene que mandar /sumame\n"
		ayuda += "Cuando todas las personas se hayan sumado mandá /sortear\n"
		ayuda += "Si querés ver en que grupos estás jugando mandá /misGrupos (lo podés mandar en un grupo y la respuesta te llega sólo a vos)\n"
		ayuda += "Si querés ver a quién le tenés que regalar mandá /misAmigxs (lo podés mandar en un grupo y la respuesta te llega sólo a vos)\n"
		b.Send(m.Chat, ayuda)
	})

	b.Handle("/comenzar", func(m *tb.Message) {
		if !m.FromGroup() {
			b.Send(m.Chat, "No podés comenzar en un chat privado, agregame a un grupo con tus amigxs y mandá /comenzar ahí")
			return
		}
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
		nombreCompletoParticipante := m.Sender.FirstName + " " + m.Sender.LastName
		username := nombreCompletoParticipante
		if len(m.Sender.Username) > 0 {
			username = m.Sender.Username
		}
		err := maga.NuevoParticipante(m.Chat.ID, m.Sender.ID, nombreCompletoParticipante)
		if err != nil {
			fmt.Println("Error al agregar persona al grupo", err)
			b.Send(m.Chat, "Ups, no pude agregar a la persona al grupo ¿Ya creaste el grupo con /comenzar ?")
		} else {
			_, err = b.Send(m.Sender, "Hola, te anoté para jugar al amigx invisible en el grupo "+nombreDelGrupo+". Cuando hagan el sorteo te voy a avisar a quién le tenés que regalar algo.")
			if err != nil {
				b.Send(m.Chat, "@"+username+" no te puedo mandar mensajes, me tenés que hablar vos primero, andá a @amigxinvisiblebot y tocá Start")
			}
			b.Send(m.Chat, "Listo, ya agregué a @"+username+" al grupo.\nSi ya se sumaron todas las personas mandá /sortear\nSi querés ver quienes se sumaron mandá /listar")
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
			} else if err.Error() == "yaSorteado" {
				b.Send(m.Chat, "Ya hice el sorteo en este grupo, si querés que vuelva a notificar mandá /notificar")
			} else {
				b.Send(m.Chat, "Ups, no pude sortear ¿Ya creaste el grupo con /comenzar ?")
			}
		} else {
			mandarMensajes(b, m.Chat, sorteados, nombreDelGrupo)
		}
	})

	b.Handle("/notificar", func(m *tb.Message) {
		sorteados, err := maga.ParticipantesConAmigxs(m.Chat.ID)
		nombreDelGrupo := m.Chat.Title
		if len(nombreDelGrupo) == 0 {
			nombreDelGrupo = m.Chat.FirstName + " " + m.Chat.LastName
		}

		if err != nil {
			fmt.Println("Error al notificar", err)
			if err.Error() == "noSorteado" {
				b.Send(m.Chat, "No hice el sorteo en este grupo, si querés sortear mandá /sortear")
			} else {
				b.Send(m.Chat, "Ups, no pude mandar los mensajes ¿Ya creaste el grupo con /comenzar y sorteaste con /sortear ?")
			}
		} else {
			mandarMensajes(b, m.Chat, sorteados, nombreDelGrupo)
		}
	})

	b.Handle("/misGrupos", func(m *tb.Message) {
		gruposDeParticipante, err := maga.GruposDe(m.Sender.ID)
		if err != nil {
			fmt.Println("Error al listar grupos", err)
			b.Send(m.Sender, "Ups, no pude encontrar tus grupos ¿Ya creaste alguno grupo con /comenzar y te sumaste con /sumame ?")
		} else {
			if len(gruposDeParticipante) == 0 {
				b.Send(m.Sender, "Todavía no te anotaste en ningún grupo, te podés sumar mandando /sumame en algún grupo")
			} else {
				listaDeGrupos := "Estás jugando en:\n"
				for _, participante := range gruposDeParticipante {
					listaDeGrupos += " * " + participante.Nombre + "\n"
				}
				b.Send(m.Sender, listaDeGrupos)
			}
		}
	})

	b.Handle("/misAmigxs", func(m *tb.Message) {
		gruposyAmigxs, err := maga.AmigxsDe(m.Sender.ID)
		if err != nil {
			fmt.Println("Error al listar amigxs", err)
			b.Send(m.Sender, "Ups, no pude encontrar tus amigxs ¿Ya creaste algun grupo con /comenzar te sumaste con /sumame y sorteaste con /sortear ?")
		} else {
			if len(gruposyAmigxs) == 0 {
				b.Send(m.Sender, "Todavía no tenés amigxs en ningún grupo, te podés sumar mandando /sumame en algún grupo y después sortear con /sortear")
			} else {
				listaDeGruposYAmigxs := "Estos son tus amigxs:\n"
				for _, grupoAmigx := range gruposyAmigxs {
					listaDeGruposYAmigxs += "\\* En el grupo *" + grupoAmigx.Grupo + "* le tenés que regalar a *" + grupoAmigx.Amigx + "*\n"
				}
				b.Send(m.Sender, listaDeGruposYAmigxs, tb.ModeMarkdownV2)
			}
		}
	})

	b.Handle("/terminar", func(m *tb.Message) {
		err := maga.Borrar(m.Chat.ID)

		if err != nil {
			fmt.Println("Error al borrar", err)
			b.Send(m.Chat, "Ups, no pude borrar el grupo, probá más tarde")
		} else {
			b.Send(m.Chat, "Listo, ya borré todo, si querés volver a jugar mandá /comenzar")
		}
	})

	return b, nil
}

func mandarMensajes(b *tb.Bot, chat *tb.Chat, sorteados []*modelo.Participante, nombreDelGrupo string) {
	var err error
	notifiquéA := 0
	for _, participante := range sorteados {
		mensaje := "Hola, " + participante.Nombre +
			" soy La Maga y te escribo porque estás jugando al amigx invisible en el grupo " + nombreDelGrupo +
			". La persona a la que le tenés que hacer un regalo es: " + participante.Amigx.Nombre + "!! Pensá en algo lindo para regalarle!"
		_, err = b.Send(&tb.User{ID: participante.Identificador}, mensaje)
		if err == nil {
			notifiquéA++
		} else {
			fmt.Println("Error al notificar", err)
			b.Send(chat, participante.Nombre+" no te pude mandar un mensaje, andá a @amigxinvisiblebot y tocá Start")
		}
	}
	if notifiquéA < len(sorteados) {
		b.Send(chat, "Ups, no le pude mandar el mensaje a algunas personas, probá de nuevo en un rato")
	} else {
		b.Send(chat, "Listo, cada participante recibió un mensaje privado con el nombre de la persona a la que le tiene que regalar algo")
	}
}
