package modules

import (
	"context"
	"database/sql"
	"encoding/json"

	// "database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/cunkz/goyummy/bin/config"
	"github.com/cunkz/goyummy/bin/helpers/auth"
	"github.com/cunkz/goyummy/bin/helpers/db"
	"github.com/cunkz/goyummy/bin/helpers/utils"
)

func RegisterModules(app *fiber.App, cfg *config.AppConfig) {
	// Initalize Auth
	authMap, _ := auth.BuildAuthMap(cfg)

	for _, m := range cfg.Modules {
		mSlug := utils.ToSlug(m.Name)
		baseRoute := fmt.Sprintf("/api/%s/v1", mSlug)
		authMiddleware := authMap[m.Auth]
		dbEngine := config.GetDBEngineByName(cfg, m.Database)
		if dbEngine == "mongo" {
			registerModuleMongo(app, authMiddleware, baseRoute, m)
		} else {
			registerModule(app, authMiddleware, baseRoute, m)
		}
	}
}

func registerModule(app *fiber.App, authMiddleware fiber.Handler, baseRoute string, m config.Module) {
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

	for _, op := range m.Operations {
		switch strings.ToLower(op) {
		case "create":
			// ----------------------------
			// CREATE (INSERT)
			// ----------------------------
			createHandler := func(c *fiber.Ctx) error {
				body := map[string]string{}
				if err := c.BodyParser(&body); err != nil {
					return err
				}

				args := []any{}
				for _, f := range m.Fields {
					switch f {
					case "id":
						args = append(args, uuid.New().String())
					case "created_at", "updated_at":
						args = append(args, time.Now())
					default:
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
			}
			if authMiddleware != nil {
				app.Post(baseRoute, authMiddleware, createHandler)
			} else {
				app.Post(baseRoute, createHandler)
			}
			log.Info().Msgf("Add Route POST %s", baseRoute)
		case "read_list":
			// ----------------------------
			// GET ALL
			// ----------------------------
			getHandler := func(c *fiber.Ctx) error {
				// Make SELECT list: "id, name, description ..."
				cols := append([]string{"id"}, m.Fields...)
				query := "SELECT " + strings.Join(cols, ",") + " FROM " + m.Table

				rows, err := db.Query(query)
				if err != nil {
					return err
				}
				defer rows.Close()

				list := []map[string]any{}

				for rows.Next() {
					// prepare scan targets
					scanTargets := make([]any, len(cols))

					// id (string)
					var id sql.NullString
					scanTargets[0] = &id

					// other fields (string)
					nullFields := make([]sql.NullString, len(m.Fields))
					for i := range nullFields {
						scanTargets[i+1] = &nullFields[i]
					}

					// execute scan
					if err := rows.Scan(scanTargets...); err != nil {
						return err
					}

					// build output
					item := map[string]any{
						"id": id.String,
					}

					for i, f := range m.Fields {
						if nullFields[i].Valid {
							item[f] = nullFields[i].String
						} else {
							item[f] = nil
						}
					}

					list = append(list, item)
				}

				// return c.JSON(list)

				return utils.ResponseSuccess(c, list, "Successfully read data")
			}
			if authMiddleware != nil {
				app.Get(baseRoute, authMiddleware, getHandler)
			} else {
				app.Get(baseRoute, getHandler)
			}
			log.Info().Msgf("Add Route GET %s", baseRoute)
		case "read_single":
			// ----------------------------
			// GET by ID
			// ----------------------------
			getHandler := func(c *fiber.Ctx) error {
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
			}
			if authMiddleware != nil {
				app.Get(baseRoute+"/:id", authMiddleware, getHandler)
			} else {
				app.Get(baseRoute+"/:id", getHandler)
			}
			log.Info().Msgf("Add Route GET %s", baseRoute+"/:id")
		case "update":
			// ----------------------------
			// UPDATE
			// ----------------------------
			updateHandler := func(c *fiber.Ctx) error {
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
			}
			if authMiddleware != nil {
				app.Patch(baseRoute+"/:id", authMiddleware, updateHandler)
			} else {
				app.Patch(baseRoute+"/:id", updateHandler)
			}
			log.Info().Msgf("Add Route PATCH %s", baseRoute+"/:id")
		case "delete":
			// ----------------------------
			// DELETE
			// ----------------------------
			deleteHandler := func(c *fiber.Ctx) error {
				id := c.Params("id")

				_, err := db.Exec("DELETE FROM "+m.Table+" WHERE id=$1", id)
				if err != nil {
					return err
				}

				return utils.ResponseSuccess(c, fiber.Map{"deleted": true}, "Successfully delete data")
			}
			if authMiddleware != nil {
				app.Delete(baseRoute+"/:id", authMiddleware, deleteHandler)
			} else {
				app.Delete(baseRoute+"/:id", deleteHandler)
			}
			log.Info().Msgf("Add Route DELETE %s", baseRoute+"/:id")
		default:
			log.Info().Msgf("Invalid Operation for Module: %s", m.Name)
		}
	}
}

func registerModuleMongo(app *fiber.App, authMiddleware fiber.Handler, baseRoute string, m config.Module) {
	mongoDB := db.MongoDBs[m.Database]
	col := mongoDB.Collection(m.Table)

	for _, op := range m.Operations {
		switch strings.ToLower(op) {
		case "create":
			// ----------------------------
			// CREATE (INSERT)
			// ----------------------------
			app.Post(baseRoute, func(c *fiber.Ctx) error {
				ctx := context.Background()

				body := map[string]any{}
				if err := c.BodyParser(&body); err != nil {
					return utils.ResponseError(c, 400, err.Error())
				}

				id := uuid.New().String()
				body["id"] = id
				body["created_at"] = time.Now()
				body["updated_at"] = time.Now()
				_, err := col.InsertOne(ctx, body)
				if err != nil {
					return utils.ResponseError(c, 500, err.Error())
				}
				return utils.ResponseSuccess(c, fiber.Map{"id": id}, "Data has been created")
			})
			log.Info().Msgf("Add Route POST %s", baseRoute)
		case "read_list":
			// ----------------------------
			// GET ALL
			// ----------------------------
			app.Get(baseRoute, func(c *fiber.Ctx) error {
				ctx := context.Background()

				cursor, err := col.Find(ctx, bson.M{})
				if err != nil {
					return utils.ResponseError(c, 500, err.Error())
				}
				defer cursor.Close(ctx)

				results := []bson.M{}
				if err := cursor.All(ctx, &results); err != nil {
					return utils.ResponseError(c, 500, err.Error())
				}

				return utils.ResponseSuccess(c, results, "Data has been created")
			})
			log.Info().Msgf("Add Route GET %s", baseRoute)
		case "read_single":
			// ----------------------------
			// GET by ID
			// ----------------------------
			app.Get(baseRoute+"/:id", func(c *fiber.Ctx) error {
				ctx := context.Background()
				id := c.Params("id")

				result := bson.M{}
				err := col.FindOne(ctx, bson.M{"id": id}).Decode(&result)
				if err == mongo.ErrNoDocuments {
					return utils.ResponseError(c, 404, "Data not found")
				}
				if err != nil {
					return utils.ResponseError(c, 500, err.Error())
				}

				return utils.ResponseSuccess(c, result, "Successfully read data")
			})
			log.Info().Msgf("Add Route GET %s", baseRoute+"/:id")
		case "update":
			// ----------------------------
			// UPDATE
			// ----------------------------
			app.Patch(baseRoute+"/:id", func(c *fiber.Ctx) error {
				ctx := context.Background()

				// Parse ID from URL
				id := c.Params("id")

				// Parse request body into dynamic map
				body := map[string]any{}
				if err := c.BodyParser(&body); err != nil {
					return utils.ResponseError(c, 400, "Invalid JSON Body")
				}

				// Prevent updating primary key fields
				delete(body, "_id")
				delete(body, "id")

				// If body is empty, do nothing
				if len(body) == 0 {
					return utils.ResponseError(c, 400, "No fields to update")
				}

				// Add updated_at automatically (optional)
				body["updated_at"] = time.Now()

				b, _ := json.Marshal(body)
				jsonString := string(b)

				log.Info().Msg(id)
				log.Info().Msg(jsonString)

				// Do partial update with $set
				filter := bson.M{"id": id}
				_, err := col.UpdateOne(ctx, filter, bson.M{"$set": body})
				if err != nil {
					return utils.ResponseError(c, 500, err.Error())
				}
				return utils.ResponseSuccess(c, fiber.Map{"updated": true}, "Successfully update data")
			})
			log.Info().Msgf("Add Route PATCH %s", baseRoute+"/:id")
		case "delete":
			// ----------------------------
			// DELETE
			// ----------------------------
			app.Delete(baseRoute+"/:id", func(c *fiber.Ctx) error {
				ctx := context.Background()
				id := c.Params("id")

				_, err := col.DeleteOne(ctx, bson.M{"id": id})
				if err != nil {
					return utils.ResponseError(c, 500, err.Error())
				}

				return utils.ResponseSuccess(c, fiber.Map{"deleted": true}, "Successfully delete data")
			})
			log.Info().Msgf("Add Route DELETE %s", baseRoute+"/:id")
		default:
			log.Info().Msgf("Invalid Operation for Module: %s", m.Name)
		}
	}
}
