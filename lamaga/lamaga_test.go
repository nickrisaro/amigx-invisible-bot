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

	suite.Error(err, "Debería fallar al sortear en un grupo que ya sorteó")
	suite.Nil(participantes, "No debería haber participantes si ya había sorteado")
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

func (suite *LaMagaTestSuite) TestLaMagaTeDaLosParticipantesConSusAmigxs() {
	IDNuevoGrupo := int64(rand.Int())
	suite.maga.NuevoGrupo(IDNuevoGrupo, "Mi grupo")
	IDUnParticipante := rand.Int()
	suite.maga.NuevoParticipante(IDNuevoGrupo, IDUnParticipante, "Nick")
	IDOtroParticipante := rand.Int()
	suite.maga.NuevoParticipante(IDNuevoGrupo, IDOtroParticipante, "Nay")
	suite.maga.Sortear(IDNuevoGrupo)

	participantes, err := suite.maga.ParticipantesConAmigxs(IDNuevoGrupo)
	suite.NoError(err, "No debería fallar al buscar participantes y amigxs")
	suite.NotNil(participantes, "Debería haber participantes")
	suite.NotNil(participantes[0].Amigx, "Nick debería tener amigx")
	suite.NotNil(participantes[1].Amigx, "Nay debería tener amigx")
	suite.Equal(participantes[0].Amigx.Nombre, "Nay", "Nay debería ser amiga de Nick")
	suite.Equal(participantes[1].Amigx.Nombre, "Nick", "Nick debería ser amigo de Nay")
}

func (suite *LaMagaTestSuite) TestLaMagaNoTeDaLosParticipantesConSusAmigxsSiNoSorteaste() {
	IDNuevoGrupo := int64(rand.Int())
	suite.maga.NuevoGrupo(IDNuevoGrupo, "Mi grupo")
	IDUnParticipante := rand.Int()
	suite.maga.NuevoParticipante(IDNuevoGrupo, IDUnParticipante, "Nick")
	IDOtroParticipante := rand.Int()
	suite.maga.NuevoParticipante(IDNuevoGrupo, IDOtroParticipante, "Nay")

	participantes, err := suite.maga.ParticipantesConAmigxs(IDNuevoGrupo)

	suite.Error(err, "Debería fallar al buscar participantes y amigxs si no hizo el sorteo")
	suite.Nil(participantes, "No debería haber participantes si no hizo el sorteo")
}

func (suite *LaMagaTestSuite) TestLaMagaNoTeDaLosParticipantesConSusAmigxsSiNoHayGrupo() {
	IDNuevoGrupo := int64(rand.Int())

	participantes, err := suite.maga.ParticipantesConAmigxs(IDNuevoGrupo)

	suite.Error(err, "Debería fallar al buscar participantes y amigxs si no hay grupo")
	suite.Nil(participantes, "No debería haber participantes si no hay grupo")
}

func (suite *LaMagaTestSuite) TestLaMagaTeBorraUnGrupo() {
	IDNuevoGrupo := int64(rand.Int())
	suite.maga.NuevoGrupo(IDNuevoGrupo, "Mi grupo")

	err := suite.maga.Borrar(IDNuevoGrupo)

	suite.NoError(err, "No debería fallar al borrar un grupo")
	grupoDeLaDB := modelo.Grupo{Identificador: IDNuevoGrupo}
	resultado := suite.db.Where(grupoDeLaDB).First(&grupoDeLaDB)
	suite.Error(resultado.Error, "No debería haber encontrado el grupo")
}

