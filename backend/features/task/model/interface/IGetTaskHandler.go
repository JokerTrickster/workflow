package _interface

import "github.com/labstack/echo/v4"

type IGetTaskHandler interface {
	GetTask(c echo.Context) error
}