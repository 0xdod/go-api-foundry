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

	// 1. Start Postgres Container
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

	// 2. Run Migrations
	wd, err := os.Getwd()
	suite.Require().NoError(err)

	// Try multiple paths for migrations
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

	// 3. Connect to DB
	suite.logger = log.NewLoggerWithJSONOutput()
	suite.db = pgClient.NewDB(connStr, suite.logger)
	err = suite.db.Connect(ctx)
	suite.Require().NoError(err)

	// 4. Setup Router and Server
	suite.appConfig = &config.ApplicationConfig{
		PGDB:   suite.db,
		Logger: suite.logger,
	}

	suite.appConfig.RouterService = router.CreateRouterService(suite.logger, nil, &router.RouterConfig{
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
		RequestTimeout:    30 * time.Second,
	})

	// Mount Controllers
	// 1. Account Controller (to create user accounts)
	accountFactory := account.NewServiceFactory(suite.appConfig.PGDB, suite.appConfig.Logger)
	suite.appConfig.RouterService.MountController(accountFactory.CreateController())

	// 2. Transaction Controller
	// We need to construct this manually or use a factory if one exists.
	// Step 174 showed transaction/factory.go created. Let's use it if possible.
	// Assuming NewServiceFactory exists in transaction package.
	transactionFactory := transaction.NewServiceFactory(suite.appConfig.PGDB, suite.appConfig.Logger)
	suite.appConfig.RouterService.MountController(transactionFactory.CreateController())

	suite.server = httptest.NewServer(suite.appConfig.RouterService.GetEngine())
	suite.baseURL = suite.server.URL

	// Seed User
	// We need to seed a user in the users table because accounts referencing user_id need it
	// Assuming users table schema: id, email, ...
	// Since I don't have the User model/sqlc easily accessible here, I'll direct SQL exec.
	// The migration for accounts adds user_id FK, but maybe it's nullable or we need to respect it.
	// Wait, migration 000003 adds user_id. Is it FK? "ADD COLUMN user_id BIGINT NULL". No FK constraint explicit in the `ADD COLUMN` line shown in Step 92.
	// But let's check if there is a users table.
	// Step 88 shows waitlist_entries.
	// Step 89 shows accounts, transactions, ledger_entries, balances.
	// User table might be missing from the provided context or migrations listing?
	// Ah, I see "000001_init.up.sql" has waitlist_entries.
	// Maybe "users" table is not shown or I missed it.
	// If `user_id` in accounts does not have references constraint, I can insert any ID.
	// Step 92: `ADD COLUMN user_id BIGINT NULL;` - no REFERENCES clause.
	// So I can use any integer.
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
	// 1. Create User Account
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

	// 2. Deposit
	reqDeposit := map[string]interface{}{
		"account_id": accountID,
		"amount":     50.00, // $50.00
		"reference":  "dep-" + uuid.New().String(),
	}
	jsonDeposit, _ := json.Marshal(reqDeposit)

	httpReq, _ := http.NewRequest(http.MethodPost, suite.baseURL+"/v1/transactions/deposit", bytes.NewBuffer(jsonDeposit))
	httpReq.Header.Set("Idempotency-Key", "idem-dep-1")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-ID", "100") // Middleware requires User-ID? Usually Idempotency needs user_id context.

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

	// 3. Verify Balance
	// Get Account Balance
	resp, err = http.Get(fmt.Sprintf("%s/v1/accounts/%d/balance", suite.baseURL, accountID))
	suite.Require().NoError(err)
	defer resp.Body.Close()

	var balResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&balResp)

	balData := balResp["data"].(map[string]interface{})
	// 50.00 * 100 = 5000 cents
	suite.Equal(float64(5000), balData["balance"])
}

func (suite *TransactionAPITestSuite) TestWithdraw() {
	// 1. Create User Account
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

	// 2. Deposit Funds First (so we can withdraw)
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

	// 3. Withdraw
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

	// 4. Verify Balance (100 - 50 = 50 => 5000 cents)
	resp, err = http.Get(fmt.Sprintf("%s/v1/accounts/%d/balance", suite.baseURL, accountID))
	suite.Require().NoError(err)
	defer resp.Body.Close()

	var balResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&balResp)
	balData := balResp["data"].(map[string]interface{})
	suite.Equal(float64(5000), balData["balance"])
}

func TestTransactionAPISuite(t *testing.T) {
	suite.Run(t, new(TransactionAPITestSuite))
}
