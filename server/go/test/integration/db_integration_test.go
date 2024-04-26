package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	db_client "github.com/0xlunch/user-service/internal/db"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestDB(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Database Suite")
}

var _ = Describe("Database", Ordered, func() {
	var (
		container  tc.Container
		ctx        context.Context
		connString string
	)

	dbName := "users"
	dbUser := "user"
	dbPassword := "password"

	BeforeAll(func() {
		ctx = context.Background()
		// Start PostgreSQL container
		c, err := postgres.RunContainer(ctx,
			tc.WithImage("docker.io/postgres:16-alpine"),
			//postgres.WithInitScripts(filepath.Join("testdata", "init-user-db.sh")),
			//postgres.WithConfigFile(filepath.Join("testdata", "my-postgres.conf")),
			postgres.WithDatabase(dbName),
			postgres.WithUsername(dbUser),
			postgres.WithPassword(dbPassword),
			tc.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).WithStartupTimeout(5*time.Second)),
		)
		Expect(err).NotTo(HaveOccurred())

		connString, err = c.ConnectionString(ctx)
		Expect(err).NotTo(HaveOccurred())

		// set container reference
		container = c
	})

	// Cleanup
	AfterAll(func() {
		err := container.Terminate(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	When("fetching from the database", func() {

		It("should successfully connect to the database", func() {
			fmt.Println(connString)

			db, err := db_client.NewDB(connString)
			Expect(err).NotTo(HaveOccurred())
			Expect(db).NotTo(BeNil())
		})
	})

})
