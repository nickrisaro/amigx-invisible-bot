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
		return nil, errors.New("yaSorteado")
	}

	cantidadDeParticipantes := len(grupoDeLaDB.Participantes)

	if cantidadDeParticipantes < 2 {
		return nil, errors.New("faltanParticipantes")
	}

	idsSorteados := make(map[int]bool, cantidadDeParticipantes)
	sorteados := make([]int, cantidadDeParticipantes)

	for i := 0; i < cantidadDeParticipantes; {
		idRandom := rand.Intn(cantidadDeParticipantes)
		_, idYaSorteado := idsSorteados[idRandom]

		if idRandom != i && !idYaSorteado {
			sorteados[i] = idRandom
			idsSorteados[idRandom] = true
			i++
		}
	}

	for i, participante := range grupoDeLaDB.Participantes {
		idAmigx := sorteados[i]
		amigx := grupoDeLaDB.Participantes[idAmigx]
		participante.Amigx = amigx

		resultado = lm.miBaseDeDatos.Save(participante)
		if resultado.Error != nil {
			return nil, resultado.Error
		}
	}

	grupoDeLaDB.YaSorteo = true
	resultado = lm.miBaseDeDatos.Omit("Participantes").Save(grupoDeLaDB)
	if resultado.Error != nil {
		return nil, resultado.Error
	}

	return grupoDeLaDB.Participantes, nil
}

func (lm *LaMaga) ParticipantesConAmigxs(identificadorDeGrupo int64) ([]*modelo.Participante, error) {
	grupoDeLaDB := modelo.Grupo{Identificador: identificadorDeGrupo}
	resultado := lm.miBaseDeDatos.Where(&grupoDeLaDB).Preload("Participantes").First(&grupoDeLaDB)
	if resultado.Error != nil {
		return nil, resultado.Error
	}

	if !grupoDeLaDB.YaSorteo {
		return nil, errors.New("noSorteado")
	}

	for _, participante := range grupoDeLaDB.Participantes {
		amigxDeLaDB := modelo.Participante{}
		resultado := lm.miBaseDeDatos.First(&amigxDeLaDB, *participante.AmigxID)
		if resultado.Error != nil {
			return nil, resultado.Error
		}
		participante.Amigx = &amigxDeLaDB
	}

	return grupoDeLaDB.Participantes, nil
}

func (lm *LaMaga) Borrar(identificadorDeGrupo int64) error {
	grupoDeLaDB := modelo.Grupo{Identificador: identificadorDeGrupo}
	resultado := lm.miBaseDeDatos.Where(&grupoDeLaDB).First(&grupoDeLaDB)
	if resultado.Error != nil {
		return resultado.Error
	}

	resultado = lm.miBaseDeDatos.Select("Participantes").Delete(grupoDeLaDB)

	return resultado.Error
}

func (lm *LaMaga) GruposDe(identificadorDeParticipante int) ([]*modelo.Grupo, error) {
	grupos := make([]*modelo.Grupo, 0)
	resultado := lm.miBaseDeDatos.Table("Grupos").
		Select("Grupos.*").
		Joins("left join Participantes on Participantes.grupo_id = Grupos.id").
		Where("Participantes.identificador = ?", identificadorDeParticipante).
		Scan(&grupos)
	return grupos, resultado.Error
}

func (lm *LaMaga) AmigxsDe(identificadorDeParticipante int) ([]GrupoAmigx, error) {
	grupos := make([]GrupoAmigx, 0)
	resultado := lm.miBaseDeDatos.Table("Grupos").
		Select("Grupos.Nombre Grupo, Amigx.Nombre Amigx").
		Joins("left join Participantes participante on participante.grupo_id = Grupos.id").
		Joins("left join Participantes Amigx on participante.amigx_id = Amigx.id").
		Where("participante.identificador = ?", identificadorDeParticipante).
		Where("Grupos.Ya_Sorteo = true").
		Scan(&grupos)
	return grupos, resultado.Error
}

type GrupoAmigx struct {
	Grupo string
	Amigx string
}
