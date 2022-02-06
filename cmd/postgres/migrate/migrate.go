package migrate

import (
	"embed" //nolint:golint,nolintlint // we should use that

	"github.com/d7561985/pb-ab/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	//nolint:golint,nolintlint // we should use that
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

const (
	dbConnect    = "postgresql://postgres@localhost/db?sslmode=disable"
	EnvMongoAddr = "POSTGRES_ADDR"
	fAddr        = "addr"
)

//go:embed migration/*.sql
var fs embed.FS

func New() *Migrate { return &Migrate{} }

type Migrate struct{}

func (m *Migrate) Cli() *cli.Command {
	return &cli.Command{
		Name:    "migrate",
		Aliases: []string{"m"},
		Usage:   "perform migration Job, should be executed before first launch of every new version",
		Action:  m.Action,
		Flags:   []cli.Flag{&cli.StringFlag{Name: fAddr, Value: dbConnect, EnvVars: []string{EnvMongoAddr}}}}
}

func (*Migrate) Action(cxt *cli.Context) error {
	cfg := config.Postgres{Addr: cxt.String(fAddr)}

	d, err := iofs.New(fs, "migration")
	if err != nil {
		return errors.WithStack(err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, cfg.Addr)
	if err != nil {
		return errors.WithStack(err)
	}

	// // or m.Step(2) if you want to explicitly set the number of migrations to run
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return errors.WithStack(err)
	}

	println("migration completed")

	return nil
}
