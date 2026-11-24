package database_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/Ender-swe/cueia.io/apps/server/internal/config"
	"github.com/Ender-swe/cueia.io/apps/server/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// The workspace, document registration and task rows are original Cueia fixtures.
// Expected relationships and uniqueness rules come from the task recovery design.
func TestTaskRecoveryMigration(t *testing.T) {
	ctx := context.Background()
	dir := os.DirFS(config.MigrationsDir("../../../migrations"))
	for _, install := range []string{"fresh", "upgrade"} {
		t.Run(install, func(t *testing.T) {
			conn := migrationConnection(t)
			exec := func(query string, args ...any) {
				t.Helper()
				if _, err := conn.Exec(ctx, query, args...); err != nil {
					t.Fatal(err)
				}
			}
			id := func(query string, args ...any) string {
				t.Helper()
				var value string
				if err := conn.QueryRow(ctx, query, args...).Scan(&value); err != nil {
					t.Fatal(err)
				}
				return value
			}
			rejects := func(code, query string, args ...any) {
				t.Helper()
				_, err := conn.Exec(ctx, query, args...)
				var pgErr *pgconn.PgError
				if !errors.As(err, &pgErr) || pgErr.Code != code {
					t.Fatalf("expected SQLSTATE %s, got %v for %s", code, err, query)
				}
			}

			const workspace = "INSERT INTO workspaces DEFAULT VALUES RETURNING id"
			var owner, file string
			if install == "upgrade" {
				if err := database.Migrate(ctx, conn, bootstrapMigration(t, dir)); err != nil {
					t.Fatal(err)
				}
				owner = id(workspace)
				file = id(`INSERT INTO files(workspace_id,filename,media_type,byte_size,sha256_digest,storage_key)
                    VALUES ($1,'agenda.pdf','application/pdf',3,sha256('abc'::bytea),'original-agenda') RETURNING id`, owner)
			}
			for range 2 {
				if err := database.Migrate(ctx, conn, dir); err != nil {
					t.Fatal(err)
				}
			}
			if install == "fresh" {
				owner = id(workspace)
			} else if got := id(`SELECT id FROM files WHERE workspace_id=$1 AND storage_key='original-agenda'
                AND byte_size=3 AND sha256_digest=sha256('abc'::bytea)`, owner); got != file {
				t.Fatalf("upgraded file ID = %s, want %s", got, file)
			}
			other := id(workspace)
			const main = "INSERT INTO conversations(workspace_id,kind,title) VALUES ($1,'main','Main')"
			chat := id(main+" RETURNING id", owner)
			rejects("23505", main, owner)
			foreignChat := id(main+" RETURNING id", other)
			exec("INSERT INTO conversations(workspace_id,kind,title) VALUES ($1,'side','Workshop')", owner)

			const task = `INSERT INTO tasks(workspace_id,conversation_id,task_type,request,initial_plan,current_run)
                VALUES ($1,$2,'reminder','{}','{}',1)`
			rejects("23503", task, owner, chat) // Missing current run at commit.
			const run = `INSERT INTO task_runs(workspace_id,task_id,run_number,state,version,handler_version,checkpoint_version,checkpoint,progress)
                VALUES ($1,$2,$3,'queued',1,1,1,'{}','{}')`
			var taskID string
			err := pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
				if err := tx.QueryRow(ctx, task+" RETURNING id", owner, chat).Scan(&taskID); err != nil {
					return err
				}
				_, err := tx.Exec(ctx, run, owner, taskID, 1)
				return err
			})
			if err != nil {
				t.Fatal(err)
			}
			rejects("23503", "UPDATE tasks SET conversation_id=$1 WHERE id=$2", foreignChat, taskID)
			rejects("23503", run, other, taskID, 1)
			const attempt = `INSERT INTO task_attempts(workspace_id,task_id,run_number,lease_token,checkpoint_version,checkpoint,progress)
                VALUES ($1,$2,1,gen_random_uuid(),1,'{}','{}')`
			rejects("23503", attempt, other, taskID)
			exec(attempt, owner, taskID)
			rejects("23505", attempt, owner, taskID)
			exec("UPDATE task_attempts SET finished_at=now() WHERE workspace_id=$1 AND task_id=$2", owner, taskID)
			exec(attempt, owner, taskID)

			const input = `INSERT INTO task_inputs(workspace_id,task_id,run_number,state,prompt,answer_schema)
                VALUES ($1,$2,1,'pending','{}','{}')`
			rejects("23503", input, other, taskID)
			exec(input, owner, taskID)
			rejects("23505", input, owner, taskID)
			const outcome = `INSERT INTO task_outcomes(workspace_id,task_id,run_number,status,message,result)
                VALUES ($1,$2,$3,'failed','Reminder expired','{}')`
			rejects("23503", outcome, other, taskID, 1)
			exec(outcome, owner, taskID, 1)
			rejects("23505", outcome, owner, taskID, 1)
			exec(run, owner, taskID, 2)
			exec(outcome, owner, taskID, 2)
			const replay = `INSERT INTO mutation_responses(workspace_id,operation_scope,request_key,request_fingerprint,http_status,response)
                VALUES ($1,'create','sample',sha256('payload'::bytea),201,'{}')`
			exec(replay, owner)
			rejects("23505", replay, owner)
			exec(replay, other)
		})
	}
}
