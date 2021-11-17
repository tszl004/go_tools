package core_errors

import (
	"errors"
	"github.com/tszl004/go_tools/core_const"
)

type RespErr struct {
	error
	Code int
}

var (
	ErrAuthInvalid = RespErr{errors.New("ErrAuthInvalid"), core_const.CodeTokenInvalid}

	ErrBasePath = errors.New("ErrBasePath")

	ErrCasbinNoAuthorization     = errors.New("ErrCasbinNoAuthorization")
	ErrContainerKeyAlreadyExists = errors.New("ErrContainerKeyAlreadyExists")
	ErrCrudSaveFail              = RespErr{errors.New("save failed"), core_const.CodeCurdUpdateFail}
	ErrCrudCreateFail            = RespErr{errors.New("add failed"), core_const.CodeCurdCreatFail}
	ErrCrudDelFail               = RespErr{errors.New("del failed"), core_const.CodeCurdDeleteFail}
	ErrCrudSelectFail            = RespErr{errors.New("data doesn't exists"), core_const.CodeCurdDeleteFail}

	ErrDbDriverNotExists = errors.New("ErrDbDriverNotExists")
	ErrDbDialFail        = errors.New("ErrDbDialFail")

	ErrFuncEventNotCall       = errors.New("ErrFuncEventNotCall")
	ErrFuncEventNotRegister   = errors.New("ErrFuncEventNotRegister")
	ErrFuncEventAlreadyExists = errors.New("ErrFuncEventAlreadyExists")

	ErrGormInitFail = errors.New("ErrGormInitFail")
	ErrGormNotInit  = errors.New("ErrGormNotInit")

	ErrFilesUploadOpenFail = errors.New("ErrFilesUploadOpenFail")
	ErrFilesUploadReadFail = errors.New("ErrFilesUploadReadFail")

	ErrSmsSenderInvalid = errors.New("ErrSmsSenderInvalid")

	ErrTokenParseFail    = RespErr{errors.New("ErrTokenParseFail"), core_const.CodeTokenInvalid}
	ErrTokenInvalid      = RespErr{errors.New("ErrTokenInvalid"), core_const.CodeTokenInvalid}
	ErrTokenMalFormed    = RespErr{errors.New("ErrTokenMalFormed"), core_const.CodeTokenInvalid}
	ErrTokenNotActiveYet = RespErr{errors.New("ErrTokenNotActiveYet"), core_const.CodeTokenInvalid}

	ErrValidatorBindParamsFail = errors.New("ErrValidatorBindParamsFail")
	ErrValidatorInvalid        = RespErr{errors.New(core_const.MsgValidatorInvalid), core_const.CodeValidatorFail}
)
