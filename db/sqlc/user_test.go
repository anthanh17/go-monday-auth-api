package db

import (
	"context"
	"database/sql"
	"monday-auth-api/util"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createRamdomUser(t *testing.T) User {
	arg := CreateUserParams{
		UserName: util.RandomString(5),
		FullName: util.RandomString(10),
		Mail:     util.RandomEmail(),
		Role:     "user",
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.UserName, user.UserName)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Mail, user.Mail)
	require.Equal(t, arg.Role, user.Role)

	require.NotZero(t, user.ID)
	require.NotZero(t, user.CreatedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	createRamdomUser(t)
}

func TestGetUser(t *testing.T) {
	userCreate := createRamdomUser(t)
	userGet, err := testQueries.GetUser(context.Background(), userCreate.ID)

	require.NoError(t, err)
	require.NotEmpty(t, userGet)

	require.Equal(t, userCreate.ID, userGet.ID)
	require.Equal(t, userCreate.UserName, userGet.UserName)
	require.Equal(t, userCreate.FullName, userGet.FullName)
	require.Equal(t, userCreate.Mail, userGet.Mail)
	require.Equal(t, userCreate.Role, userGet.Role)
	require.WithinDuration(t, userCreate.CreatedAt, userGet.CreatedAt, time.Second)
}

func TestUpdateGmailUser(t *testing.T) {
	userCreate := createRamdomUser(t)

	arg := UpdateGmailUserParams{
		ID:   userCreate.ID,
		Mail: util.RandomEmail(),
	}
	userGet, err := testQueries.UpdateGmailUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, userGet)

	// Update mail
	require.Equal(t, arg.Mail, userGet.Mail)

	require.Equal(t, userCreate.ID, userGet.ID)
	require.Equal(t, userCreate.UserName, userGet.UserName)
	require.Equal(t, userCreate.FullName, userGet.FullName)
	require.Equal(t, userCreate.Role, userGet.Role)
	require.WithinDuration(t, userCreate.CreatedAt, userGet.CreatedAt, time.Second)
}

func TestDeletelUser(t *testing.T) {
	userCreate := createRamdomUser(t)

	err := testQueries.DeleteUser(context.Background(), userCreate.ID)
	require.NoError(t, err)

	userGet, err := testQueries.GetUser(context.Background(), userCreate.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, userGet)
}

func TestListUsers(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRamdomUser(t)
	}

	arg := ListUsersParams{
		Limit:  5,
		Offset: 5,
	}

	listUsers, err := testQueries.ListUsers(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, listUsers, 5)

	for _, user := range listUsers {
		require.NotEmpty(t, user)
	}
}
