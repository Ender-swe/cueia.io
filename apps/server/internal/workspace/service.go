package workspace

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"time"

	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/sample"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUnavailable = errors.New("workspace unavailable")

type Service struct {
	DB                                     *pgxpool.Pool
	Provider                               provider.Provider
	Deployment, ProviderName, Region, Mode string
}
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

func (s *Service) Provision(ctx context.Context, sub string) (string, error) {
	if sub == "" {
		return "", fs.ErrInvalid
	}
	b, err := s.reserve(ctx, sub)
	if err != nil {
		return "", err
	}
	if b.State == "ready" {
		return b.Workspace, nil
	}
	owner := provider.Owner{Deployment: s.Deployment, Workspace: b.Workspace}
	// Serialize provider allocation across API instances without holding the reservation transaction.
	conn, err := s.DB.Acquire(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Release()
	if _, err = conn.Exec(ctx, "SELECT pg_advisory_lock(hashtextextended($1,3))", b.Workspace); err != nil {
		return "", err
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := conn.Exec(cleanup, "SELECT pg_advisory_unlock(hashtextextended($1,3))", b.Workspace); err != nil {
			conn.Conn().Close(cleanup)
		}
	}()
	var state string
	if err = conn.QueryRow(ctx, "SELECT provisioning_state FROM workspace_machines WHERE workspace_id=$1", b.Workspace).Scan(&state); err != nil {
		return "", err
	}
	if state == "ready" {
		return b.Workspace, nil
	}
	m, err := s.Provider.FindOrAllocate(ctx, owner)
	if err != nil {
		return "", err
	}
	d, err := s.Provider.Open(ctx, owner, m)
	if err != nil {
		return "", err
	}
	defer d.Close()
	tx, err := conn.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	if err = lock(ctx, tx, b.Workspace, false); err != nil {
		return "", err
	}
	if err = d.Reconcile(); err != nil {
		return "", err
	}
	if s.Mode == "sample" {
		if err = seed(ctx, tx, b.Workspace, d); err != nil {
			return "", err
		}
	}
	_, err = tx.Exec(ctx, `UPDATE workspace_machines SET vm_resource_id=$2,workspace_disk_resource_id=$3,provisioning_state='ready',updated_at=now() WHERE workspace_id=$1`, b.Workspace, m.ID, m.DiskID)
	if err != nil {
		return "", err
	}
	return b.Workspace, tx.Commit(ctx)
}

func lock(ctx context.Context, tx pgx.Tx, w string, shared bool) error {
	query := "SELECT pg_advisory_xact_lock(hashtextextended($1,2))"
	if shared {
		query = "SELECT pg_advisory_xact_lock_shared(hashtextextended($1,2))"
	}
	_, err := tx.Exec(ctx, query, w)
	return err
}
func seed(ctx context.Context, tx pgx.Tx, w string, d provider.Filesystem) error {
	v, err := sample.Load()
	if err != nil {
		return err
	}
	fileID, err := publish(ctx, tx, w, d, "Attachments/"+v.Filename, sample.PDF)
	if err != nil {
		return err
	}
	var thread, msg string
	if err = tx.QueryRow(ctx, "INSERT INTO mail_threads(workspace_id,subject) VALUES($1,$2) RETURNING id", w, v.Subject).Scan(&thread); err != nil {
		return err
	}
	if err = tx.QueryRow(ctx, `INSERT INTO mail_messages(workspace_id,thread_id,sender_address,body_text,sent_at) VALUES($1,$2,$3,$4,$5) RETURNING id`, w, thread, v.SenderAddress, v.BodyText, v.SentAt).Scan(&msg); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO mail_attachments VALUES($1,$2,$3)", w, msg, fileID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO calendar_events(workspace_id,title,starts_at,ends_at,time_zone) VALUES($1,$2,$3,$4,$5)", w, v.Title, v.StartsAt, v.EndsAt, v.TimeZone); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "INSERT INTO workspace_preferences(workspace_id,settings) VALUES($1,$2)", w, v.Preferences)
	return err
}

