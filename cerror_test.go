package cerror

import (
	"reflect"
	"testing"

	"google.golang.org/grpc/codes"
)

var (
	errLevelX = Error{
		Type:        codes.InvalidArgument,
		Code:        "OTHER_WEIRD_ERROR",
		Description: "Test weird error",
		Cause:       "Test Cause",
		ComesFrom:   nil,
	}

	errLevel0 = Error{
		Type:        codes.InvalidArgument,
		Code:        "EIA",
		Description: "The argument 'password' is invalid",
		Cause:       "'{1}' lower than {2}",
		ComesFrom:   nil,
		Meta: map[string]string{
			"arg": "'password",
			"req": "len > 6",
		},
	}

	errLevel1 = Error{
		Type:        codes.InvalidArgument,
		Code:        "ECPw",
		Description: "The argument 'password' is invalid",
		Cause:       "'password' is invalid",
		ComesFrom:   &errLevel0,
		Meta:        nil,
	}

	errLevel2 = Error{
		Type:        codes.PermissionDenied,
		Code:        "LOGIN_ERR",
		Description: "The login process is invalid",
		Cause:       "",
		ComesFrom:   &errLevel1,
		Meta: map[string]string{
			"sample_meta": "test",
		},
	}
)

func TestError_GetParents(t *testing.T) {
	tests := []struct {
		name string
		err  Error
		want []*Error
	}{
		{
			name: "with parent",
			err:  errLevel2,
			want: []*Error{&errLevel1, &errLevel0},
		},
		{
			name: "with no parent",
			err:  errLevel0,
			want: []*Error{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := tt.err.GetParents(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Error.GetParents() = %v, wantComesFrom %v", got, tt.want)
			}
		})
	}
}

func TestError_IsError(t *testing.T) {
	tests := []struct {
		name string
		err  Error
		want bool
	}{
		{
			name: "is error",
			err:  errLevel0,
			want: true,
		},
		{
			name: "not error",
			err:  Error{},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.IsError(); got != tt.want {
				t.Errorf("Error.IsError() = %v, wantComesFrom %v", got, tt.want)
			}
		})
	}
}

func TestError_Equals(t *testing.T) {
	tests := []struct {
		name string
		err1 Error
		err2 Error
		want bool
	}{
		{
			name: "is error",
			err1: errLevel0,
			err2: errLevel0,
			want: true,
		},
		{
			name: "is error",
			err1: errLevel1,
			err2: errLevel0,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err1.Equals(tt.err2); got != tt.want {
				t.Errorf("Error.IsError() = %v, wantComesFrom %v", got, tt.want)
			}
		})
	}
}

func TestError_SetParam(t *testing.T) {
	tests := []struct {
		name   string
		err    Error
		params map[string]int
		want   string
	}{
		{
			name:   "with params",
			err:    errLevel0,
			params: map[string]int{"password": 1, "6": 2},
			want:   "'password' lower than 6",
		},
		{
			name:   "wrong param",
			err:    errLevel0,
			params: map[string]int{"password": 7, "6": 2},
			want:   "'{1}' lower than 6",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for val, pos := range tt.params {
				tt.err = tt.err.SetParam(pos, val)
			}
			if tt.err.Cause != tt.want {
				t.Errorf("Error.SetParam() = %v, wantComesFrom %v", tt.err.Cause, tt.want)
			}
		})
	}
}

func TestError_From(t *testing.T) {
	tests := []struct {
		name          string
		err1          Error
		err2          Error
		wantComesFrom Error
	}{
		{
			name:          "is error",
			err1:          errLevel0,
			err2:          errLevelX,
			wantComesFrom: errLevelX,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err1.From(tt.err2); !reflect.DeepEqual(*got.ComesFrom, tt.wantComesFrom) {
				t.Errorf("Error.From() = %v, wantComesFrom %v", got, tt.wantComesFrom)
			}
			if tt.err1.ComesFrom != nil {
				t.Errorf("Error.From() = %s, original comesFrom is not nil, %v", tt.name, tt.wantComesFrom)
			}
		})
	}
}
