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

func TestLaMagaTestSuite(t *testing.T) {
	suite.Run(t, new(LaMagaTestSuite))
}