func (s *Service) open(ctx context.Context, w string) (provider.Filesystem, error) {
	var m provider.Machine
	err := s.DB.QueryRow(ctx, "SELECT vm_resource_id,workspace_disk_resource_id FROM workspace_machines WHERE workspace_id=$1 AND provisioning_state='ready'", w).Scan(&m.ID, &m.DiskID)
	if err != nil {
		return nil, err
	}
	owner := provider.Owner{Deployment: s.Deployment, Workspace: w}
	disk, err := s.Provider.Open(ctx, owner, m)
	var replaced provider.Replacement
	if errors.As(err, &replaced) {
		current := replaced.Current
		if current.DiskID != m.DiskID || current.ID == "" {
			return nil, ErrUnavailable
		}
		disk, err = s.Provider.Open(ctx, owner, current)
		if err == nil {
			// Concurrent readers can confirm the same replacement. An older
			// observation must never overwrite a newer persisted binding.
			tag, updateErr := s.DB.Exec(ctx, `UPDATE workspace_machines SET vm_resource_id=$1,updated_at=now()
				WHERE workspace_id=$2 AND workspace_disk_resource_id=$3 AND vm_resource_id IN ($4,$1)
				AND provisioning_state='ready'`, current.ID, w, m.DiskID, m.ID)
			if updateErr != nil || tag.RowsAffected() != 1 {
				disk.Close()
				return nil, ErrUnavailable
			}
		}
	}
	if err != nil {
		return nil, fmt.Errorf("%w: open machine", ErrUnavailable)
	}
	err = pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		if err := lock(ctx, tx, w, false); err != nil {
			return err
		}
		return disk.Reconcile()
	})
	if err != nil {
		disk.Close()
		return nil, err
	}
	return disk, nil
}

func publish(ctx context.Context, tx pgx.Tx, w string, d provider.Filesystem, name string, data []byte) (string, error) {
	key, err := d.Publish(name, data)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(data)
	var id string
	err = tx.QueryRow(ctx, `INSERT INTO files(workspace_id,filename,media_type,byte_size,sha256_digest,storage_key)
 VALUES($1,$2,'application/pdf',$3,$4,$5) ON CONFLICT(workspace_id,storage_key) DO NOTHING RETURNING id`, w, path.Base(name), len(data), digest[:], key).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, "SELECT id FROM files WHERE workspace_id=$1 AND storage_key=$2", w, key).Scan(&id)
	}
	return id, err
}

func (s *Service) Register(ctx context.Context, w, name string) (string, error) {
	if !fs.ValidPath(name) || !strings.EqualFold(path.Ext(name), ".pdf") {
		return "", fs.ErrInvalid
	}
	d, err := s.open(ctx, w)
	if err != nil {
		return "", err
	}
	defer d.Close()
	if id, ok, err := s.attachment(ctx, w, name, d); ok || err != nil {
		return id, err
	}
	data, err := fs.ReadFile(d, name)
	if err != nil {
		return "", err
	}
	var id string
	err = pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		if err := lock(ctx, tx, w, true); err != nil {
			return err
		}
		var err error
		id, err = publish(ctx, tx, w, d, name, data)
		return err
	})
	return id, err
}

func (s *Service) attachment(ctx context.Context, w, name string, d provider.Filesystem) (string, bool, error) {
	dir, err := attachmentDirectory(d)
	if err != nil {
		return "", false, err
	}
	if path.Dir(name) != dir {
		return "", false, nil
	}
	var id string
	err = s.DB.QueryRow(ctx, `SELECT f.id FROM files f JOIN mail_attachments a ON a.workspace_id=f.workspace_id AND a.file_id=f.id WHERE f.workspace_id=$1 AND f.filename=$2`, w, path.Base(name)).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	return id, err == nil, err
}

func (s *Service) Document(ctx context.Context, w, id string) ([]byte, error) {
	var key string
	err := s.DB.QueryRow(ctx, "SELECT storage_key FROM files WHERE workspace_id=$1 AND id=$2", w, id).Scan(&key)
	if err != nil {
		return nil, err
	}
	d, err := s.open(ctx, w)
	if err != nil {
		return nil, err
	}
	defer d.Close()
	return d.ReadVersion(key)
}
