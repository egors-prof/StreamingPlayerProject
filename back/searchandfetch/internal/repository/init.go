package repository

import (
	
	"fmt"

	"github.com/egors-prof/searchService/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)



type Repository struct{
	DB *sqlx.DB
}


func InitRepository(cfg *config.Config)(*Repository,error){
	connectStr := fmt.Sprintf(
		`host=%s
			user=%s
			password=%s
			dbname=%s
			sslmode=disable`,
		cfg.PostgresConfig.PostgresHost,
		cfg.PostgresConfig.PostgresUser,
		cfg.PostgresConfig.PostgresPassword,
		cfg.PostgresConfig.PostgresDatabase,
	)
	
	db,err:=sqlx.Open("postgres",connectStr)
	if err!=nil{
		return &Repository{},err
	}
	return &Repository{DB:db},nil
}