package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
	"github.com/vagonaizer/authenitfication-service/internal/config"
)

type Migration struct {
	Version  int
	Name     string
	Filename string
	Content  string
}

func main() {
	var (
		direction = flag.String("direction", "up", "Migration direction: up or down")
		status    = flag.Bool("status", false, "Show migration status")
	)
	flag.Parse()

	// Загружаем .env файл
	loadEnvFile()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Выводим информацию о подключении для отладки
	log.Printf("Connecting to database: host=%s port=%s user=%s dbname=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Name)

	db, err := connectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := createMigrationsTable(db); err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	if *status {
		if err := showStatus(db, cfg.Database.MigrationsPath); err != nil {
			log.Fatalf("Failed to show status: %v", err)
		}
		return
	}

	switch *direction {
	case "up":
		if err := migrateUp(db, cfg.Database.MigrationsPath); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
	case "down":
		if err := migrateDown(db, cfg.Database.MigrationsPath); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
	default:
		log.Fatalf("Invalid direction: %s. Use 'up' or 'down'", *direction)
	}
}

// loadEnvFile загружает переменные из .env файла
func loadEnvFile() {
	file, err := os.Open(".env")
	if err != nil {
		log.Printf("Warning: .env file not found: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Устанавливаем переменную окружения только если она еще не установлена
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading .env file: %v", err)
	}
}

func connectDB(cfg *config.Config) (*sql.DB, error) {
	// Принудительно используем IPv4 для localhost
	host := cfg.Database.Host
	if host == "localhost" {
		host = "127.0.0.1"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	log.Printf("DSN (without password): host=%s port=%s user=%s dbname=%s sslmode=%s",
		host, cfg.Database.Port, cfg.Database.User, cfg.Database.Name, cfg.Database.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.Exec(query)
	return err
}

func loadMigrations(migrationsPath string) ([]Migration, error) {
	var migrations []Migration

	err := filepath.WalkDir(migrationsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		filename := d.Name()
		parts := strings.SplitN(filename, "_", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid migration filename format: %s", filename)
		}

		var version int
		if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
			return fmt.Errorf("invalid version in filename %s: %v", filename, err)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %v", path, err)
		}

		name := strings.TrimSuffix(parts[1], ".sql")
		migrations = append(migrations, Migration{
			Version:  version,
			Name:     name,
			Filename: filename,
			Content:  string(content),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func getAppliedMigrations(db *sql.DB) (map[int]bool, error) {
	applied := make(map[int]bool)

	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

func migrateUp(db *sql.DB, migrationsPath string) error {
	migrations, err := loadMigrations(migrationsPath)
	if err != nil {
		return err
	}

	applied, err := getAppliedMigrations(db)
	if err != nil {
		return err
	}

	var executed int
	for _, migration := range migrations {
		if applied[migration.Version] {
			continue
		}

		log.Printf("Applying migration %d: %s", migration.Version, migration.Name)

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %v", err)
		}

		if _, err := tx.Exec(migration.Content); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %d: %v", migration.Version, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migration.Version); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %v", migration.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %v", migration.Version, err)
		}

		executed++
		log.Printf("Migration %d applied successfully", migration.Version)
	}

	if executed == 0 {
		log.Println("No migrations to apply")
	} else {
		log.Printf("Applied %d migrations", executed)
	}

	return nil
}

func migrateDown(db *sql.DB, migrationsPath string) error {
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return err
	}

	if len(applied) == 0 {
		log.Println("No migrations to rollback")
		return nil
	}

	// Find the latest applied migration
	var latestVersion int
	for version := range applied {
		if version > latestVersion {
			latestVersion = version
		}
	}

	log.Printf("Rolling back migration %d", latestVersion)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = $1", latestVersion); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove migration record %d: %v", latestVersion, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback %d: %v", latestVersion, err)
	}

	log.Printf("Migration %d rolled back successfully", latestVersion)
	log.Println("Note: This tool only removes the migration record. Manual schema changes may be required.")

	return nil
}

func showStatus(db *sql.DB, migrationsPath string) error {
	migrations, err := loadMigrations(migrationsPath)
	if err != nil {
		return err
	}

	applied, err := getAppliedMigrations(db)
	if err != nil {
		return err
	}

	fmt.Println("Migration Status:")
	fmt.Println("================")

	if len(migrations) == 0 {
		fmt.Println("No migrations found")
		return nil
	}

	for _, migration := range migrations {
		status := "PENDING"
		if applied[migration.Version] {
			status = "APPLIED"
		}
		fmt.Printf("%03d %-50s %s\n", migration.Version, migration.Name, status)
	}

	appliedCount := len(applied)
	totalCount := len(migrations)
	fmt.Printf("\nSummary: %d/%d migrations applied\n", appliedCount, totalCount)

	return nil
}
