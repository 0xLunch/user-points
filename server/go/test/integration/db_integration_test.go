package integration_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	db_client "github.com/0xlunch/user-service/internal/db"
	"github.com/google/uuid"
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
		container      tc.Container
		ctx            context.Context
		connString     string
		testDB         *db_client.DB
		loggedInUserId uuid.UUID
	)

	dbName := "user_points"
	dbUser := "user"
	dbPassword := "password"

	// Setup
	BeforeAll(func() {
		ctx = context.Background()
		// Start PostgreSQL testcontainer
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

		// ensure we have a db connection
		fmt.Println(connString)

		db, err := db_client.NewDB(connString)
		Expect(err).NotTo(HaveOccurred())
		Expect(db).NotTo(BeNil())

		// reference testDB
		testDB = db
	})

	// Cleanup
	AfterAll(func() {
		err := container.Terminate(ctx)
		Expect(err).NotTo(HaveOccurred())
		if testDB != nil {
			testDB.Pool.Close()
		}
	})

	// Test registering user
	When("registering users to the database", func() {

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

	// Test logging in
	When("logging user in", func() {

		It("should login user with correct password", func() {
			Expect(testDB).NotTo(BeNil())
			// login user
			ctx := context.Background()
			username := "username"
			password := "paragon"
			userID, err := testDB.LoginUser(ctx, username, password)
			Expect(err).NotTo(HaveOccurred())
			// save logged in user
			loggedInUserId = userID
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

	// Test points
	When("getting and updating user points", func() {

		// userPoints default to 0
		userPoints := 0

		It("should fetch user's points", func() {
			Expect(testDB).NotTo(BeNil())
			// ensure logged in user is saved
			Expect(loggedInUserId).NotTo(Equal(uuid.Nil))
			ctx := context.Background()

			// get points
			points, err := testDB.GetUserPoints(ctx, loggedInUserId)
			Expect(err).NotTo(HaveOccurred())
			Expect(points).To(Equal(userPoints))
		})

		It("should update user's points", func() {
			Expect(testDB).NotTo(BeNil())
			// ensure logged in user is saved
			Expect(loggedInUserId).NotTo(Equal(uuid.Nil))
			ctx := context.Background()

			// update points
			updatePoints := 500
			err := testDB.UpdateUserPoints(ctx, loggedInUserId, updatePoints)
			Expect(err).NotTo(HaveOccurred())

			// fetch updated points
			points, err := testDB.GetUserPoints(ctx, loggedInUserId)
			Expect(err).NotTo(HaveOccurred())
			Expect(points).To(Equal(updatePoints))
			// update userPoints reference
			userPoints = updatePoints
		})

		It("should add to user's points", func() {
			Expect(testDB).NotTo(BeNil())
			// ensure logged in user is saved
			Expect(loggedInUserId).NotTo(Equal(uuid.Nil))
			ctx := context.Background()

			// add points
			addPoints := 200
			err := testDB.AddUserPoints(ctx, loggedInUserId, addPoints)
			Expect(err).NotTo(HaveOccurred())
			// fetch updated points
			points, err := testDB.GetUserPoints(ctx, loggedInUserId)
			Expect(err).NotTo(HaveOccurred())
			Expect(points).To(Equal(userPoints + addPoints))
		})

		It("should not add zero or negative points", func() {
			Expect(testDB).NotTo(BeNil())
			// ensure logged in user is saved
			Expect(loggedInUserId).NotTo(Equal(uuid.Nil))
			ctx := context.Background()

			// add points
			addPoints := 0
			err := testDB.AddUserPoints(ctx, loggedInUserId, addPoints)
			Expect(err).To(HaveOccurred(), "points value must be >=1")
			addPoints--
			err = testDB.AddUserPoints(ctx, loggedInUserId, addPoints)
			Expect(err).To(HaveOccurred(), "points value must be >=1")
		})

		It("should not update user's points to negative", func() {
			Expect(testDB).NotTo(BeNil())
			// ensure logged in user is saved
			Expect(loggedInUserId).NotTo(Equal(uuid.Nil))
			ctx := context.Background()

			// update points
			updatePoints := -1
			err := testDB.UpdateUserPoints(ctx, loggedInUserId, updatePoints)
			Expect(err).To(HaveOccurred(), "points value must be positive")
		})

	})

})
