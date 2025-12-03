package modules

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/cunkz/goyummy/bin/config"
	"github.com/cunkz/goyummy/bin/helpers/db"
	"github.com/cunkz/goyummy/bin/helpers/utils"
)

func RegisterModules(app *fiber.App, modules []config.Module) {
	for _, m := range modules {
		registerModule(app, m)
	}
}

func registerModule(app *fiber.App, m config.Module) {
	db := db.PostgresDBs[m.Database] // for now: postgres engine

	m.Fields = append(m.Fields, "id")
	m.Fields = append(m.Fields, "created_at")
	m.Fields = append(m.Fields, "updated_at")
	fields := strings.Join(m.Fields, ",")
	placeholders := make([]string, len(m.Fields))
	for i := range m.Fields {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	placeholdersStr := strings.Join(placeholders, ",")

	mSlug := utils.ToSlug(m.Name)
	baseRoute := fmt.Sprintf("/api/%s/v1", mSlug)
	for _, op := range m.Operations {
		switch strings.ToLower(op) {
		case "create":
			// ----------------------------
			// CREATE (INSERT)
			// ----------------------------
			app.Post(baseRoute, func(c *fiber.Ctx) error {
				body := map[string]string{}
				if err := c.BodyParser(&body); err != nil {
					return err
				}

				args := []any{}
				for _, f := range m.Fields {
					if f == "id" {
						args = append(args, uuid.New().String())
					} else if f == "created_at" || f == "updated_at" {
						args = append(args, time.Now())
					} else {
						args = append(args, body[f])
					}
				}

				query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id",
					m.Table, fields, placeholdersStr)

				var id string
				err := db.QueryRow(query, args...).Scan(&id)
				if err != nil {
					return err
				}

				return utils.ResponseSuccess(c, fiber.Map{"id": id}, "Data has been created")
			})
			log.Info().Msgf("Add Route POST %s", baseRoute)
		case "read_list":
			// ----------------------------
			// GET ALL
			// ----------------------------
			app.Get(baseRoute, func(c *fiber.Ctx) error {
				rows, err := db.Query("SELECT id, " + fields + " FROM " + m.Table)
				if err != nil {
					return err
				}
				defer rows.Close()

				list := []map[string]any{}
				for rows.Next() {
					values := make([]any, 0, len(m.Fields)+1)
					valuesPtrs := make([]any, 0, len(m.Fields)+1)

					var id int
					values = append(values, &id)
					valuesPtrs = append(valuesPtrs, &id)

					for range m.Fields {
						var v string
						values = append(values, &v)
						valuesPtrs = append(valuesPtrs, &v)
					}

					rows.Scan(valuesPtrs...)
					item := map[string]any{"id": id}

					for i, f := range m.Fields {
						item[f] = *(values[i+1].(*string))
					}

					list = append(list, item)
				}

				return utils.ResponseSuccess(c, list, "Successfully read data")
			})
			log.Info().Msgf("Add Route GET %s", baseRoute)
		case "read_single":
			// ----------------------------
			// GET by ID
			// ----------------------------
			app.Get(baseRoute+"/:id", func(c *fiber.Ctx) error {
				id := c.Params("id")

				query := "SELECT " + fields + " FROM " + m.Table + " WHERE id=$1"

				row := db.QueryRow(query, id)

				result := map[string]any{"id": id}
				values := make([]any, len(m.Fields))
				for i := range m.Fields {
					var v string
					values[i] = &v
				}

				err := row.Scan(values...)
				if err != nil {
					return utils.ResponseError(c, 404, "Data not found")
				}

				for i, f := range m.Fields {
					result[f] = *(values[i].(*string))
				}

				return utils.ResponseSuccess(c, result, "Successfully read data")
			})
			log.Info().Msgf("Add Route GET %s", baseRoute+"/:id")
		case "update":
			// ----------------------------
			// UPDATE
			// ----------------------------
			app.Patch(baseRoute+"/:id", func(c *fiber.Ctx) error {
				id := c.Params("id")
				body := map[string]string{}
				_ = c.BodyParser(&body)

				sets := []string{}
				args := []any{}
				argNum := 1

				for _, f := range m.Fields {
					if v, ok := body[f]; ok {
						sets = append(sets, fmt.Sprintf("%s=$%d", f, argNum))
						args = append(args, v)
						argNum++
					}
				}

				// Add function refresh updated_at
				sets = append(sets, fmt.Sprintf("%s=$%d", "updated_at", argNum))
				args = append(args, time.Now())
				argNum++

				if len(sets) == 0 {
					return utils.ResponseError(c, 400, "No fields provided")
				}

				args = append(args, id)

				query := fmt.Sprintf("UPDATE %s SET %s WHERE id=$%d",
					m.Table, strings.Join(sets, ", "), argNum)

				_, err := db.Exec(query, args...)
				if err != nil {
					return err
				}

				return utils.ResponseSuccess(c, fiber.Map{"updated": true}, "Successfully update data")
			})
			log.Info().Msgf("Add Route PATCH %s", baseRoute+"/:id")
		case "delete":
			// ----------------------------
			// DELETE
			// ----------------------------
			app.Delete(baseRoute+"/:id", func(c *fiber.Ctx) error {
				id := c.Params("id")

				_, err := db.Exec("DELETE FROM "+m.Table+" WHERE id=$1", id)
				if err != nil {
					return err
				}

				return utils.ResponseSuccess(c, fiber.Map{"deleted": true}, "Successfully delete data")
			})
			log.Info().Msgf("Add Route DELETE %s", baseRoute+"/:id")
		default:
			log.Info().Msgf("Invalid Operation for Module: %s", m.Name)
		}
	}
}
