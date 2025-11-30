package crocgodyl

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type Limits struct {
	Memory      int64  `json:"memory"`
	Swap        int64  `json:"swap"`
	Disk        int64  `json:"disk"`
	IO          int64  `json:"io"`
	CPU         int64  `json:"cpu"`
	Threads     string `json:"threads"`
	OOMDisabled bool   `json:"oom_disabled"`
}

type FeatureLimits struct {
	Allocations int `json:"allocations"`
	Backups     int `json:"backups"`
	Databases   int `json:"databases"`
}

type ClientServer struct {
	ServerOwner bool   `json:"server_owner"`
	Identifier  string `json:"identifier"`
	UUID        string `json:"uuid"`
	InternalID  int    `json:"internal_id"`
	Name        string `json:"name"`
	Node        string `json:"node"`
	SFTP        struct {
		IP   string `json:"ip"`
		Port int64  `json:"port"`
	} `json:"sftp_details"`
	Description      string        `json:"description"`
	Limits           Limits        `json:"limits"`
	Invocation       string        `json:"invocation"`
	DockerImage      string        `json:"docker_image"`
	EggFeatures      []string      `json:"egg_features"`
	FeatureLimits    FeatureLimits `json:"feature_limits"`
	Status           string        `json:"status"`
	Suspended        bool          `json:"is_suspended"`
	Installing       bool          `json:"is_installing"`
	Transferring     bool          `json:"is_transferring"`
	UnderMaintenance bool          `json:"is_node_under_maintenance"`
}

