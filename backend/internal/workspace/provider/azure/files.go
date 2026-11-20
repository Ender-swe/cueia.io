package azure

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"net"
	"strings"
	"time"

	"github.com/Ender-swe/cueia.io/backend/internal/workspace/provider"
	"github.com/Ender-swe/cueia.io/backend/internal/workspace/volume"
	"golang.org/x/crypto/ssh"
)

type disk struct{ ssh *ssh.Client }
type file struct {
	*bytes.Reader
	info volume.Entry
}

func (f file) Close() error               { return nil }
func (f file) Stat() (fs.FileInfo, error) { return f.info, nil }

func (d *disk) call(op, name string, data []byte) (volume.Response, error) {
	var res volume.Response
	payload, err := json.Marshal(volume.Request{Operation: op, Name: name, Data: data})
	if err != nil {
		return res, err
	}
	session, err := d.ssh.NewSession()
	if err != nil {
		return res, err
	}
	defer session.Close()
	session.Stdin = bytes.NewReader(payload)
	output, err := session.Output("cueia-files")
	if err != nil {
		return res, errors.New("workspace file service unavailable")
	}
	if err = json.Unmarshal(output, &res); err != nil {
		return res, err
	}
	switch res.Error {
	case "":
		return res, nil
	case "missing":
		return res, fs.ErrNotExist
	case "invalid":
		return res, fs.ErrInvalid
	default:
		return res, errors.New("workspace disk unavailable")
	}
}
func (d *disk) Open(name string) (fs.File, error) {
	r, err := d.call("read", name, nil)
	if err != nil {
		return nil, err
	}
	return file{bytes.NewReader(r.Data), r.Stat}, nil
}
func (d *disk) ReadDir(name string) ([]fs.DirEntry, error) {
	r, err := d.call("list", name, nil)
	if err != nil {
		return nil, err
	}
	out := make([]fs.DirEntry, len(r.Entries))
	for i, e := range r.Entries {
		out[i] = e
	}
	return out, nil
}
func (d *disk) WriteFile(name string, data []byte, _ fs.FileMode) error {
	_, err := d.call("write", name, data)
	return err
}
func (d *disk) ReadVersion(key string) ([]byte, error) {
	r, err := d.call("version", key, nil)
	return r.Data, err
}
func (d *disk) Publish(name string, data []byte) (string, error) {
	r, err := d.call("publish", name, data)
	return r.Key, err
}
func (d *disk) Reconcile() error { _, err := d.call("reconcile", "", nil); return err }
func (d *disk) Close() error     { return d.ssh.Close() }

func (p *Provider) Open(ctx context.Context, o provider.Owner, m provider.Machine) (provider.Filesystem, error) {
	name, err := p.name(o)
	if err != nil {
		return nil, err
	}
	obs, err := p.observe(ctx, name)
	if err != nil {
		return nil, err
	}
	if err := provider.CheckBinding(m, obs.machine); err != nil {
		return nil, err
	}
	nic, err := p.resources.GetByID(ctx, obs.nic, networkVersion, nil)
	if err != nil {
		return nil, err
	}
	var network struct {
		IPConfigurations []struct {
			Properties struct {
				PrivateIPAddress string
				PublicIPAddress  any
			}
		}
	}
	if err = properties(nic.Properties, &network); err != nil {
		return nil, err
	}
	if len(network.IPConfigurations) != 1 {
		return nil, fs.ErrPermission
	}
	cfg := network.IPConfigurations[0].Properties
	ip := net.ParseIP(cfg.PrivateIPAddress)
	if cfg.PublicIPAddress != nil || ip == nil || !ip.IsPrivate() {
		return nil, fs.ErrPermission
	}
	conn, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, "tcp", net.JoinHostPort(ip.String(), "22"))
	if err != nil {
		return nil, err
	}
	checker := ssh.CertChecker{IsHostAuthority: func(k ssh.PublicKey, _ string) bool { return bytes.Equal(k.Marshal(), p.config.HostCA.Marshal()) }}
	config := &ssh.ClientConfig{User: "cueiafiles", Auth: []ssh.AuthMethod{ssh.PublicKeys(p.config.FileClientKey)}, HostKeyCallback: checker.CheckHostKey}
	// The guest host certificate identifies this specific VM incarnation.
	host := strings.ToLower(name + "-" + m.ID)
	conn.SetDeadline(time.Now().Add(15 * time.Second))
	client, chans, requests, err := ssh.NewClientConn(conn, net.JoinHostPort(host, "22"), config)
	if err != nil {
		conn.Close()
		return nil, err
	}
	sc := ssh.NewClient(client, chans, requests)
	return &disk{sc}, nil
}
