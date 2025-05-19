package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/Myriad-Dreamin/blog-backend/pkg/iou"
	"github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

func BackupBlog(src *sql.DB) error {
	info, available := debug.ReadBuildInfo()

	if !available {
		log.Println("Build info not available")
	}

	// make directory
	err := os.MkdirAll("./.data/backup/tmp", 0755)
	if err != nil {
		return errors.Wrap(err, "error create directory")
	}

	// write build info to file
	err = iou.WriteJsonToFile("./.data/backup/tmp/build-info.json", info)
	if err != nil {
		return errors.Wrap(err, "error write build info to file")
	}

	if src == nil {
		db, err := sql.Open("sqlite3", "./.data/blog.db")
		if err != nil {
			return errors.Wrap(err, "error open db")
		}
		defer db.Close()
		src = db
	}

	// open target db
	destDb, err := sql.Open("sqlite3", "./.data/backup/tmp/blog.db")
	if err != nil {
		return errors.Wrap(err, "error open target db")

	}
	defer destDb.Close()

	// backup db
	err = Backup(destDb, src)
	if err != nil {
		return errors.Wrap(err, "error backup db")
	}

	// close target db
	err = destDb.Close()
	if err != nil {
		return errors.Wrap(err, "error close target db")
	}

	// move target db to backup
	var timestamp = time.Now().Format("2006-01-02_15-04-05")
	err = os.Rename("./.data/backup/tmp", fmt.Sprintf("./.data/backup/%s", timestamp))
	if err != nil {
		return errors.Wrap(err, "error move target db to backup")
	}

	return nil
}

func Backup(destDb, srcDb *sql.DB) error {
	destConn, err := destDb.Conn(context.Background())
	if err != nil {
		return err
	}

	srcConn, err := srcDb.Conn(context.Background())
	if err != nil {
		return err
	}

	return destConn.Raw(func(destConn interface{}) error {
		return srcConn.Raw(func(srcConn interface{}) error {
			destSQLiteConn, ok := destConn.(*sqlite3.SQLiteConn)
			if !ok {
				return fmt.Errorf("can't convert destination connection to SQLiteConn")
			}

			srcSQLiteConn, ok := srcConn.(*sqlite3.SQLiteConn)
			if !ok {
				return fmt.Errorf("can't convert source connection to SQLiteConn")
			}

			b, err := destSQLiteConn.Backup("main", srcSQLiteConn, "main")
			if err != nil {
				return fmt.Errorf("error initializing SQLite backup: %w", err)
			}

			done, err := b.Step(-1)
			if !done {
				return fmt.Errorf("step of -1, but not done")
			}
			if err != nil {
				return fmt.Errorf("error in stepping backup: %w", err)
			}

			err = b.Finish()
			if err != nil {
				return fmt.Errorf("error finishing backup: %w", err)
			}

			return err
		})
	})
}
