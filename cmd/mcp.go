package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/princepal9120/testgen-cli/internal/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run TestGen as an MCP server over stdio",
	Long: `Run TestGen as an MCP-compatible stdio server.

This exposes generate, analyze, and validate tools through the shared app layer.`,
	RunE: runMCP,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

func runMCP(cmd *cobra.Command, args []string) error {
	server := mcp.NewServer(Version)
	if err := server.Run(context.Background(), os.Stdin, os.Stdout); err != nil {
		return fmt.Errorf("mcp server failed: %w", err)
	}
	return nil
}
