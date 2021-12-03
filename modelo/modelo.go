package modelo

import "gorm.io/gorm"

type Grupo struct {
	gorm.Model
	Identificador int64 `gorm:"unique"`
	Nombre        string
	Participantes []*Participante
	YaSorteo      bool
}

type Participante struct {
	gorm.Model
	GrupoID       uint
	Identificador int
	Nombre        string
	AmigoID       uint
	Amigo         *Participante
}

func NewGrupo(identificador int64, nombre string) *Grupo {
	return &Grupo{Identificador: identificador, Nombre: nombre}
}

func NewParticipante(identificador int, nombre string) *Participante {
	return &Participante{Identificador: identificador, Nombre: nombre}
}

func (g *Grupo) Agregar(participante *Participante) {
	enElGrupo := false

	for _, participanteEnElGrupo := range g.Participantes {
		if participanteEnElGrupo.Identificador == participante.Identificador {
			enElGrupo = true
			break
		}
	}

	if !enElGrupo {
		g.Participantes = append(g.Participantes, participante)
	}
}
