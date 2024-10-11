package cmd

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

func NewServeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "serve",
		RunE: func(cmd *cobra.Command, _ []string) error {
			convertQueryArgs := func(c *gin.Context, s any) []string {
				var (
					queryArgs = make(map[string]string)
					args      []string
				)
				t := reflect.TypeOf(s)
				for i := range t.NumField() {
					field := t.Field(i)
					queryArgs[field.Name] = c.Query(field.Name)
				}
				for k, v := range queryArgs {
					if v == "" {
						continue
					}
					args = append(args, "--"+k, v)
				}
				return args
			}

			flagMap := map[string]any{
				"add":     filterAddFlags{},
				"remove":  filterRemoveFlags{},
				"list":    filterListFlags{},
				"disable": filterDisableFlags{},
				"enable":  filterEnableFlags{},
			}

			runCommandAndSendOutput := func(ctx context.Context, cmd *cobra.Command, c *gin.Context, args []string) {
				o := new(bytes.Buffer)
				cmd.SetOut(o)
				cmd.SetArgs(args)

				if err := cmd.ExecuteContext(ctx); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"status": "error",
						"error":  err.Error(),
					})
					return
				}
				c.JSON(http.StatusOK, gin.H{"status": "success", "data": o.String()})
			}

			router := gin.Default()
			router.GET("/filter/:action", func(c *gin.Context) {
				action := c.Param("action")
				args := []string{action}
				args = append(args, convertQueryArgs(c, flagMap[action])...)
				runCommandAndSendOutput(cmd.Context(), NewFilterCommand(), c, args)
			})

			router.GET("/get", func(c *gin.Context) {
				runCommandAndSendOutput(cmd.Context(), NewGetCommand(), c, convertQueryArgs(c, getFlags{}))
			})

			router.GET("/enable", func(c *gin.Context) {
				runCommandAndSendOutput(cmd.Context(), NewEnableCommand(), c, convertQueryArgs(c, enableFlags{}))
			})

			router.GET("/disable", func(c *gin.Context) {
				runCommandAndSendOutput(cmd.Context(), NewEnableCommand(), c, convertQueryArgs(c, disableFlags{}))
			})

			port := "8080"
			fmt.Fprintf(cmd.OutOrStdout(), "Starting server on port %s...\n", port)
			if err := router.Run(":" + port); err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Failed to start server: %v\n", err)
			}

			return nil
		},
	}
	cmd.Hidden = true
	return cmd
}
