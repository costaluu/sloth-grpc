package sql_service

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc/metadata"
)

/* Request & Validation Types */

type ColumnType uint8

const (
	STR     ColumnType = 0
	NON_STR ColumnType = 1
)

const (
	VALIDATION_ABORTED      string = "ABORTED"
	VALIDATION_UNAUTHORIZED        = "UNAUTHORIZED"
	VALIDATION_OK                  = "OK"
)

const (
	REQUEST_OK           string = "OK"
	REQUEST_ABORTED             = "ABORTED"
	REQUEST_UNAUTHORIZED        = "UNAUTHORIZED"
	REQUEST_NOT_FOUND           = "NOT FOUND"
)

const (
	RESULT_INTERNAL          string = "INTERNAL ERROR"
	RESULT_INVALID_FORM             = "INVALID FORM"
	RESULT_OK                       = "OK"
	RESULT_BAD_DB_CONNECTION        = "BAD DATABSE CONNECTION"
	RESULT_QUERY_ERROR              = "QUERY ERROR"
)

const (
	ERROR_NULL string = "NULL"
)

type SQLServiceServer struct {
	UnimplementedSQLServicesServer
}

func requestOutput(status string, result string, e string) (*Output, error) {
	return &Output{
		Status: status,
		Result: result,
		Error:  e,
	}, nil
}

func validation(ctx context.Context, request *Input) (string, string) {

	metaData, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return VALIDATION_ABORTED, RESULT_INTERNAL
	}

	if len(metaData["auth"]) == 0 {
		return VALIDATION_ABORTED, RESULT_INVALID_FORM
	}

	if request.GetQuery() == "" {
		return VALIDATION_ABORTED, RESULT_INVALID_FORM
	}

	if metaData["auth"][0] != os.Getenv("PASS") {
		return VALIDATION_UNAUTHORIZED, RESULT_OK
	}

	return VALIDATION_OK, RESULT_OK
}

