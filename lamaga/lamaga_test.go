package lamaga_test

import (
	"math/rand"
	"testing"

	"github.com/nickrisaro/invisible-bot/lamaga"
	"github.com/nickrisaro/invisible-bot/modelo"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const conexiónALaBase = "file::memory:?cache=shared"

type LaMagaTestSuite struct {
	suite.Suite
	db   *gorm.DB
	maga *lamaga.LaMaga
}

func (suite *LaMagaTestSuite) SetupTest() {
	db, err := gorm.Open(sqlite.Open(conexiónALaBase), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	suite.NoError(err, "Debería conectarse a la base de datos")
	suite.NotNil(db, "La base no debería ser nula")
	suite.db = db

	err = suite.db.AutoMigrate(&modelo.Grupo{}, &modelo.Participante{})
	suite.NoError(err, "Debería ejecutar las migraciones")

	suite.maga = lamaga.NewMaga(suite.db)

	suite.NotNil(suite.maga, "La maga no debería ser nil")
}

func (suite *LaMagaTestSuite) TestLaMagaPuedeCrearUnGrupo() {
	IDNuevoGrupo := int64(rand.Int())
	err := suite.maga.NuevoGrupo(IDNuevoGrupo, "Mi grupo")

	suite.NoError(err, "No debería fallar al crear el grupo")
	grupoDeLaDB := modelo.Grupo{Identificador: IDNuevoGrupo}
	resultado := suite.db.Preload("Participantes").Where(grupoDeLaDB).First(&grupoDeLaDB)
	suite.NoError(resultado.Error, "Debería haber encontrado el grupo")
	suite.Empty(grupoDeLaDB.Participantes, "No debería tener participantes")
	suite.Equal("Mi grupo", grupoDeLaDB.Nombre)
}

func (suite *LaMagaTestSuite) TestLaMagaNoCreaDosVecesElMismoGrupo() {
	IDNuevoGrupo := int64(rand.Int())
	suite.maga.NuevoGrupo(IDNuevoGrupo, "Mi grupo")

	err := suite.maga.NuevoGrupo(IDNuevoGrupo, "Mi grupo")

	suite.Error(err, "Debería fallar al crear el grupo 2 veces")
}

func (suite *LaMagaTestSuite) TestLaMagaPuedeAgregarUnParticipanteAUnGrupo() {
	IDNuevoGrupo := int64(rand.Int())
	suite.maga.NuevoGrupo(IDNuevoGrupo, "Mi grupo")

	IDNuevoParticipante := rand.Int()
	err := suite.maga.NuevoParticipante(IDNuevoGrupo, IDNuevoParticipante, "Nick")

	suite.NoError(err, "No debería fallar al crear el participante")
	grupoDeLaDB := modelo.Grupo{Identificador: IDNuevoGrupo}
	resultado := suite.db.Preload("Participantes").Where(grupoDeLaDB).First(&grupoDeLaDB)
	suite.NoError(resultado.Error, "Debería haber encontrado el grupo")
	suite.NotEmpty(grupoDeLaDB.Participantes, "No debería tener participantes")
	suite.Equal("Nick", grupoDeLaDB.Participantes[0].Nombre)
}

func (suite *LaMagaTestSuite) TestLaMagaNoAgregaUnParticipanteSiNoHayUnGrupo() {
	IDNuevoGrupo := int64(rand.Int())
	IDNuevoParticipante := rand.Int()

	err := suite.maga.NuevoParticipante(IDNuevoGrupo, IDNuevoParticipante, "Nick")

	suite.Error(err, "Debería fallar al agregar participantes a un grupo inexistente")
}

func (suite *LaMagaTestSuite) TestLaMagaNosDaLosParticipantesDeUnGrupo() {
	IDNuevoGrupo := int64(rand.Int())
	suite.maga.NuevoGrupo(IDNuevoGrupo, "Mi grupo")
	IDNuevoParticipante := rand.Int()
	suite.maga.NuevoParticipante(IDNuevoGrupo, IDNuevoParticipante, "Nick")

	participantes, err := suite.maga.QuienesParticipan(IDNuevoGrupo)

	suite.NoError(err, "No debería fallar al crear el participante")
	suite.NotEmpty(participantes, "No debería tener participantes")
	suite.Equal("Nick", participantes[0])
}

func (suite *LaMagaTestSuite) TestLaMagaNoListaParticipanteSiNoHayUnGrupo() {
	IDNuevoGrupo := int64(rand.Int())

	participantes, err := suite.maga.QuienesParticipan(IDNuevoGrupo)

	suite.Error(err, "Debería fallar al buscar participantes de un grupo inexistente")
	suite.Nil(participantes, "No debería haber participantes si no hay grupo")
}

func (suite *LaMagaTestSuite) TestLaMagaNoSorteaSiNoHayUnGrupo() {
	IDNuevoGrupo := int64(rand.Int())

	participantes, err := suite.maga.Sortear(IDNuevoGrupo)

	suite.Error(err, "Debería fallar al sortear en un grupo inexistente")
	suite.Nil(participantes, "No debería haber participantes si no hay grupo")
}

func (suite *LaMagaTestSuite) TestLaMagaNoSorteaSiYaSorteó() {
	IDNuevoGrupo := int64(rand.Int())
	suite.maga.NuevoGrupo(IDNuevoGrupo, "Mi grupo")
	IDNuevoParticipante := rand.Int()
	suite.maga.NuevoParticipante(IDNuevoGrupo, IDNuevoParticipante, "Nick")
	grupoDeLaDB := modelo.Grupo{Identificador: IDNuevoGrupo}
	suite.db.Where(grupoDeLaDB).First(&grupoDeLaDB)
	grupoDeLaDB.YaSorteo = true
	suite.db.Save(grupoDeLaDB)

	participantes, err := suite.maga.Sortear(IDNuevoGrupo)

	suite.NoError(err, "No debería fallar al sortear en un grupo que ya sorteó")
	suite.Nil(participantes[0].Amigx, "No debería haber asignado un amigx si ya había sorteado")
}

func (suite *LaMagaTestSuite) TestLaMagaNoSorteaSiHayUnSoloParticipante() {
	IDNuevoGrupo := int64(rand.Int())
	suite.maga.NuevoGrupo(IDNuevoGrupo, "Mi grupo")
	IDNuevoParticipante := rand.Int()
	suite.maga.NuevoParticipante(IDNuevoGrupo, IDNuevoParticipante, "Nick")

	participantes, err := suite.maga.Sortear(IDNuevoGrupo)

	suite.Error(err, "Debería fallar si hay un solo participante")
	suite.Nil(participantes, "No debería haber sorteado")
}

func (suite *LaMagaTestSuite) TestLaMagaSorteaAmigxs() {
	IDNuevoGrupo := int64(rand.Int())
	suite.maga.NuevoGrupo(IDNuevoGrupo, "Mi grupo")
	IDUnParticipante := rand.Int()
	suite.maga.NuevoParticipante(IDNuevoGrupo, IDUnParticipante, "Nick")
	IDOtroParticipante := rand.Int()
	suite.maga.NuevoParticipante(IDNuevoGrupo, IDOtroParticipante, "Nay")

	participantes, err := suite.maga.Sortear(IDNuevoGrupo)

	suite.NoError(err, "No debería fallar al sortear")
	suite.Equal(participantes[0].Amigx.Nombre, "Nay", "Nay debería ser amiga de Nick")
	suite.Equal(participantes[1].Amigx.Nombre, "Nick", "Nick debería ser amigo de Nay")

	grupoDeLaDB := modelo.Grupo{Identificador: IDNuevoGrupo}
	resultado := suite.db.Preload("Participantes").Where(grupoDeLaDB).First(&grupoDeLaDB)
	suite.NoError(resultado.Error, "Debería haber encontrado el grupo")
	suite.True(grupoDeLaDB.YaSorteo, "Debería estar sorteado")
	suite.Equal(*grupoDeLaDB.Participantes[0].AmigxID, grupoDeLaDB.Participantes[1].ID, "Nay debería ser amiga de Nick")
	suite.Equal(*grupoDeLaDB.Participantes[1].AmigxID, grupoDeLaDB.Participantes[0].ID, "Nick debería ser amigo de Nay")
}

func TestLaMagaTestSuite(t *testing.T) {
	suite.Run(t, new(LaMagaTestSuite))
}
