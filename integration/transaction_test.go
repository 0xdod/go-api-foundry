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
	"github.com/akeren/go-api-foundry/domain/transaction"
	"github.com/akeren/go-api-foundry/internal/log"
	pgClient "github.com/akeren/go-api-foundry/internal/postgres"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TransactionAPITestSuite struct {
	suite.Suite
	pgContainer *postgres.PostgresContainer
	db          *pgClient.DB
	server      *httptest.Server
	baseURL     string
	logger      *log.Logger
	appConfig   *config.ApplicationConfig
}

func (suite *TransactionAPITestSuite) SetupSuite() {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		suite.T().Skip("Skipping integration tests. Set RUN_INTEGRATION_TESTS=true to run them")
	}

	ctx := context.Background()
	var err error

	dbName := "testdb"
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
	transactionFactory := transaction.NewServiceFactory(suite.appConfig.DB, suite.appConfig.Logger)
	suite.appConfig.RouterService.MountController(transactionFactory.CreateController())

	suite.server = httptest.NewServer(suite.appConfig.RouterService.GetEngine())
	suite.baseURL = suite.server.URL
}

func (suite *TransactionAPITestSuite) TearDownSuite() {
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

func (suite *TransactionAPITestSuite) TestDeposit() {
	reqAccount := map[string]interface{}{
		"type":     "user",
		"name":     "Deposit Test Account",
		"code":     "DEP-001",
		"currency": "USD",
		"user_id":  100,
	}
	jsonBody, _ := json.Marshal(reqAccount)
	resp, err := http.Post(suite.baseURL+"/v1/accounts", "application/json", bytes.NewBuffer(jsonBody))
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)

	var accResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&accResp)
	resp.Body.Close()
	accData := accResp["data"].(map[string]interface{})
	accountID := uint64(accData["id"].(float64))

	reqDeposit := map[string]interface{}{
		"account_id": accountID,
		"amount":     50.00,
		"reference":  "dep-" + uuid.New().String(),
	}
	jsonDeposit, _ := json.Marshal(reqDeposit)

	httpReq, _ := http.NewRequest(http.MethodPost, suite.baseURL+"/v1/transactions/deposit", bytes.NewBuffer(jsonDeposit))
	httpReq.Header.Set("Idempotency-Key", "idem-dep-1")
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(httpReq)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusCreated, resp.StatusCode)

	var txnResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&txnResp)
	suite.Require().NoError(err)

	txnData := txnResp["data"].(map[string]interface{})
	suite.Equal("deposit", txnData["type"])

	resp, err = http.Get(fmt.Sprintf("%s/v1/accounts/%d/balance", suite.baseURL, accountID))
	suite.Require().NoError(err)
	defer resp.Body.Close()

	var balResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&balResp)

	balData := balResp["data"].(map[string]interface{})
	suite.Equal(float64(50), balData["balance"])
}

func (suite *TransactionAPITestSuite) TestWithdraw() {
	reqAccount := map[string]interface{}{
		"type":     "user",
		"name":     "Withdraw Test Account",
		"code":     "WD-001",
		"currency": "USD",
		"user_id":  101,
	}
	jsonBody, _ := json.Marshal(reqAccount)
	resp, err := http.Post(suite.baseURL+"/v1/accounts", "application/json", bytes.NewBuffer(jsonBody))
	suite.Require().NoError(err)

	var accResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&accResp)
	resp.Body.Close()
	accData := accResp["data"].(map[string]interface{})
	accountID := uint64(accData["id"].(float64))

	reqDeposit := map[string]interface{}{
		"account_id": accountID,
		"amount":     100.00,
		"reference":  "dep-wd-" + uuid.New().String(),
	}
	jsonDeposit, _ := json.Marshal(reqDeposit)
	httpReqDep, _ := http.NewRequest(http.MethodPost, suite.baseURL+"/v1/transactions/deposit", bytes.NewBuffer(jsonDeposit))
	httpReqDep.Header.Set("Idempotency-Key", "idem-wd-setup")
	httpReqDep.Header.Set("Content-Type", "application/json")
	httpReqDep.Header.Set("User-ID", "101")
	client := &http.Client{}
	resp, _ = client.Do(httpReqDep)
	resp.Body.Close()

	reqWithdraw := map[string]interface{}{
		"account_id": accountID,
		"amount":     50.00,
		"reference":  "wd-" + uuid.New().String(),
	}
	jsonWithdraw, _ := json.Marshal(reqWithdraw)

	httpReq, _ := http.NewRequest(http.MethodPost, suite.baseURL+"/v1/transactions/withdraw", bytes.NewBuffer(jsonWithdraw))
	httpReq.Header.Set("Idempotency-Key", "idem-wd-1")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-ID", "101")

	resp, err = client.Do(httpReq)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusCreated, resp.StatusCode)

	resp, err = http.Get(fmt.Sprintf("%s/v1/accounts/%d/balance", suite.baseURL, accountID))
	suite.Require().NoError(err)
	defer resp.Body.Close()

	var balResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&balResp)
	balData := balResp["data"].(map[string]interface{})
	suite.Equal(float64(50), balData["balance"])
}

func (suite *TransactionAPITestSuite) TestGetHistory() {
	reqAccount := map[string]interface{}{
		"type":     "user",
		"name":     "History Test Account",
		"code":     "HIST-001",
		"currency": "USD",
		"user_id":  102,
	}
	jsonBody, _ := json.Marshal(reqAccount)
	resp, _ := http.Post(suite.baseURL+"/v1/accounts", "application/json", bytes.NewBuffer(jsonBody))

	var accResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&accResp)
	resp.Body.Close()
	accData := accResp["data"].(map[string]interface{})
	accountID := uint64(accData["id"].(float64))

	reqDeposit := map[string]interface{}{
		"account_id": accountID,
		"amount":     200.00,
		"reference":  "dep-hist-" + uuid.New().String(),
	}
	jsonDeposit, _ := json.Marshal(reqDeposit)
	httpReqDep, _ := http.NewRequest(http.MethodPost, suite.baseURL+"/v1/transactions/deposit", bytes.NewBuffer(jsonDeposit))
	client := &http.Client{}
	client.Do(httpReqDep)

	reqWithdraw := map[string]interface{}{
		"account_id": accountID,
		"amount":     50.00,
		"reference":  "wd-hist-" + uuid.New().String(),
	}
	jsonWithdraw, _ := json.Marshal(reqWithdraw)
	httpReqWd, _ := http.NewRequest(http.MethodPost, suite.baseURL+"/v1/transactions/withdraw", bytes.NewBuffer(jsonWithdraw))
	client.Do(httpReqWd)

	resp, err := http.Get(fmt.Sprintf("%s/v1/transactions?account_id=%d", suite.baseURL, accountID))
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	var histResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&histResp)
	suite.Require().NoError(err)

	entries := histResp["data"].([]interface{})

	suite.Len(entries, 2)
}

func TestTransactionAPISuite(t *testing.T) {
	suite.Run(t, new(TransactionAPITestSuite))
}