func (s *SQLServiceServer) RawQuery(ctx context.Context, request *Input) (*Output, error) {
	requestValidationStatus, requestValidationResult := validation(ctx, request)

	if requestValidationStatus == VALIDATION_ABORTED || requestValidationStatus == VALIDATION_UNAUTHORIZED {
		return requestOutput(requestValidationStatus, requestValidationResult, ERROR_NULL)
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME")))
	defer db.Close()

	if err != nil {
		return requestOutput(REQUEST_ABORTED, RESULT_BAD_DB_CONNECTION, err.Error())
	}

	_, err1 := db.Exec(request.GetQuery())

	if err1 != nil {
		return requestOutput(REQUEST_ABORTED, RESULT_QUERY_ERROR, err1.Error())
	}

	return requestOutput(REQUEST_OK, RESULT_OK, ERROR_NULL)
}

func (s *SQLServiceServer) InsertQuery(ctx context.Context, request *Input) (*Output, error) {
	requestValidationStatus, requestValidationResult := validation(ctx, request)

	if requestValidationStatus == VALIDATION_ABORTED || requestValidationStatus == VALIDATION_UNAUTHORIZED {
		return requestOutput(requestValidationStatus, requestValidationResult, ERROR_NULL)
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME")))
	defer db.Close()

	if err != nil {
		return requestOutput(REQUEST_ABORTED, RESULT_BAD_DB_CONNECTION, err.Error())
	}

	_, err1 := db.Exec(request.GetQuery())

	if err1 != nil {
		return requestOutput(REQUEST_ABORTED, RESULT_QUERY_ERROR, err1.Error())
	}

	return requestOutput(REQUEST_OK, RESULT_OK, ERROR_NULL)
}

func (s *SQLServiceServer) SelectQuery(ctx context.Context, request *Input) (*Output, error) {
	requestValidationStatus, requestValidationResult := validation(ctx, request)

	if requestValidationStatus == VALIDATION_ABORTED || requestValidationStatus == VALIDATION_UNAUTHORIZED {
		return requestOutput(requestValidationStatus, requestValidationResult, ERROR_NULL)
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME")))
	defer db.Close()

	if err != nil {
		return requestOutput(REQUEST_ABORTED, RESULT_BAD_DB_CONNECTION, err.Error())
	}

	rows, err1 := db.Query(request.GetQuery())

	if err1 != nil {
		return requestOutput(REQUEST_ABORTED, RESULT_QUERY_ERROR, err1.Error())
	}

	defer rows.Close()

	columns, err2 := rows.Columns()

	if err2 != nil {
		return requestOutput(REQUEST_ABORTED, RESULT_INTERNAL, err2.Error())
	}

	colTypes, err3 := rows.ColumnTypes()

	if err3 != nil {
		return requestOutput(REQUEST_ABORTED, RESULT_INTERNAL, err3.Error())
	}

	var columnsType []ColumnType

	for _, s := range colTypes {
		if s.DatabaseTypeName() == "VARCHAR" {
			columnsType = append(columnsType, STR)
		} else {
			columnsType = append(columnsType, NON_STR)
		}
	}

	emptyResponse := true
	jsonResponse := "["

	for rows.Next() {
		values := make([]interface{}, len(columns))

		for i := range values {
			values[i] = new(interface{})
		}

		if err2 = rows.Scan(values...); err2 != nil {
			return requestOutput(REQUEST_ABORTED, RESULT_INTERNAL, err2.Error())
		}

		column := "{"

		for i := range values {
			if i+1 == len(values) {
				if columnsType[i] == STR {
					column = column + fmt.Sprintf(`"%s":"%s"}`, columns[i], fmt.Sprintf("%s", *values[i].(*interface{})))
				} else {
					column = column + fmt.Sprintf(`"%s":%s}`, columns[i], fmt.Sprintf("%s", *values[i].(*interface{})))
				}
			} else {
				if columnsType[i] == STR {
					column = column + fmt.Sprintf(`"%s":"%s",`, columns[i], fmt.Sprintf("%s", *values[i].(*interface{})))
				} else {
					column = column + fmt.Sprintf(`"%s":%s,`, columns[i], fmt.Sprintf("%s", *values[i].(*interface{})))
				}
			}
		}

		jsonResponse = jsonResponse + column + ","
		emptyResponse = false
	}

	if emptyResponse == true {
		return requestOutput(REQUEST_NOT_FOUND, RESULT_OK, ERROR_NULL)
	} else {
		jsonResponse = jsonResponse[:len(jsonResponse)-1] + "]"
		return requestOutput(REQUEST_OK, jsonResponse, ERROR_NULL)
	}
}

func (s *SQLServiceServer) UpdateQuery(ctx context.Context, request *Input) (*Output, error) {
	requestValidationStatus, requestValidationResult := validation(ctx, request)

	if requestValidationStatus == VALIDATION_ABORTED || requestValidationStatus == VALIDATION_UNAUTHORIZED {
		return requestOutput(requestValidationStatus, requestValidationResult, ERROR_NULL)
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME")))
	defer db.Close()

	if err != nil {
		return requestOutput(REQUEST_ABORTED, RESULT_BAD_DB_CONNECTION, err.Error())
	}

	_, err1 := db.Exec(request.GetQuery())

	if err1 != nil {
		return requestOutput(REQUEST_ABORTED, RESULT_QUERY_ERROR, err1.Error())
	}

	return requestOutput(REQUEST_OK, RESULT_OK, ERROR_NULL)
}

func (s *SQLServiceServer) DeleteQuery(ctx context.Context, request *Input) (*Output, error) {
	requestValidationStatus, requestValidationResult := validation(ctx, request)

	if requestValidationStatus == VALIDATION_ABORTED || requestValidationStatus == VALIDATION_UNAUTHORIZED {
		return requestOutput(requestValidationStatus, requestValidationResult, ERROR_NULL)
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME")))
	defer db.Close()

	if err != nil {
		return requestOutput(REQUEST_ABORTED, RESULT_BAD_DB_CONNECTION, err.Error())
	}

	_, err1 := db.Exec(request.GetQuery())

	if err1 != nil {
		return requestOutput(REQUEST_ABORTED, RESULT_QUERY_ERROR, err1.Error())
	}

	return requestOutput(REQUEST_OK, RESULT_OK, ERROR_NULL)
}
