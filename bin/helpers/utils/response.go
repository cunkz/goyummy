package utils

import "github.com/gofiber/fiber/v2"

type JSONResponse struct {
	Status  bool        `json:"status"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func ResponseSuccess(c *fiber.Ctx, data interface{}, message string) error {
	resp := JSONResponse{
		Status:  true,
		Code:    fiber.StatusOK,
		Data:    data,
		Message: message,
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

func ResponseError(c *fiber.Ctx, code int, message string) error {
	resp := JSONResponse{
		Status:  false,
		Code:    code,
		Data:    nil,
		Message: message,
	}
	return c.Status(code).JSON(resp)
}
