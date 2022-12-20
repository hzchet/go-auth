package dbuser

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"server/internal/utils"
)

const usersTable = "users"

type User struct {
	Email     string    `db:"email"`
	Password  string    `db:"password"`
}

func (d *Database) CreateUser(ctx context.Context, user *utils.User) (string, error) {
	dbUser := userFromService(user)

	sqlText, bound, err := squirrel.Insert(usersTable).Columns(
		"email",
		"password").Values(
		dbUser.Email,
		dbUser.Password).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("cannot build sql: %w", err)
	}

	_, err = d.db.ExecContext(ctx, sqlText, bound...)
	if err != nil {
		return "", fmt.Errorf("cannot insert user: %w", err)
	}

	return user.Email, nil
}

func (d *Database) GetUserByEmail(ctx context.Context, email string) (utils.User, error) {
	res := User{}

	query := squirrel.Select(
		"email",
		"password").
		From(usersTable).
		Where(squirrel.Eq{"email": email}).
		PlaceholderFormat(squirrel.Dollar)

	sqlText, bound, err := query.ToSql()
	if err != nil {
		return utils.User{}, fmt.Errorf("cannot build SQL: %w", err)
	}

	if err = d.db.GetContext(ctx, &res, sqlText, bound...); err != nil {
		return utils.User{}, fmt.Errorf("cannot get user: %w", err)
	}

	return res.ToService(), nil
}

func (u *User) ToService() utils.User {
	return utils.User{
		Email:     u.Email,
		Password:  u.Password,
	}
}

func userFromService(u *utils.User) User {
	return User{
		Email:     u.Email,
		Password:  u.Password,
	}
}
