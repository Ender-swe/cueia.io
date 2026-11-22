package claim

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/Ender-swe/cueia.io/apps/server/internal/config"
	"github.com/jackc/pgx/v5"
)

// The child reads the candidate before signalling readiness. Its parent
// releases both children only after they have reported the expected version.
func TestClaimChild(t *testing.T) {
	id := testClaimID()
	if id == "" {
		t.Skip("child of TestCompetingClaimProcesses")
	}
	ctx := t.Context()
	conn, err := pgx.Connect(ctx, config.TestDatabaseURL())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	var version int64
	if err := conn.QueryRow(ctx, "SELECT version FROM claim_records WHERE id=$1 AND status='eligible'", id).Scan(&version); err != nil {
		t.Fatal(err)
	}
	fmt.Println("ready", version)
	var release [1]byte
	if _, err := io.ReadFull(os.Stdin, release[:]); err != nil {
		t.Fatal(err)
	}
	won, err := AtVersion(ctx, conn, id, version)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("claimed", won)
}

func TestCompetingClaimProcesses(t *testing.T) {
	dsn := config.TestDatabaseURL()
	if dsn == "" {
		t.Skip("requires migrated CUEIA_TEST_DATABASE_URL")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(context.Background())
	var id, workspace string
	err = conn.QueryRow(ctx, `WITH w AS (INSERT INTO workspaces DEFAULT VALUES RETURNING id)
		INSERT INTO claim_records(workspace_id,status,version) SELECT id,'eligible',4 FROM w RETURNING id,workspace_id`).Scan(&id, &workspace)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Exec(context.Background(), "DELETE FROM workspaces WHERE id=$1", workspace)
	defer conn.Exec(context.Background(), "DELETE FROM claim_records WHERE id=$1", id)
	type child struct {
		cmd    *exec.Cmd
		input  io.WriteCloser
		output *bufio.Scanner
	}
	var children []child
	for range 2 {
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestClaimChild$")
		cmd.Env = append(os.Environ(), "CUEIA_TEST_CLAIM_ID="+id)
		input, err := cmd.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		output, err := cmd.StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		children = append(children, child{cmd, input, bufio.NewScanner(output)})
	}
	for _, c := range children {
		if !c.output.Scan() || c.output.Text() != "ready 4" {
			t.Fatal("child did not observe eligible version 4")
		}
	}
	for _, c := range children {
		if _, err := c.input.Write([]byte{1}); err != nil {
			t.Fatal(err)
		}
		c.input.Close()
	}
	wins := 0
	for _, c := range children {
		if !c.output.Scan() {
			t.Fatal("missing claim result")
		}
		switch c.output.Text() {
		case "claimed true":
			wins++
		case "claimed false":
		default:
			t.Fatal("unexpected claim result")
		}
		if err := c.cmd.Wait(); err != nil {
			t.Fatal(err)
		}
	}
	var version int
	var status string
	err = conn.QueryRow(ctx, "SELECT version,status FROM claim_records WHERE id=$1", id).Scan(&version, &status)
	if err != nil || wins != 1 || version != 5 || status != "claimed" {
		t.Fatal("claim result", wins, version, status, err)
	}
}
