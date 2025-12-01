package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func CreateDefaultHanlder(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Create",
		"data": fiber.Map{
			"uuid": uuid.New().String(),
		},
		"code": 200,
	})
}

func ReadListDefaultHanlder(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Read List",
		"data": fiber.Map{
			"uuid": uuid.New().String(),
		},
		"code": 200,
	})
}

func ReadSingleDefaultHanlder(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Read Single",
		"data": fiber.Map{
			"uuid": uuid.New().String(),
		},
		"code": 200,
	})
}

func UpdateDefaultHanlder(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Update",
		"data": fiber.Map{
			"uuid": uuid.New().String(),
		},
		"code": 200,
	})
}

func DeleteDefaultHanlder(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Delete",
		"data": fiber.Map{
			"uuid": uuid.New().String(),
		},
		"code": 200,
	})
}
