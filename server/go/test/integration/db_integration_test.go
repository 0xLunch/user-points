package integration_test

import (
	"context"
	"fmt"
	"path/filepath"
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

	dbName := "user_points"
	dbUser := "user"
	dbPassword := "password"

	// Setup
	BeforeAll(func() {
		ctx = context.Background()
		// Start PostgreSQL container
		c, err := postgres.RunContainer(ctx,
			tc.WithImage("docker.io/postgres:16-alpine"),
			postgres.WithDatabase(dbName),
			postgres.WithUsername(dbUser),
			postgres.WithPassword(dbPassword),
			postgres.WithInitScripts(filepath.Join("scripts", "init.sql")),
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

	// Test DB connection and
	When("registering users to the database", func() {

		var testDB *db_client.DB

		// close pool connection after tests
		AfterAll(func() {
			if testDB != nil {
				testDB.Pool.Close()
			}
		})

		It("should successfully connect to the database", func() {
			fmt.Println(connString)

			db, err := db_client.NewDB(connString)
			Expect(err).NotTo(HaveOccurred())
			Expect(db).NotTo(BeNil())

			// reference testDB
			testDB = db
		})

		It("should register a user", func() {
			Expect(testDB).NotTo(BeNil())
			// register user
			ctx := context.Background()
			username := "username"
			password := "paragon"
			err := testDB.RegisterUser(ctx, username, password)
			Expect(err).NotTo(HaveOccurred())

			user := &db_client.User{}

			// confirm user entry
			err = testDB.Pool.QueryRow(ctx, `SELECT * FROM users_view WHERE username = $1`, username).
				Scan(&user.ID, &user.Username, &user.Points)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println(user)
			Expect(user.Username).To(Equal(username))
		})

		It("should not allow duplicate usernames", func() {
			Expect(testDB).NotTo(BeNil())
			// register same user as previous
			ctx := context.Background()
			username := "username"
			password := "paragon"
			err := testDB.RegisterUser(ctx, username, password)
			Expect(err).To(HaveOccurred(), "username already exists")
		})

	})

	When("logging user in", func() {
		var testDB *db_client.DB

		// close pool connection after tests
		AfterAll(func() {
			if testDB != nil {
				testDB.Pool.Close()
			}
		})

		It("should successfully connect to the database", func() {
			fmt.Println(connString)

			db, err := db_client.NewDB(connString)
			Expect(err).NotTo(HaveOccurred())
			Expect(db).NotTo(BeNil())

			// reference testDB
			testDB = db
		})

		It("should login user with correct password", func() {
			Expect(testDB).NotTo(BeNil())
			// login user
			ctx := context.Background()
			username := "username"
			password := "paragon"
			_, err := testDB.LoginUser(ctx, username, password)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should not login user with incorrect password", func() {
			Expect(testDB).NotTo(BeNil())
			// login user
			ctx := context.Background()
			username := "username"
			password := "parallel"
			_, err := testDB.LoginUser(ctx, username, password)
			Expect(err).To(HaveOccurred(), "invalid password")
		})

		It("should not login user with invalid username", func() {
			Expect(testDB).NotTo(BeNil())
			// login user
			ctx := context.Background()
			username := "usernamedoesnotexist"
			password := "parallel"
			_, err := testDB.LoginUser(ctx, username, password)
			Expect(err).To(HaveOccurred(), "invalid username")
		})
	})

})
