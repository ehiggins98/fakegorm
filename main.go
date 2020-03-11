package gorm

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

type expectation struct {
	Function string
	Params   []interface{}
	Output   []interface{}
	Error    error
}

type DB struct {
	db           *sql.DB
	Error        error
	expectations []expectation
	index        int
	testErrors   []error
}

type Model struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func Open(dialect string, args ...interface{}) (*DB, error) {
	db, _, err := sqlmock.New()
	return &DB{db: db}, err
}

func (it *DB) ExpectationsMet() error {
	if len(it.testErrors) > 0 {
		return it.testErrors[0]
	} else if it.index < len(it.expectations) {
		return errors.New("Not all expectations met")
	} else {
		return nil
	}
}

func (it *DB) ExpectCall(fn string) *expectation {
	it.expectations = append(it.expectations, expectation{Function: fn})
	return &it.expectations[len(it.expectations)-1]
}

func (e *expectation) WithParams(params ...interface{}) *expectation {
	for i := range params {
		e.Params = append(e.Params, params[i])
	}
	return e
}

func (e *expectation) WithOutput(output ...interface{}) *expectation {
	for i := range output {
		copy := Copy(output[i])
		e.Output = append(e.Output, copy)
	}

	return e
}

func (e *expectation) WithError(err error) *expectation {
	e.Error = err
	return e
}

func (it *DB) query(params ...interface{}) {
	fnName := it.getFnName()
	if it.index >= len(it.expectations) {
		msg := fmt.Sprintf("Unexpected call to %s", fnName)
		it.testErrors = append(it.testErrors, errors.New(msg))
		return
	}

	expected := it.expectations[it.index].Function
	if fnName != expected {
		msg := fmt.Sprintf("Unexpected call to %s; expected %s", fnName, expected)
		it.testErrors = append(it.testErrors, errors.New(msg))
	}

	err := it.verifyParams(fnName, params...)
	if err != nil {
		it.testErrors = append(it.testErrors, err)
	}

	err = it.setOutputParams(params...)
	if err != nil {
		it.testErrors = append(it.testErrors, err)
	}

	it.setError()

	it.index++
}

func (it *DB) getFnName() string {
	pc := make([]uintptr, 1)
	runtime.Callers(3, pc)

	frames := runtime.CallersFrames(pc)
	frame, _ := frames.Next()
	pathParts := strings.Split(frame.Function, ".")
	return pathParts[len(pathParts)-1]
}

func (it *DB) verifyParams(fn string, params ...interface{}) error {
	if len(it.expectations[it.index].Params) == 0 {
		return nil
	} else if len(it.expectations[it.index].Params) > len(params) {
		return errors.New("Not enough parameters")
	} else if len(it.expectations[it.index].Params) < len(params) {
		return errors.New("Too many parameters")
	}

	for i := range params {
		expected := it.expectations[it.index].Params[i]
		if !reflect.DeepEqual(params[i], expected) {
			msg := fmt.Sprintf("Unexpected parameter %v in func %s; expected %v", params[i], fn, expected)
			return errors.New(msg)
		}
	}

	return nil
}

func (it *DB) setOutputParams(params ...interface{}) error {
	if len(it.expectations[it.index].Output) == 0 {
		return nil
	}

	for i := range params {
		if i >= len(it.expectations[it.index].Output) {
			break
		}

		t := reflect.TypeOf(params[i])
		if t.Kind() != reflect.Ptr {
			msg := fmt.Sprintf("Out parameters must be pointers. Got kind %v.", t.Kind())
			return errors.New(msg)
		}

		deepCopy(it.expectations[it.index].Output[i], params[i])
	}

	return nil
}

func (it *DB) setError() {
	it.Error = it.expectations[it.index].Error
}

func deepCopy(src interface{}, dst interface{}) error {
	encoded, err := json.Marshal(src)
	if err != nil {
		return err
	}

	err = json.Unmarshal(encoded, dst)
	return err
}

func (it *DB) Close() error {
	return nil
}

func (it *DB) DB() *sql.DB {
	return it.db
}

func (it *DB) New() *DB {
	db, _, _ := sqlmock.New()
	return &DB{db: db}
}

func (it *DB) Select(params ...interface{}) *DB {
	it.query(params...)
	return it
}

func (it *DB) First(params ...interface{}) *DB {
	it.query(params...)
	return it
}

func (it *DB) Find(params ...interface{}) *DB {
	it.query(params...)
	return it
}

func (it *DB) Related(params ...interface{}) *DB {
	it.query(params...)
	return it
}

func (it *DB) Update(attrs ...interface{}) *DB {
	it.query(attrs...)
	return it
}

func (it *DB) Save(value interface{}) *DB {
	it.query(value)
	return it
}

func (it *DB) Create(value interface{}) *DB {
	it.query(value)
	return it
}

func (it *DB) CreateTable(values ...interface{}) *DB {
	it.query(values...)
	return it
}

func (it *DB) HasTable(value interface{}) bool {
	it.query(value)
	return true
}

func (it *DB) Where(params ...interface{}) *DB {
	it.query(params...)
	return it
}

func (it *DB) Model(value interface{}) *DB {
	it.query(value)
	return it
}

func (it *DB) Table(value interface{}) *DB {
	it.query(value)
	return it
}

func (it *DB) Joins(value interface{}) *DB {
	it.query(value)
	return it
}

func (it *DB) Scan(value interface{}) *DB {
	it.query(value)
	return it
}

func (it *DB) Delete(value interface{}) *DB {
	it.query(value)
	return it
}