func (suite *LaMagaTestSuite) TestLaMagaTeBorraUnGrupoYSusParticipantes() {
	IDNuevoGrupo := int64(rand.Int())
	suite.maga.NuevoGrupo(IDNuevoGrupo, "Mi grupo")
	grupoDeLaDB := modelo.Grupo{Identificador: IDNuevoGrupo}
	suite.db.Where(grupoDeLaDB).First(&grupoDeLaDB)
	idGrupoDB := grupoDeLaDB.ID
	IDUnParticipante := rand.Int()
	suite.maga.NuevoParticipante(IDNuevoGrupo, IDUnParticipante, "Nick")
	IDOtroParticipante := rand.Int()
	suite.maga.NuevoParticipante(IDNuevoGrupo, IDOtroParticipante, "Nay")

	err := suite.maga.Borrar(IDNuevoGrupo)

	suite.NoError(err, "No debería fallar al borrar un grupo")
	grupoDeLaDB = modelo.Grupo{Identificador: IDNuevoGrupo}
	resultado := suite.db.Where(grupoDeLaDB).First(&grupoDeLaDB)
	suite.Error(resultado.Error, "No debería haber encontrado el grupo")
	participantesDeLaDB := make([]*modelo.Participante, 0)
	resultado = suite.db.Where(&modelo.Participante{GrupoID: idGrupoDB}).Find(&participantesDeLaDB)
	suite.Equal(resultado.RowsAffected, int64(0), "No debería haber encontrado los participantes")
	suite.Empty(participantesDeLaDB, "No debería haber encontrado los participantes")
}

func (suite *LaMagaTestSuite) TestLaMagaNoBorraUnGrupoSiNoExiste() {
	IDNuevoGrupo := int64(rand.Int())

	err := suite.maga.Borrar(IDNuevoGrupo)

	suite.Error(err, "Debería fallar al borrar un grupo si no está creado")
}

func (suite *LaMagaTestSuite) TestLaMagaTeDiceEnQueGruposTeAnotaste() {
	IDNuevoGrupo := int64(rand.Int())
	suite.maga.NuevoGrupo(IDNuevoGrupo, "Mi grupo")
	IDUnParticipante := rand.Int()
	suite.maga.NuevoParticipante(IDNuevoGrupo, IDUnParticipante, "Nick")
	IDOtroParticipante := rand.Int()
	suite.maga.NuevoParticipante(IDNuevoGrupo, IDOtroParticipante, "Nay")
	suite.maga.Sortear(IDNuevoGrupo)
	IDOtroGrupo := int64(rand.Int())
	suite.maga.NuevoGrupo(IDOtroGrupo, "Mi otro grupo")
	suite.maga.NuevoParticipante(IDOtroGrupo, IDUnParticipante, "Nick")

	grupos, err := suite.maga.GruposDe(IDUnParticipante)

	suite.NoError(err, "No debería fallar al buscar grupos")
	suite.Len(grupos, 2, "Debería haber dos grupos")
	suite.Equal(IDNuevoGrupo, grupos[0].Identificador, "No coincide el Ientificador de Grupo")
	suite.Equal("Mi grupo", grupos[0].Nombre, "No coincide el nombre del Grupo")
	suite.Equal(IDOtroGrupo, grupos[1].Identificador, "No coincide el Ientificador de Grupo")
	suite.Equal("Mi otro grupo", grupos[1].Nombre, "No coincide el nombre del Grupo")
}

func (suite *LaMagaTestSuite) TestLaMagaTeDiceTodxsTusAmigxs() {
	IDNuevoGrupo := int64(rand.Int())
	suite.maga.NuevoGrupo(IDNuevoGrupo, "Mi grupo")
	IDUnParticipante := rand.Int()
	suite.maga.NuevoParticipante(IDNuevoGrupo, IDUnParticipante, "Nick")
	IDOtroParticipante := rand.Int()
	suite.maga.NuevoParticipante(IDNuevoGrupo, IDOtroParticipante, "Nay")
	suite.maga.Sortear(IDNuevoGrupo)
	IDOtroGrupo := int64(rand.Int())
	suite.maga.NuevoGrupo(IDOtroGrupo, "Mi otro grupo")
	suite.maga.NuevoParticipante(IDOtroGrupo, IDUnParticipante, "Nick")

	grupoAmigx, err := suite.maga.AmigxsDe(IDUnParticipante)

	suite.NoError(err, "No debería fallar al buscar amigxs")
	suite.Len(grupoAmigx, 1, "Debería haber un amigx")
	suite.Equal("Mi grupo", grupoAmigx[0].Grupo, "No coincide el nombre del Grupo")
	suite.Equal("Nay", grupoAmigx[0].Amigx, "No coincide el nombre del Amigx")
}

func TestLaMagaTestSuite(t *testing.T) {
	suite.Run(t, new(LaMagaTestSuite))
}
