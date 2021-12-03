package modelo_test

import (
	"testing"

	"github.com/nickrisaro/invisible-bot/modelo"
	"github.com/stretchr/testify/assert"
)

func TestSePuedeCrearUnGrupo(t *testing.T) {
	g := modelo.NewGrupo(1234, "Mi grupo")

	assert.NotNil(t, g, "El grupo no debería ser nil")
	assert.Equal(t, int64(1234), g.Identificador, "No tiene el identificador correcto")
	assert.Equal(t, "Mi grupo", g.Nombre, "No tiene el nombre correcto")
	assert.Empty(t, g.Participantes, "No debería tener participantes")
	assert.False(t, g.YaSorteo, "No debería estar sorteado")
}

func TestSePuedeCrearUnParticipante(t *testing.T) {
	p := modelo.NewParticipante(123, "Nick Risaro")

	assert.NotNil(t, p, "El particpante no debería ser nil")
	assert.Equal(t, 123, p.Identificador, "No tiene el identificador correcto")
	assert.Equal(t, "Nick Risaro", p.Nombre, "No tiene el nombre correcto")
	assert.Nil(t, p.Amigx, "No debería tener un amigx aún")
}

func TestSePuedeAgregarUnParticipanteAUnGrupo(t *testing.T) {
	g := modelo.NewGrupo(1234, "Mi grupo")
	p := modelo.NewParticipante(123, "Nick Risaro")

	g.Agregar(p)

	assert.NotEmpty(t, g.Participantes, "Debería tener participantes")
	assert.Len(t, g.Participantes, 1, "Debería haber un único participante")
	assert.Equal(t, p, g.Participantes[0], "Nick debería estar en el grupo")
}

func TestNoSePuedeAgregarDosVecesUnParticipanteAUnGrupo(t *testing.T) {
	g := modelo.NewGrupo(1234, "Mi grupo")
	p := modelo.NewParticipante(123, "Nick Risaro")

	g.Agregar(p)
	g.Agregar(p)

	assert.NotEmpty(t, g.Participantes, "Debería tener participantes")
	assert.Len(t, g.Participantes, 1, "Debería haber un único participante")
	assert.Equal(t, p, g.Participantes[0], "Nick debería estar en el grupo")
}
