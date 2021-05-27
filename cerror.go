package cerror

// Hector Oliveros - 2019
// hector.oliveros.leon@gmail.com

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"golang.org/x/xerrors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"strings"
)

const PrimitiveErrorTag = "PRIMITIVE_ERROR"
const maxDep = 1000 // avoid infinite loops

var Nil = Error{Code: ""}

// Custom error class
type Error struct {
	Description string            `json:"description"`
	Type        codes.Code        `json:"type"`
	Code        string            `json:"code"`
	Cause       string            `json:"cause"`
	IsSensible  bool              `json:"isSensible"`
	Severity    Severity          `json:"severity"`
	Meta        map[string]string `json:"meta"`
	ComesFrom   *Error            `json:"comesFrom"`
	Frame       xerrors.Frame     `json:"frame"`
}

/***** golang.org/x/xerrors compatible functions *******/

// copy from: As method of xerrors
// As finds the first error in err's chain that matches the type to which target
// points, and if so, sets the target to its value and returns true. An error
// matches a type if it is assignable to the target type, or if it has a method
// As(interface{}) bool such that As(target) returns true. As will panic if target
// is not a non-nil pointer to a type which implements error or is of interface type.
//
// The As method should set the target to its value and return true if err
// matches the type to which target points.
func (err Error) As(target *Error) bool {
	if target == nil {
		panic("errors: target cannot be nil")
	}
	current := &err
	for {
		if current == nil {
			return false
		}
		if current.Code == target.Code {
			*target = *current
			return true
		}
		current = current.ComesFrom
	}
	return false
}

// Is reports whether any error in err's chain matches target.
//
// An error is considered to match a target if it is equal to that target or if
// it implements a method Is(error) bool such that Is(target) returns true.
func (err Error) Is(target Error) bool {
	current := &err
	for {
		if current == nil {
			return false
		}
		if current.Code == target.Code {
			return true
		}
		current = current.ComesFrom
	}
}

// FormatError prints the receiver's first error and returns the next error in
// the error chain, if any.
func (err Error) FormatError(p xerrors.Printer) error {
	code := ""
	if len(err.Code) > 0 {
		code = err.Code
	}

	desc := ""
	if len(err.Description) > 0 {
		desc = err.Description
	}
	p.Printf("%s: %s", code, desc)
	err.Frame.Format(p)
	return err.ComesFrom
}

func (err Error) Format(f fmt.State, c rune) {
	xerrors.FormatError(err, f, c)
}

func (err Error) Error() string {
	st := status.New(err.Type, err.Description)
	parentErrors := err.GetParents()

	errDetList := make([]proto.Message, 1)
	errCodeList := make([]string, 1)
	if len(parentErrors) > 0 {
		for _, err := range parentErrors {
			if err == nil {
				continue
			}
			errCodeList = append(errCodeList, err.Code)
		}
	}

	debugInfo := &errdetails.DebugInfo{
		StackEntries: errCodeList,
		Detail:       "",
	}
	errDetList = append(errDetList, debugInfo)

	switch err.Type {
	case codes.Unknown:
		return status.New(codes.NotFound, err.Description).Err().Error()

	case codes.InvalidArgument:
		for fieldName, desc := range err.Meta {
			v := &errdetails.BadRequest_FieldViolation{
				Field:       fieldName,
				Description: desc,
			}
			br := &errdetails.BadRequest{}
			br.FieldViolations = append(br.FieldViolations, v)
		}
		sb, _ := st.WithDetails(errDetList...)
		return sb.Err().Error()
	}
	return status.New(err.Type, err.Description+" | "+err.Code).Err().Error()
}

func (err Error) GetParents() []*Error {
	list := make([]*Error, 0)
	cont := 0
	currentErr := &err
	for {
		cont++
		if currentErr.ComesFrom == nil || cont > maxDep {
			break
		}
		list = append(list, currentErr.ComesFrom)
		currentErr = currentErr.ComesFrom
	}
	return list
}

func (err Error) IsError() bool {
	return err.Code != Nil.Code || err.ComesFrom != nil
}

func (err Error) Equals(err2 Error) bool {
	return err.Code == err2.Code
}

func (err Error) SetParam(num int, val string) Error {
	errR := err
	param := "{" + strconv.Itoa(num) + "}"
	errR.Cause = strings.Replace(errR.Cause, param, val, 1)
	return errR
}

func From(err2 error) Error {
	err := Error{}
	return err.From(err2)
}

func (err Error) From(err2 error) Error {
	errNew := err
	var ptr Error
	switch err2.(type) {
	case Error:
		ptr = err2.(Error)
	default:
		ptr = Error{
			Code:        PrimitiveErrorTag,
			Description: err2.Error(),
		}
	}
	errNew.ComesFrom = &ptr
	return errNew
}

func (err Error) SetCause(cause string) Error {
	errNew := err
	errNew.Cause = cause
	return errNew
}

func (err Error) AddMeta(key string, val string) Error {
	errNew := err
	if errNew.Meta == nil {
		errNew.Meta = make(map[string]string)
	}
	errNew.Meta[key] = val
	return errNew
}