func (c *Client) GetServers() ([]*ClientServer, error) {
	req := c.newRequest("GET", "", nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Data []struct {
			Attributes *ClientServer `json:"attributes"`
		} `json:"data"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	servers := make([]*ClientServer, 0, len(model.Data))
	for _, s := range model.Data {
		servers = append(servers, s.Attributes)
	}

	return servers, nil
}

func (c *Client) GetServer(identifier string) (*ClientServer, error) {
	req := c.newRequest("GET", "/servers/"+identifier, nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Attributes ClientServer `json:"attributes"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return &model.Attributes, err
}

type WebSocketAuth struct {
	Socket string `json:"socket"`
	Token  string `json:"token"`
}

func (c *Client) GetServerWebSocket(identifier string) (*WebSocketAuth, error) {
	req := c.newRequest("GET", fmt.Sprintf("/servers/%s/websocket", identifier), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Data WebSocketAuth `json:"data"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return &model.Data, nil
}

type ResourceUsage struct {
	MemoryBytes    int64   `json:"memory_bytes"`
	DiskBytes      int64   `json:"disk_bytes"`
	CPUAbsolute    float64 `json:"cpu_absolute"`
	NetworkRxBytes int64   `json:"network_rx_bytes"`
	NetworkTxBytes int64   `json:"network_tx_bytes"`
	Uptime         int64   `json:"uptime"`
}

type Resources struct {
	State     string        `json:"current_state,omitempty"`
	Suspended bool          `json:"is_suspended"`
	Usage     ResourceUsage `json:"resources"`
}

func (c *Client) GetServerResources(identifier string) (*Resources, error) {
	req := c.newRequest("GET", fmt.Sprintf("/servers/%s/resources", identifier), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Attributes Resources `json:"attributes"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return &model.Attributes, nil
}

func (c *Client) SendServerCommand(identifier, command string) error {
	data, _ := json.Marshal(map[string]string{"command": command})
	body := bytes.Buffer{}
	body.Write(data)

	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/command", identifier), &body)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

func (c *Client) SetServerPowerState(identifier, state string) error {
	data, _ := json.Marshal(map[string]string{"signal": state})
	body := bytes.Buffer{}
	body.Write(data)

	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/power", identifier), &body)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

type ClientDatabase struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Host     struct {
		Address string `json:"address"`
		Port    int64  `json:"port"`
	} `json:"host"`
	ConnectionsFrom string `json:"connections_from"`
	MaxConnections  int    `json:"max_connections"`
}

func (c *Client) GetServerDatabases(identifier string) ([]*ClientDatabase, error) {
	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/command", identifier), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Data []struct {
			Attributes *ClientDatabase `json:"attributes"`
		} `json:"data"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	dbs := make([]*ClientDatabase, 0, len(model.Data))
	for _, d := range model.Data {
		dbs = append(dbs, d.Attributes)
	}

	return dbs, nil
}

func (c *Client) CreateDatabase(identifier, remote, database string) (*ClientDatabase, error) {
	data, _ := json.Marshal(map[string]string{"remote": remote, "database": database})
	body := bytes.Buffer{}
	body.Write(data)

	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/databases", identifier), &body)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Attributes ClientDatabase `json:"attributes"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return &model.Attributes, nil
}

func (c *Client) RotateDatabasePassword(identifier, id string) (*ClientDatabase, error) {
	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/databases/%s/rotate-password", identifier, id), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Attributes ClientDatabase `json:"attributes"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return &model.Attributes, nil
}

func (c *Client) DeleteDatabase(identifier, id string) error {
	req := c.newRequest("DELETE", fmt.Sprintf("/servers/%s/databases/%s", identifier, id), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

type File struct {
	Name       string     `json:"name"`
	Mode       string     `json:"mode"`
	ModeBits   string     `json:"mode_bits"`
	Size       int64      `json:"size"`
	IsFile     bool       `json:"is_file"`
	IsSymlink  bool       `json:"is_symlink"`
	MimeType   string     `json:"mimetype"`
	CreatedAt  *time.Time `json:"created_at"`
	ModifiedAt *time.Time `json:"modified_at,omitempty"`
}

func (c *Client) GetServerFiles(identifier, root string) ([]*File, error) {
	req := c.newRequest("GET", fmt.Sprintf("/servers/%s/files/list?directory=%s", identifier, url.PathEscape(root)), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Data []struct {
			Attributes *File `json:"attributes"`
		} `json:"data"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	files := make([]*File, 0, len(model.Data))
	for _, f := range model.Data {
		files = append(files, f.Attributes)
	}

	return files, nil
}

func (c *Client) GetServerFileContents(identifier, file string) ([]byte, error) {
	req := c.newRequest("GET", fmt.Sprintf("/servers/%s/files/contents?file=%s", identifier, url.PathEscape(file)), nil)
	req.Header.Set("Accept", "application/json,text/plain")

	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	return validate(res)
}

type Downloader struct {
	client *Client
	Name   string
	Path   string
	url    string
}

func (d *Downloader) Client() *Client {
	return d.client
}

func (d *Downloader) URL() string {
	return d.url
}

func (d *Downloader) Execute() error {
	info, err := os.Stat(d.Path)
	if err == nil {
		if !info.IsDir() {
			return errors.New("refusing to overwrite existing file path")
		}
	}

	res, err := d.client.Http.Get(d.URL())
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("recieved an unexpected response: %s", res.Status)
	}

	file, err := os.OpenFile(d.Name, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	io.Copy(file, res.Body)
	return nil
}

func (c *Client) DownloadServerFile(identifier, file string) (*Downloader, error) {
	files, err := c.GetServerFiles(identifier, "/")
	if err != nil {
		return nil, err
	}

	_, name := filepath.Split(file)
	for _, f := range files {
		if f.Name == name {
			if f.MimeType == "inode/directory" {
				return nil, errors.New("cannot download a directory")
			}

			break
		}
	}

	req := c.newRequest("GET", fmt.Sprintf("/servers/%s/files/download?file=%s", identifier, url.PathEscape(file)), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Attributes struct {
			URL string `json:"url"`
		} `json:"attributes"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	path, _ := url.PathUnescape(file)
	dl := &Downloader{
		client: c,
		Name:   name,
		Path:   path,
		url:    model.Attributes.URL,
	}

	return dl, nil
}

type RenameDescriptor struct {
	Root  string `json:"root"`
	Files []struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"files"`
}

func (c *Client) RenameServerFiles(identifier string, files RenameDescriptor) error {
	data, _ := json.Marshal(files)
	body := bytes.Buffer{}
	body.Write(data)

	req := c.newRequest("PUT", fmt.Sprintf("/servers/%s/files/rename", identifier), &body)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

func (c *Client) CopyServerFile(identifier, location string) error {
	data, _ := json.Marshal(map[string]string{"location": location})
	body := bytes.Buffer{}
	body.Write(data)

	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/files/copy", identifier), &body)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

func (c *Client) WriteServerFileBytes(identifier, name, header string, content []byte) error {
	body := bytes.Buffer{}
	body.Write(content)

	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/files/write?file=%s", identifier, url.PathEscape(name)), &body)
	req.Header.Set("Content-Type", header)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

func (c *Client) WriteServerFile(identifier, name, content string) error {
	return c.WriteServerFileBytes(identifier, name, "text/plain", []byte(content))
}

type CompressDescriptor struct {
	Root  string   `json:"root"`
	Files []string `json:"files"`
}

func (c *Client) CompressServerFiles(identifier string, files CompressDescriptor) error {
	data, _ := json.Marshal(files)
	body := bytes.Buffer{}
	body.Write(data)

	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/files/compress", identifier), &body)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

type DecompressDescriptor struct {
	Root string `json:"root"`
	File string `json:"file"`
}

func (c *Client) DecompressServerFile(identifier string, file DecompressDescriptor) error {
	data, _ := json.Marshal(file)
	body := bytes.Buffer{}
	body.Write(data)

	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/files/decompress", identifier), &body)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

type DeleteFilesDescriptor struct {
	Root  string   `json:"root"`
	Files []string `json:"files"`
}

func (c *Client) DeleteServerFiles(identifier string, files DeleteFilesDescriptor) error {
	data, _ := json.Marshal(files)
	body := bytes.Buffer{}
	body.Write(data)

	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/files/delete", identifier), &body)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

type CreateFolderDescriptor struct {
	Root string `json:"root"`
	Name string `json:"name"`
}

func (c *Client) CreateServerFileFolder(identifier string, file CreateFolderDescriptor) error {
	data, _ := json.Marshal(file)
	body := bytes.Buffer{}
	body.Write(data)

	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/files/create-folder", identifier), &body)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

type ChmodDescriptor struct {
	Root  string `json:"root"`
	Files []struct {
		File string `json:"file"`
		Mode uint32 `json:"mode"`
	} `json:"files"`
}

func (c *Client) ChmodServerFiles(identifier string, files ChmodDescriptor) error {
	data, _ := json.Marshal(files)
	body := bytes.Buffer{}
	body.Write(data)

	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/files/chmod", identifier), &body)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

type PullDescriptor struct {
	URL        string `json:"url"`
	Directory  string `json:"directory,omitempty"`
	Filename   string `json:"filename,omitempty"`
	UseHeader  bool   `json:"use_header,omitempty"`
	Foreground bool   `json:"foreground,omitempty"`
}

func (c *Client) PullServerFile(identifier string, file PullDescriptor) error {
	data, _ := json.Marshal(file)
	body := bytes.Buffer{}
	body.Write(data)

	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/files/pull", identifier), &body)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

type Uploader struct {
	client *Client
	url    string
	Path   string
}

func (u *Uploader) Client() *Client {
	return u.client
}

func (u *Uploader) URL() string {
	return u.url
}

func (u *Uploader) Execute() error {
	if u.Path == "" {
		return errors.New("no file path has been specified")
	}

	info, err := os.Stat(u.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("file path does not exist")
		}

		return err
	}

	if info.IsDir() {
		return errors.New("path must go to a file not a directory")
	}

	file, err := os.Open(u.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	body := bytes.Buffer{}
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("files", info.Name())
	io.Copy(part, file)
	writer.Close()

	res, err := u.client.Http.Post(u.URL(), writer.FormDataContentType(), &body)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("recieved an unexpected response: %s", res.Status)
	}

	return nil
}

func (c *Client) GetUploadUrl(identifier string) (string, error) {
	req := c.newRequest("GET", fmt.Sprintf("/servers/%s/files/upload", identifier), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return "", err
	}

	buf, err := validate(res)
	if err != nil {
		return "", err
	}

	var model struct {
		Attributes struct {
			URL string `json:"url"`
		} `json:"attributes"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return "", err
	}

	return model.Attributes.URL, nil
}

func (c *Client) UploadServerFile(identifier string) (*Uploader, error) {
	uploadUrl, err := c.GetUploadUrl(identifier)
	if err != nil {
		return nil, err
	}

	up := &Uploader{client: c, url: uploadUrl}
	return up, nil
}

type AllocationAttributes struct {
	ID      int64  `json:"id"`
	IP      string `json:"ip"`
	IPAlias string `json:"ip_alias"`
	Port    int64  `json:"port"`
	Notes   string `json:"notes"`
	Default bool   `json:"is_default"`
}

func (c *Client) GetAllocations(identifier string) ([]*AllocationAttributes, error) {
	req := c.newRequest("GET", fmt.Sprintf("/servers/%s", identifier), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Attributes struct {
			Relationships struct {
				Allocations struct {
					Data []struct {
						Allocations *AllocationAttributes `json:"attributes"`
					} `json:"data"`
				} `json:"allocations"`
			} `json:"relationships"`
		} `json:"attributes"`
	}

	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	allocations := make([]*AllocationAttributes, 0, len(model.Attributes.Relationships.Allocations.Data))
	for _, s := range model.Attributes.Relationships.Allocations.Data {
		allocations = append(allocations, s.Allocations)
	}

	return allocations, nil
}

func (c *Client) CreateAllocation(identifier string) (*AllocationAttributes, error) {
	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/network/allocations", identifier), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Allocations *AllocationAttributes `json:"attributes"`
	}

	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return model.Allocations, nil
}

func (c *Client) ChangeNotes(identifier string, allocationID int64, notes string) (*AllocationAttributes, error) {
	data, _ := json.Marshal(map[string]string{"notes": notes})
	body := bytes.Buffer{}
	body.Write(data)

	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/network/allocations/%d", identifier, allocationID), &body)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Allocations *AllocationAttributes `json:"attributes"`
	}

	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return model.Allocations, nil
}

func (c *Client) MakePrimary(identifier string, allocationID int64) (*AllocationAttributes, error) {
	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/network/allocations/%d/primary", identifier, allocationID), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Allocations *AllocationAttributes `json:"attributes"`
	}

	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return model.Allocations, nil
}

func (c *Client) DeleteAllocation(identifier string, allocationID int64) error {
	req := c.newRequest("DELETE", fmt.Sprintf("/servers/%s/network/allocations/%d", identifier, allocationID), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

type Meta struct {
	StartupCommand    string            `json:"startup_command"`
	DockerImages      map[string]string `json:"docker_images"`
	RawStartupCommand string            `json:"raw_startup_command"`
}

func (c *Client) GetStartupInfo(identifier string) (*Meta, error) {
	req := c.newRequest("GET", fmt.Sprintf("/servers/%s/startup", identifier), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Meta Meta `json:"meta"`
	}

	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return &model.Meta, nil
}

func (c *Client) UpdateDockerImage(identifier string, dockerImage string) error {
	data, _ := json.Marshal(map[string]string{"docker_image": dockerImage})
	body := bytes.Buffer{}
	body.Write(data)

	req := c.newRequest("PUT", fmt.Sprintf("/servers/%s/settings/docker-image", identifier), &body)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

type EggVariables struct {
	Object     string `json:"object"`
	Attributes struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		EnvVariable  string `json:"env_variable"`
		DefaultValue string `json:"default_value"`
		ServerValue  string `json:"server_value"`
		IsEditable   bool   `json:"is_editable"`
		Rules        string `json:"rules"`
	} `json:"attributes"`
}

func (c *Client) GetVariables(identifier string) ([]*EggVariables, error) {
	req := c.newRequest("GET", fmt.Sprintf("/servers/%s/startup", identifier), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Data []*EggVariables `json:"data"`
	}

	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return model.Data, nil
}

func (c *Client) PutVariable(identifier, key, value string) error {
	data, _ := json.Marshal(map[string]string{"key": key, "value": value})
	body := bytes.Buffer{}
	body.Write(data)

	req := c.newRequest("PUT", fmt.Sprintf("/servers/%s/startup/variable", identifier), &body)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

func (c *Client) Reinstall(identifier string) error {
	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/settings/reinstall", identifier), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

type SchedulesInfo struct {
	Id             int         `json:"id"`
	Name           string      `json:"name"`
	Cron           Cron        `json:"cron"`
	IsActive       bool        `json:"is_active"`
	IsProcessing   bool        `json:"is_processing"`
	OnlyWhenOnline bool        `json:"only_when_online"`
	LastRunAt      interface{} `json:"last_run_at"`
	NextRunAt      time.Time   `json:"next_run_at"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

type Cron struct {
	DayOfWeek  string `json:"day_of_week"`
	DayOfMonth string `json:"day_of_month"`
	Hour       string `json:"hour"`
	Minute     string `json:"minute"`
	Month      string `json:"month"`
}

type Attributes struct {
	Schedule SchedulesInfo
}

type Data struct {
	Attributes Attributes `json:"attributes"`
}

func (c *Client) GetSchedules(identifier string) (*[]Data, error) {
	req := c.newRequest("GET", fmt.Sprintf("/servers/%s/schedules", identifier), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Data []Data `json:"data"`
	}

	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return &model.Data, nil
}

type Schedule struct {
	DayOfMonth     string `json:"day_of_month"`
	DayOfWeek      string `json:"day_of_week"`
	Hour           string `json:"hour"`
	IsActive       bool   `json:"is_active"`
	Minute         string `json:"minute"`
	Month          string `json:"month"`
	Name           string `json:"name"`
	OnlyWhenOnline bool   `json:"only_when_online"`
}

func (c *Client) GetSchedule(identifier string, scheduleID int64) (*Schedule, error) {
	req := c.newRequest("GET", fmt.Sprintf("/servers/%s/schedules/%d", identifier, scheduleID), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Schedule Schedule
	}

	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return &model.Schedule, nil
}

func (c *Client) CreateSchedules(identifier string, newSchedule Schedule) error {
	data, _ := json.Marshal(newSchedule)
	body := bytes.Buffer{}
	body.Write(data)

	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/schedules", identifier), &body)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

func (c *Client) UpdateSchedule(identifier string, updatedSchedule Schedule, scheduleID int64) error {
	data, _ := json.Marshal(updatedSchedule)
	body := bytes.Buffer{}
	body.Write(data)

	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/schedules/%d", identifier, scheduleID), &body)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

func (c *Client) ExecuteSchedule(identifier string, scheduleID int64) error {
	req := c.newRequest("GET", fmt.Sprintf("/servers/%s/schedules/%d/execute", identifier, scheduleID), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

func (c *Client) DeleteSchedule(identifier string, scheduleID int64) error {
	req := c.newRequest("DELETE", fmt.Sprintf("/servers/%s/schedules/%d", identifier, scheduleID), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

type TasksInfo struct {
	Action     string `json:"action"`
	Payload    string `json:"payload"`
	TimeOffset string `json:"time_offset"`
}

func (c *Client) GetScheduleTasks(identifier string, scheduleID int64) ([]*TasksInfo, error) {
	req := c.newRequest("GET", fmt.Sprintf("/servers/%s/schedules/%d/tasks", identifier, scheduleID), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Schedules []*TasksInfo `json:"tasks"`
	}

	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return model.Schedules, nil
}

type Task struct {
	Action            string `json:"action"`
	ContinueOnFailure bool   `json:"continue_on_failure"`
	Payload           string `json:"payload"`
	TimeOffset        string `json:"time_offset"`
}

func (c *Client) CreateScheduleTasks(identifier string, scheduleID int64, task Task) error {
	data, _ := json.Marshal(task)
	body := bytes.Buffer{}
	body.Write(data)

	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/schedules/%d/tasks", identifier, scheduleID), &body)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

func (c *Client) UpdateScheduleTasks(identifier string, scheduleID int64, taskID int64, task Task) error {
	data, _ := json.Marshal(task)
	body := bytes.Buffer{}
	body.Write(data)

	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/schedules/%d/tasks/%d", identifier, scheduleID, taskID), &body)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

func (c *Client) DeleteScheduleTasks(identifier string, scheduleID int64, taskID int64) error {
	req := c.newRequest("DELETE", fmt.Sprintf("/servers/%s/schedules/%d/tasks/%d", identifier, scheduleID, taskID), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

type BackupInfo struct {
	Uuid         string    `json:"uuid"`
	Name         string    `json:"name"`
	IgnoredFiles []string  `json:"ignored_files"`
	Sha256Hash   string    `json:"sha256_hash"`
	Bytes        int       `json:"bytes"`
	CreatedAt    time.Time `json:"created_at"`
	CompletedAt  time.Time `json:"completed_at"`
	IsSuccessful bool      `json:"is_successful"`
	IsLocked     bool      `json:"is_locked"`
}

func (c *Client) GetBackups(identifier string) ([]*BackupInfo, error) {
	req := c.newRequest("GET", fmt.Sprintf("/servers/%s/backups", identifier), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Data []*BackupInfo `json:"data"`
	}

	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return model.Data, nil
}

func (c *Client) CreateBackups(identifier string, name string, ignored string, isLocked bool) error {
	backupData := map[string]interface{}{
		"name":      name,
		"ignored":   ignored,
		"is_locked": isLocked,
	}
	data, _ := json.Marshal(backupData)
	body := bytes.Buffer{}
	body.Write(data)

	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/backups", identifier), &body)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

func (c *Client) GetBackup(identifier string, backupID string) (*BackupInfo, error) {
	req := c.newRequest("GET", fmt.Sprintf("/servers/%s/backups/%s", identifier, backupID), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Data *BackupInfo `json:"data"`
	}

	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return model.Data, nil
}

type DownloadBackupURL struct {
	URL string `json:"url"`
}

func (c *Client) DownloadBackup(identifier string, backupID string) (*DownloadBackupURL, error) {
	req := c.newRequest("GET", fmt.Sprintf("/servers/%s/backups/%s/download", identifier, backupID), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Attributes DownloadBackupURL `json:"attributes"`
	}

	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return &model.Attributes, nil
}

func (c *Client) LockBackup(identifier string, backupID string) error {
	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/backups/%s/lock	", identifier, backupID), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

func (c *Client) RestoreBackup(identifier string, backupID string, truncate bool) error {
	backupData := map[string]bool{"truncate": truncate}
	data, _ := json.Marshal(backupData)
	body := bytes.Buffer{}
	body.Write(data)

	req := c.newRequest("POST", fmt.Sprintf("/servers/%s/backups/%s/restore", identifier, backupID), &body)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

func (c *Client) DeleteBackup(identifier string, backupID string) error {
	req := c.newRequest("DELETE", fmt.Sprintf("/servers/%s/backups/%s", identifier, backupID), nil)
	res, err := c.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}
