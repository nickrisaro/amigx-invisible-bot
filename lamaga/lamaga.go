package lamaga

import (
	"errors"
	"math/rand"
	"strings"

	"github.com/nickrisaro/invisible-bot/modelo"
	"gorm.io/gorm"
)

type LaMaga struct {
	miBaseDeDatos *gorm.DB
}

func NewMaga(baseDeDatos *gorm.DB) *LaMaga {
	return &LaMaga{miBaseDeDatos: baseDeDatos}
}

func (lm *LaMaga) NuevoGrupo(identificador int64, nombre string) error {
	grupo := modelo.NewGrupo(identificador, nombre)
	resultado := lm.miBaseDeDatos.Create(grupo)
	if resultado.Error != nil && strings.Contains(resultado.Error.Error(), "UNIQUE") {
		return errors.New("ya existe ese grupo")
	}
	return resultado.Error
}

func (lm *LaMaga) NuevoParticipante(identificadorDeGrupo int64, identificadorDeParticipante int, nombreDeParticipante string) error {
	participante := modelo.NewParticipante(identificadorDeParticipante, nombreDeParticipante)

	grupoDeLaDB := modelo.Grupo{Identificador: identificadorDeGrupo}
	resultado := lm.miBaseDeDatos.Where(&grupoDeLaDB).Preload("Participantes").First(&grupoDeLaDB)
	if resultado.Error != nil {
		return resultado.Error
	}

	grupoDeLaDB.Agregar(participante)

	resultado = lm.miBaseDeDatos.Session(&gorm.Session{FullSaveAssociations: true}).Save(grupoDeLaDB)
	return resultado.Error
}

func (lm *LaMaga) QuienesParticipan(identificadorDeGrupo int64) ([]string, error) {
	grupoDeLaDB := modelo.Grupo{Identificador: identificadorDeGrupo}
	resultado := lm.miBaseDeDatos.Where(&grupoDeLaDB).Preload("Participantes").First(&grupoDeLaDB)
	if resultado.Error != nil {
		return nil, resultado.Error
	}

	nombresDeParticipantes := make([]string, len(grupoDeLaDB.Participantes))

	for i, participante := range grupoDeLaDB.Participantes {
		nombresDeParticipantes[i] = participante.Nombre
	}

	return nombresDeParticipantes, nil
}

func (lm *LaMaga) Sortear(identificadorDeGrupo int64) ([]*modelo.Participante, error) {
	grupoDeLaDB := modelo.Grupo{Identificador: identificadorDeGrupo}
	resultado := lm.miBaseDeDatos.Where(&grupoDeLaDB).Preload("Participantes").First(&grupoDeLaDB)
	if resultado.Error != nil {
		return nil, resultado.Error
	}

	if grupoDeLaDB.YaSorteo {
		return grupoDeLaDB.Participantes, nil
	}

	resultado = lm.miBaseDeDatos.Save(grupoDeLaDB.Participantes)
	if resultado.Error != nil {
		return nil, resultado.Error
	}

	cantidadDeParticipantes := len(grupoDeLaDB.Participantes)

	idsSorteados := make(map[int]bool, cantidadDeParticipantes)
	sorteados := make([]int, cantidadDeParticipantes)

	for i := 0; i < cantidadDeParticipantes; {
		idRandom := rand.Intn(cantidadDeParticipantes)
		if idRandom != i && !idsSorteados[idRandom] {
			sorteados[i] = idRandom
			idsSorteados[idRandom] = true
			i++
		}
	}

	for i, participante := range grupoDeLaDB.Participantes {
		idAmigx := sorteados[i]
		amigx := grupoDeLaDB.Participantes[idAmigx]
		participante.Amigo = amigx
	}

	grupoDeLaDB.YaSorteo = true
	resultado = lm.miBaseDeDatos.Save(grupoDeLaDB)
	if resultado.Error != nil {
		return nil, resultado.Error
	}

	return grupoDeLaDB.Participantes, nil
}
