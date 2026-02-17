package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/akeren/go-api-foundry/config"
	"github.com/akeren/go-api-foundry/config/router"
	"github.com/akeren/go-api-foundry/domain/account"
	"github.com/akeren/go-api-foundry/internal/log"
	pgClient "github.com/akeren/go-api-foundry/internal/postgres"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type AccountAPITestSuite struct {
	suite.Suite
	pgContainer *postgres.PostgresContainer
	db          *pgClient.DB
	server      *httptest.Server
	baseURL     string
	logger      *log.Logger
	appConfig   *config.ApplicationConfig
}

func (suite *AccountAPITestSuite) SetupSuite() {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		suite.T().Skip("Skipping integration tests. Set RUN_INTEGRATION_TESTS=true to run them")
	}

	ctx := context.Background()
	var err error

	dbName := "testdb_accounts"
	dbUser := "testuser"
	dbPassword := "testpassword"

	suite.pgContainer, err = postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	suite.Require().NoError(err)

	connStr, err := suite.pgContainer.ConnectionString(ctx, "sslmode=disable")
	suite.Require().NoError(err)

	wd, err := os.Getwd()
	suite.Require().NoError(err)

	migrationsPath := filepath.Join(wd, "..", "migrations")
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		migrationsPath = filepath.Join(wd, "migrations")
	}
	absPath, err := filepath.Abs(migrationsPath)
	suite.Require().NoError(err)

	m, err := migrate.New(
		"file://"+absPath,
		connStr,
	)
	suite.Require().NoError(err)
	err = m.Up()
	suite.Require().NoError(err)

	suite.logger = log.NewLoggerWithJSONOutput()
	suite.db = pgClient.NewDB(connStr, suite.logger)
	err = suite.db.Connect(ctx)
	suite.Require().NoError(err)
	suite.appConfig = &config.ApplicationConfig{
		DB:     suite.db,
		Logger: suite.logger,
	}

	suite.appConfig.RouterService = router.CreateRouterService(suite.logger, nil, &router.RouterConfig{
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
		RequestTimeout:    30 * time.Second,
	})

	accountFactory := account.NewServiceFactory(suite.appConfig.DB, suite.appConfig.Logger)
	suite.appConfig.RouterService.MountController(accountFactory.CreateController())

	suite.server = httptest.NewServer(suite.appConfig.RouterService.GetEngine())
	suite.baseURL = suite.server.URL
}

func (suite *AccountAPITestSuite) TearDownSuite() {
	if suite.pgContainer != nil {
		suite.pgContainer.Terminate(context.Background())
	}
	if suite.server != nil {
		suite.server.Close()
	}
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *AccountAPITestSuite) TestCreateAccount() {
	reqAccount := map[string]interface{}{
		"type":     "user",
		"name":     "Test Account",
		"code":     "ACC-001",
		"currency": "USD",
		"user_id":  200,
	}
	jsonBody, _ := json.Marshal(reqAccount)
	resp, err := http.Post(suite.baseURL+"/v1/accounts", "application/json", bytes.NewBuffer(jsonBody))
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusCreated, resp.StatusCode)

	var accResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&accResp)
	suite.Require().NoError(err)

	accData := accResp["data"].(map[string]interface{})
	suite.Equal("Test Account", accData["name"])
	suite.Equal("ACC-001", accData["code"])
	suite.Equal("USD", accData["currency"])
	// JSON numbers are float64
	suite.Equal(float64(200), accData["user_id"])
	suite.NotZero(accData["id"])
}

func (suite *AccountAPITestSuite) TestGetAccount() {
	// Create first
	reqAccount := map[string]interface{}{
		"type":     "user",
		"name":     "Fetch Account",
		"code":     "ACC-002",
		"currency": "EUR",
		"user_id":  201,
	}
	jsonBody, _ := json.Marshal(reqAccount)
	resp, _ := http.Post(suite.baseURL+"/v1/accounts", "application/json", bytes.NewBuffer(jsonBody))

	var createResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createResp)
	resp.Body.Close()
	createData := createResp["data"].(map[string]interface{})
	id := uint64(createData["id"].(float64))

	// Fetch
	resp, err := http.Get(fmt.Sprintf("%s/v1/accounts/%d", suite.baseURL, id))
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	var getResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&getResp)
	suite.Require().NoError(err)

	getData := getResp["data"].(map[string]interface{})
	suite.Equal("Fetch Account", getData["name"])
	suite.Equal("ACC-002", getData["code"])
}

func (suite *AccountAPITestSuite) TestGetAccountNotFound() {
	resp, err := http.Get(fmt.Sprintf("%s/v1/accounts/%d", suite.baseURL, 99999))
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusNotFound, resp.StatusCode)
}

func (suite *AccountAPITestSuite) TestCreateAccountDuplicateCode() {
	reqAccount := map[string]interface{}{
		"type":     "user",
		"name":     "Dup Account 1",
		"code":     "ACC-DUP",
		"currency": "USD",
		"user_id":  202,
	}
	jsonBody, _ := json.Marshal(reqAccount)
	http.Post(suite.baseURL+"/v1/accounts", "application/json", bytes.NewBuffer(jsonBody))

	resp, err := http.Post(suite.baseURL+"/v1/accounts", "application/json", bytes.NewBuffer(jsonBody))
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusConflict, resp.StatusCode)
}

func TestAccountAPISuite(t *testing.T) {
	suite.Run(t, new(AccountAPITestSuite))
}
