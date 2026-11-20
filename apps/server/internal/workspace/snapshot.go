package workspace

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
)

type Snapshot struct {
	ID          string            `json:"id"`
	Threads     json.RawMessage   `json:"threads"`
	Events      json.RawMessage   `json:"events"`
	Preferences map[string]string `json:"preferences"`
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

func (s *Service) Snapshot(ctx context.Context, w string) (Snapshot, error) {
	var v Snapshot
	err := s.DB.QueryRow(ctx, snapshotQuery, w).Scan(&v.ID, &v.Threads, &v.Events, &v.Preferences)
	return v, err
}
func (s *Service) Preferences(ctx context.Context, w string, changes map[string]string) (map[string]string, error) {
	var result map[string]string
	err := s.DB.QueryRow(ctx, `INSERT INTO workspace_preferences(workspace_id,settings) SELECT id,$2::jsonb FROM workspaces WHERE id=$1
 ON CONFLICT(workspace_id) DO UPDATE SET settings=workspace_preferences.settings || excluded.settings,updated_at=now() RETURNING settings`, w, changes).Scan(&result)
	return result, err
}
func (s *Service) Owns(ctx context.Context, w, id string) error {
	var exists bool
	err := s.DB.QueryRow(ctx, "SELECT EXISTS(SELECT FROM files WHERE workspace_id=$1 AND id=$2)", w, id).Scan(&exists)
	if err == nil && !exists {
		return pgx.ErrNoRows
	}
	return err
}

func (s *Service) Threads(ctx context.Context, w string) (json.RawMessage, error) {
	snapshot, err := s.Snapshot(ctx, w)
	return snapshot.Threads, err
}
func (s *Service) Events(ctx context.Context, w string) (json.RawMessage, error) {
	snapshot, err := s.Snapshot(ctx, w)
	return snapshot.Events, err
}
