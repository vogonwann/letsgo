package models

import (
  "database/sql"
  "time"
  "errors"
)

// Define a Snippet type to hold the data for an individual snippet. Notice how
// the fields of the struct correspond to the fields in our MySQL snippets
// table
type Snippet struct {
  ID      int
  Title   string
  Content string
  Created time.Time
  Expires time.Time
}

// Define a SnippetModel type which wraps a sql.DB connection pool.
type SnippetModel struct {
  DB *sql.DB
}

// This will insert new snippet into the database.
func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
  // Write the SQL statement we want to execute. I've split it over two lines
  // for readability (which is why it's surrounded with backquotes instead
  // of normal double quotes).
  stmt := `INSERT INTO snippets (title, content, created, expires)
  VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

  // Use the Exec() method on the embedded connection pool to execute the
  // statement. The first parameter is the SQL statement, followed by the
  // title, content and expiry values for the placeholder parameters
  result, err := m.DB.Exec(stmt, title, content, expires)
  if (err != nil) {
    return 0, err
  }

  // Use the LastInsertId() method on the result to get the ID of our
  // newly inserted record in the snippets table
  id, err := result.LastInsertId()
  if (err != nil) {
    return 0, err
  }

  return int(id), nil
}

// This will return a specific snippet based on its id.
func (m *SnippetModel) Get(id int) (*Snippet, error) {
  // Write the SQL statement we want to execute
  stmt := `SELECT id, title, content, created, expires FROM snippets
  WHERE expires > UTC_TIMESTAMP() AND id = ?`

  // Use the QueryRow() method on the connection pool to execute our
  // SQL statement. Returns a pointer to sql.Row object which holds
  // result from the database.
  row := m.DB.QueryRow(stmt, id)

  // Initialize a pointer to a new zeroed Snippet struct (empty constructor???)
  s := &Snippet{}

  // Use the row.Scan() to copy the values from each field in sql.Row to the
  // corresponding field in the Snippet struct (map???). Number of arguments
  // must be the same as number of columns returned by SQL statement
  err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
  if (err != nil) {
    // If the query returns no rows, then row.Scan() will return a
    // sql.ErrNoRows error. We use the errors.Is() function check for that
    // error specifically, and return our own ErrNoRecord error
    // instead (we'll create this in a moment).
    if (errors.Is(err, sql.ErrNoRows)) {
      return nil, ErrNoRecord
    } else {
      return nil, err
    }
  }

  // If everything went OK then return the Snippet object.
  return s, nil
}

// This will return the 10 most recently created snippets
func (m *SnippetModel) Latest() ([]*Snippet, error) {
  // SQL statement to execute
  stmt := `SELECT id, title, content, created, expires FROM snippets
  WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10`

  // Execute statement on connection pool
  rows, err := m.DB.Query(stmt)
  if err != nil {
    return nil, err
  }

  // Defer rows.Close() after we check for error
  // otherwise we will get panic trying to close nil result set
  defer rows.Close()

  // Initialize an empty slice to hold the Snippet struct
  snippets := []*Snippet{}

  // Iterate through the rows in the result set preparing each one
  // of them to acted on by rows.Scan() method. Resultset is automatically
  // closes itself when iteration is over and frees-up underlaying database
  // connection
  for rows.Next() {
    // Create a pointer to a new zeroed Snippet struct
    s := &Snippet{}

    // Use rows.Scan() to copy values from rows to the new Snippet object
    // Number of arguments must be same as number of columns in resultset
    err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
    if err != nil {
      return nil, err
    }

    // Append it to the slice of snippets
    snippets = append(snippets, s)
  }

  // When the rows.Next() loop has finished we call rows.Err() to retrieve any
  // error that was encountered during the iteration. It's important to
  // call this - don't assume that a successful iteration was completed
  // over the whole resultset.
  if err = rows.Err(); err != nil {
    return nil, err
  }

  // If everything went OK then return the Snippets slice.
  return snippets, nil
}
