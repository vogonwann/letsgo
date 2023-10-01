package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
  "time"

	"snippetbox.janjique.com/internal/models"

  "github.com/alexedwards/scs/mysqlstore"
  "github.com/alexedwards/scs/v2"
  "github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
)

type application struct {
  errorLog        *log.Logger
  infoLog         *log.Logger
  snippets        *models.SnippetModel
  templateCache   map[string]*template.Template
  formDecoder     *form.Decoder
  sessionManager  *scs.SessionManager
}

func main() {
  // Define a new command-line flag with the name 'addr', a default value of ":4000"
  // and some short help text explaining what the flag controls. The value of the
  // flag will be stored in the addr variable at the runtime.
  addr := flag.String("addr", ":4000", "HTTP network address")
  dsn := flag.String("dns", "web:pass@/snippetbox?parseTime=true", "MariaDB data source name")

  // Importatntly, we use the flag.Parse() function to parse the command-line-flag.
  // This reads in the command-line flag value and assigns it to the addr
  // variable. You need to call this *before* you use the addr variable
  // otherwise it will always contain the default
  flag.Parse()

  infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
  errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

  // To keep the main funvtion tidy I've put the code for creating a connection
  // pool into the separate openDB() function below. We pass openDB the DSN
  // from the command-line flag.
  db, err := openDB(*dsn)
  if err != nil {
    errorLog.Fatal(err)
  }

  // We also defer a call to db.Close(), so that the connection pool is closed
  // before the main() function exits.
  defer db.Close()

  // Initialize a new template cache
  templateCache, err := newTemplateCache()
  if err != nil {
    errorLog.Fatal(err)
  }

  formDecoder := form.NewDecoder()

  sessionManager := scs.New()
  sessionManager.Store = mysqlstore.New(db)
  sessionManager.Lifetime = 12 * time.Hour

  app := &application{
    errorLog:       errorLog,
    infoLog:        infoLog,
    snippets:       &models.SnippetModel{DB: db},
    templateCache:  templateCache,
    formDecoder:    formDecoder,
    sessionManager: sessionManager,
  }


  srv := &http.Server{
    Addr:     *addr,
    ErrorLog: errorLog,
    Handler:  app.routes(),
  }

  infoLog.Printf("Starting server on %s", *addr)
  err = srv.ListenAndServe()
  errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
  db, err := sql.Open("mysql", dsn)
  if err != nil {
    return nil, err
  }

  if err = db.Ping(); err != nil {
    return nil, err
  }
  return db, nil
}
