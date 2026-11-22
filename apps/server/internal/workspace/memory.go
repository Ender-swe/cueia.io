package workspace

import (
	"context"
	"crypto/sha256"
	"errors"
	"path"
	"time"

	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/sample"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type binding struct {
	Workspace, State string
	Machine          provider.Machine
}

func (s *Service) reserve(ctx context.Context, sub string) (binding, error) {
	var b binding
	err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,1))", sub); err != nil {
			return err
		}
		err := tx.QueryRow(ctx, `SELECT u.workspace_id,m.provisioning_state,coalesce(vm_resource_id,''),coalesce(workspace_disk_resource_id,'')
   FROM users u JOIN workspace_machines m ON m.workspace_id=u.workspace_id WHERE clerk_user_id=$1`, sub).Scan(&b.Workspace, &b.State, &b.Machine.ID, &b.Machine.DiskID)
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if err = tx.QueryRow(ctx, "INSERT INTO workspaces DEFAULT VALUES RETURNING id").Scan(&b.Workspace); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO users(clerk_user_id,workspace_id) VALUES($1,$2)", sub, b.Workspace); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO workspace_machines(workspace_id,provider,region,provisioning_state) VALUES($1,$2,$3,'pending')`, b.Workspace, s.ProviderName, s.Region)
		return err
	})
	return b, err
}

func lock(ctx context.Context, tx pgx.Tx, w string, shared bool) error {
	query := "SELECT pg_advisory_xact_lock(hashtextextended($1,2))"
	if shared {
		query = "SELECT pg_advisory_xact_lock_shared(hashtextextended($1,2))"
	}
	_, err := tx.Exec(ctx, query, w)
	return err
}

func lockProvision(ctx context.Context, conn *pgxpool.Conn, w string) error {
	_, err := conn.Exec(ctx, "SELECT pg_advisory_lock(hashtextextended($1,3))", w)
	return err
}

func unlockProvision(conn *pgxpool.Conn, w string) {
	cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := conn.Exec(cleanup, "SELECT pg_advisory_unlock(hashtextextended($1,3))", w); err != nil {
		conn.Conn().Close(cleanup)
	}
}

func provisioningState(ctx context.Context, conn *pgxpool.Conn, w string) (string, error) {
	var state string
	err := conn.QueryRow(ctx, "SELECT provisioning_state FROM workspace_machines WHERE workspace_id=$1", w).Scan(&state)
	return state, err
}

func markReady(ctx context.Context, tx pgx.Tx, w string, m provider.Machine) error {
	_, err := tx.Exec(ctx, `UPDATE workspace_machines SET vm_resource_id=$2,workspace_disk_resource_id=$3,provisioning_state='ready',updated_at=now() WHERE workspace_id=$1`, w, m.ID, m.DiskID)
	return err
}

func saveSample(ctx context.Context, tx pgx.Tx, w, fileID string, v sample.Invitation) error {
	var thread, msg string
	if err := tx.QueryRow(ctx, "INSERT INTO mail_threads(workspace_id,subject) VALUES($1,$2) RETURNING id", w, v.Subject).Scan(&thread); err != nil {
		return err
	}
	if err := tx.QueryRow(ctx, `INSERT INTO mail_messages(workspace_id,thread_id,sender_address,body_text,sent_at) VALUES($1,$2,$3,$4,$5) RETURNING id`, w, thread, v.SenderAddress, v.BodyText, v.SentAt).Scan(&msg); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "INSERT INTO mail_attachments VALUES($1,$2,$3)", w, msg, fileID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "INSERT INTO calendar_events(workspace_id,title,starts_at,ends_at,time_zone) VALUES($1,$2,$3,$4,$5)", w, v.Title, v.StartsAt, v.EndsAt, v.TimeZone); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, "INSERT INTO workspace_preferences(workspace_id,settings) VALUES($1,$2)", w, v.Preferences)
	return err
}

func (s *Service) readyMachine(ctx context.Context, w string) (provider.Machine, error) {
	var m provider.Machine
	err := s.DB.QueryRow(ctx, "SELECT vm_resource_id,workspace_disk_resource_id FROM workspace_machines WHERE workspace_id=$1 AND provisioning_state='ready'", w).Scan(&m.ID, &m.DiskID)
	return m, err
}

func (s *Service) recordReplacement(ctx context.Context, w string, m, current provider.Machine) bool {
	// Concurrent readers can confirm the same replacement. An older
	// observation must never overwrite a newer persisted binding.
	tag, updateErr := s.DB.Exec(ctx, `UPDATE workspace_machines SET vm_resource_id=$1,updated_at=now()
				WHERE workspace_id=$2 AND workspace_disk_resource_id=$3 AND vm_resource_id IN ($4,$1)
				AND provisioning_state='ready'`, current.ID, w, m.DiskID, m.ID)
	return updateErr == nil && tag.RowsAffected() == 1
}

func recordFileVersion(ctx context.Context, tx pgx.Tx, w, name, mediaType, key string, data []byte) (string, error) {
	digest := sha256.Sum256(data)
	var id string
	err := tx.QueryRow(ctx, `INSERT INTO files(workspace_id,filename,media_type,byte_size,sha256_digest,storage_key)
 VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(workspace_id,storage_key) DO NOTHING RETURNING id`, w, path.Base(name), mediaType, len(data), digest[:], key).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, "SELECT id FROM files WHERE workspace_id=$1 AND storage_key=$2", w, key).Scan(&id)
	}
	return id, err
}

func (s *Service) attachmentID(ctx context.Context, w, filename string) (string, bool, error) {
	var id string
	err := s.DB.QueryRow(ctx, `SELECT f.id FROM files f JOIN mail_attachments a ON a.workspace_id=f.workspace_id AND a.file_id=f.id WHERE f.workspace_id=$1 AND f.filename=$2`, w, filename).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	return id, err == nil, err
}

func (s *Service) versionKey(ctx context.Context, w, id string) (string, error) {
	var key string
	err := s.DB.QueryRow(ctx, "SELECT storage_key FROM files WHERE workspace_id=$1 AND id=$2", w, id).Scan(&key)
	return key, err
}

// One statement provides one MVCC snapshot across all collections.
const snapshotQuery = `SELECT w.id, COALESCE((
 SELECT jsonb_agg(jsonb_build_object('id',t.id,'subject',t.subject,'messages',coalesce((
  SELECT jsonb_agg(jsonb_build_object('id',m.id,'senderAddress',m.sender_address,'bodyText',m.body_text,'sentAt',m.sent_at,
   'attachments',coalesce((SELECT jsonb_agg(jsonb_build_object('fileId',f.id,'filename',f.filename) ORDER BY f.id)
     FROM mail_attachments a JOIN files f ON (a.workspace_id,a.file_id)=(f.workspace_id,f.id)
     WHERE a.workspace_id=w.id AND a.message_id=m.id),'[]'::jsonb)) ORDER BY m.sent_at,m.id)
  FROM mail_messages m WHERE m.workspace_id=w.id AND m.thread_id=t.id),'[]'::jsonb)) ORDER BY t.id)
 FROM mail_threads t WHERE t.workspace_id=w.id),'[]'::jsonb),COALESCE((
 SELECT jsonb_agg(jsonb_build_object('id',e.id,'title',e.title,'startsAt',e.starts_at,'endsAt',e.ends_at,'timeZone',e.time_zone) ORDER BY e.starts_at,e.id)
 FROM calendar_events e WHERE e.workspace_id=w.id),'[]'::jsonb),coalesce(p.settings,'{}'::jsonb)
 FROM workspaces w JOIN workspace_machines m ON m.workspace_id=w.id AND m.provisioning_state='ready'
 LEFT JOIN workspace_preferences p ON p.workspace_id=w.id WHERE w.id=$1`

func (s *Service) loadSnapshot(ctx context.Context, w string) (Snapshot, error) {
	var v Snapshot
	err := s.DB.QueryRow(ctx, snapshotQuery, w).Scan(&v.ID, &v.Threads, &v.Events, &v.Preferences)
	return v, err
}

func (s *Service) updatePreferences(ctx context.Context, w string, changes map[string]string) (map[string]string, error) {
	var result map[string]string
	err := s.DB.QueryRow(ctx, `INSERT INTO workspace_preferences(workspace_id,settings) SELECT id,$2::jsonb FROM workspaces WHERE id=$1
 ON CONFLICT(workspace_id) DO UPDATE SET settings=workspace_preferences.settings || excluded.settings,updated_at=now() RETURNING settings`, w, changes).Scan(&result)
	return result, err
}

func (s *Service) checkFileOwner(ctx context.Context, w, id string) error {
	var exists bool
	err := s.DB.QueryRow(ctx, "SELECT EXISTS(SELECT FROM files WHERE workspace_id=$1 AND id=$2)", w, id).Scan(&exists)
	if err == nil && !exists {
		return pgx.ErrNoRows
	}
	return err
}

type storedAttachment struct {
	Name string
	Key  string
}

func (s *Service) listAttachments(ctx context.Context, w string) ([]storedAttachment, error) {
	rows, err := s.DB.Query(ctx, `SELECT DISTINCT f.filename, f.storage_key FROM files f JOIN mail_attachments a ON (a.workspace_id,a.file_id)=(f.workspace_id,f.id) WHERE f.workspace_id=$1`, w)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[storedAttachment])
}
